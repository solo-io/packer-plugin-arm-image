package builder

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/chroot"
	packer_common_common "github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packer_common_commonsteps "github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/mitchellh/mapstructure"
	"github.com/solo-io/packer-plugin-arm-image/pkg/builder/embed"
	"github.com/solo-io/packer-plugin-arm-image/pkg/image"
	"github.com/solo-io/packer-plugin-arm-image/pkg/image/utils"

	getter "github.com/hashicorp/go-getter/v2"
)

const BuilderId = "yuval-k.arm-image"

var (
	knownTypes = map[utils.KnownImageType][]string{
		utils.RaspberryPi: {"/boot", "/"},
		utils.BeagleBone:  {"/"},
		utils.Kali:        {"/root", "/"},
		utils.Ubuntu:      {"/boot/firmware", "/"},
	}
	knownArgs = map[utils.KnownImageType][]string{
		utils.BeagleBone: {"-cpu", "cortex-a8"},
	}

	defaultBase = [][]string{
		{"proc", "proc", "/proc"},
		{"sysfs", "sysfs", "/sys"},
		{"bind", "/dev", "/dev"},
		{"devpts", "devpts", "/dev/pts"},
		{"binfmt_misc", "binfmt_misc", "/proc/sys/fs/binfmt_misc"},
	}
	resolvConfBindMount = []string{"bind", "/etc/resolv.conf", "/etc/resolv.conf"}

	defaultChrootTypes = map[utils.KnownImageType][][]string{
		utils.Unknown: defaultBase,
	}
)

type ResolvConfBehavior string

const (
	Off      ResolvConfBehavior = "off"
	CopyHost ResolvConfBehavior = "copy-host"
	BindHost ResolvConfBehavior = "bind-host"
	Delete   ResolvConfBehavior = "delete"
)

const ChrootKey = "mount_path"

var generatedDataKeys = map[string]string{
	ChrootKey: "MountPath",
}

type Builder struct {
	config Config
	runner *multistep.BasicRunner
}

func NewBuilder() *Builder {
	return &Builder{}
}

func (b *Builder) autoDetectType() utils.KnownImageType {
	if len(b.config.ISOUrls) < 1 {
		return ""
	}
	url := b.config.ISOUrls[0]
	return utils.GuessImageType(url)
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec {
	return b.config.FlatMapstructure().HCL2Spec()
}

func (b *Builder) Prepare(cfgs ...interface{}) ([]string, []string, error) {
	var md mapstructure.Metadata
	err := config.Decode(&b.config, &config.DecodeOpts{
		Metadata:           &md,
		PluginType:         BuilderId,
		Interpolate:        true,
		InterpolateContext: &b.config.ctx,
		InterpolateFilter:  &interpolate.RenderFilter{},
	}, cfgs...)
	if err != nil {
		return nil, nil, err
	}
	var errs *packer.MultiError
	var warnings []string
	isoWarnings, isoErrs := b.config.ISOConfig.Prepare(&b.config.ctx)
	warnings = append(warnings, isoWarnings...)
	errs = packer.MultiErrorAppend(errs, isoErrs...)

	if b.config.OutputFile == "" {
		if b.config.OutputDir != "" {
			warnings = append(warnings, "output_directory is deprecated, use output_filename instead.")
			b.config.OutputFile = filepath.Join(b.config.OutputDir, "image")
		} else {
			b.config.OutputFile = fmt.Sprintf("output-%s/image", b.config.PackerConfig.PackerBuildName)
		}
	}

	if b.config.LastPartitionExtraSize > 0 {
		warnings = append(warnings, "last_partition_extra_size is deprecated, use target_image_size to grow your image")
	}

	if b.config.ChrootMounts == nil {
		b.config.ChrootMounts = make([][]string, 0)
	}

	if len(b.config.ChrootMounts) == 0 {
		b.config.ChrootMounts = defaultChrootTypes[utils.Unknown]
		if imageDefaults, ok := defaultChrootTypes[b.config.ImageType]; ok {
			b.config.ChrootMounts = imageDefaults
		}
	}

	if len(b.config.AdditionalChrootMounts) > 0 {
		b.config.ChrootMounts = append(b.config.ChrootMounts, b.config.AdditionalChrootMounts...)
	}

	if b.config.ResolvConf == BindHost {
		b.config.ChrootMounts = append(b.config.ChrootMounts, resolvConfBindMount)
	}

	if b.config.CommandWrapper == "" {
		b.config.CommandWrapper = "{{.Command}}"
	}

	if b.config.ImageType == "" {
		// defaults...
		b.config.ImageType = b.autoDetectType()
	} else {
		if _, ok := knownTypes[b.config.ImageType]; !ok {

			var validvalues []utils.KnownImageType
			for k := range knownTypes {
				validvalues = append(validvalues, k)
			}
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("unknown image_type. must be one of: %v", validvalues))
			b.config.ImageType = ""
		}
	}
	if b.config.ImageType != "" {
		if len(b.config.ImageMounts) == 0 {
			b.config.ImageMounts = knownTypes[b.config.ImageType]
		}
		if len(b.config.QemuArgs) == 0 {
			b.config.QemuArgs = knownArgs[b.config.ImageType]
		}
	}

	if len(b.config.ImageMounts) == 0 {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("no image mounts provided. Please set the image mounts or image type."))
	}

	if b.config.QemuBinary == "" {
		b.config.QemuBinary = "qemu-arm-static"
	}
	// convert to full path
	path, err := exec.LookPath(b.config.QemuBinary)
	if err != nil {
		// not found in path, check if if we have it embedded
		if b.config.DisableEmbedded {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("qemu binary not found."))
		} else {
			// try to fetch an embedded version
			embeddedQ, err := embed.GetEmbededQemu(b.config.QemuBinary)
			if err != nil {
				errs = packer.MultiErrorAppend(errs, fmt.Errorf("embedded qemu is not available - %w", err))
			} else {
				defer embeddedQ.Close()
				qemupathincache, err := packer.CachePath(b.config.QemuBinary)
				if err != nil {
					errs = packer.MultiErrorAppend(errs, fmt.Errorf("cannot cache qemu - %w", err))
				} else if _, err := os.Stat(qemupathincache); os.IsNotExist(err) {
					// copy to cache folder, make executable, and use as path.
					// also check if it exists before copying.
					cachedFile, err := os.OpenFile(qemupathincache, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
					if err != nil {
						errs = packer.MultiErrorAppend(errs, fmt.Errorf("cannot cache - %w", err))
					} else {
						defer cachedFile.Close()
						io.Copy(cachedFile, embeddedQ)
						b.config.QemuBinary = qemupathincache
					}
				} else if err == nil {
					b.config.QemuBinary = qemupathincache
				} else {
					errs = packer.MultiErrorAppend(errs, fmt.Errorf("unknown cache error - %w", err))
				}
			}
		}
	} else {
		// found it in the path, set the config to it!
		if !strings.Contains(path, "qemu-") {
			warnings = append(warnings, "binary doesn't look like qemu-user")
		}
		b.config.QemuBinary = path
	}

	log.Println("qemu path", b.config.QemuBinary)
	if errs != nil && len(errs.Errors) > 0 {
		return nil, warnings, errs
	}

	generatedData := make([]string, 0, len(generatedDataKeys))
	for _, v := range generatedDataKeys {
		generatedData = append(generatedData, v)
	}

	return generatedData, warnings, nil
}

type wrappedCommandTemplate struct {
	Command string
}

func init() {
	// HACK: go-getter automatically decompress, which hurts caching.
	// additionally, we use native binaries to decompress which is faster anyway.
	// disable decompressors:
	getter.Decompressors = map[string]getter.Decompressor{}
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	ui.Say(fmt.Sprintf("Image type: %s", b.config.ImageType))

	wrappedCommand := func(command string) (string, error) {
		b.config.ctx.Data = &wrappedCommandTemplate{Command: command}
		return interpolate.Render(b.config.CommandWrapper, &b.config.ctx)
	}

	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("debug", b.config.PackerDebug)
	state.Put("hook", hook)
	state.Put("ui", ui)
	state.Put("wrappedCommand", packer_common_common.CommandWrapper(wrappedCommand))

	steps := []multistep.Step{
		&packer_common_commonsteps.StepDownload{
			Checksum:    b.config.ISOChecksum,
			Description: "Image",
			ResultKey:   "iso_path",
			Url:         b.config.ISOUrls,
			Extension:   b.config.TargetExtension,
			TargetPath:  b.config.TargetPath,
		},
		&stepCopyImage{FromKey: "iso_path", ResultKey: "imagefile", ImageOpener: image.NewImageOpener(ui)},
	}

	if b.config.LastPartitionExtraSize > 0 || b.config.TargetImageSize > 0 {
		steps = append(steps,
			&stepResizeLastPart{FromKey: "imagefile"},
		)
	}

	steps = append(steps,
		&stepMapImage{ImageKey: "imagefile", ResultKey: "partitions"},
	)
	if b.config.LastPartitionExtraSize > 0 || b.config.TargetImageSize > 0 {
		steps = append(steps,
			&stepResizeFs{PartitionsKey: "partitions"},
		)
	}

	steps = append(steps,
		&stepMountImage{
			PartitionsKey:    "partitions",
			ResultKey:        ChrootKey,
			MountPath:        b.config.MountPath,
			GeneratedDataKey: generatedDataKeys[ChrootKey],
		},
		&chroot.StepMountExtra{
			ChrootMounts: b.config.ChrootMounts,
		},
		&StepMountCleanup{},
	)

	if b.config.ResolvConf == CopyHost || b.config.ResolvConf == Delete {
		steps = append(steps,
			&stepHandleResolvConf{ChrootKey: ChrootKey, Delete: b.config.ResolvConf == Delete})
	}

	native := runtime.GOARCH == "arm" || runtime.GOARCH == "arm64"
	if !native {
		steps = append(steps,
			&stepQemuUserStatic{ChrootKey: ChrootKey, PathToQemuInChrootKey: "qemuInChroot", Args: Args{Args: b.config.QemuArgs}},
			&stepRegisterBinFmt{QemuPathKey: "qemuInChroot"},
		)
	}

	steps = append(steps,
		&chroot.StepChrootProvision{},
	)

	b.runner = &multistep.BasicRunner{Steps: steps}

	// Executes the steps
	b.runner.Run(ctx, state)

	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}
	// check if it is ok
	_, canceled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if canceled || halted {
		return nil, errors.New("step canceled or halted")
	}

	return &Artifact{
		image:     state.Get("imagefile").(string),
		StateData: map[string]interface{}{"generated_data": state.Get("generated_data")},
	}, nil
}

type Artifact struct {
	image     string
	StateData map[string]interface{}
}

func (a *Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Files() []string {
	return []string{a.image}
}

func (a *Artifact) Id() string {
	return ""
}

func (a *Artifact) String() string {
	return a.image
}

func (a *Artifact) State(name string) interface{} {
	return a.StateData[name]
}

func (a *Artifact) Destroy() error {
	return os.Remove(a.image)
}
