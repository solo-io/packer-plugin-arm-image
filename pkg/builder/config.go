//go:generate go run github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc struct-markdown
//go:generate go run github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc mapstructure-to-hcl2 -type Config

package builder

import (
	packer_common_common "github.com/hashicorp/packer-plugin-sdk/common"
	packer_common_commonsteps "github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/solo-io/packer-plugin-arm-image/pkg/image/utils"
)

type Config struct {
	packer_common_common.PackerConfig `mapstructure:",squash"`
	// While arm image are not ISOs, we resuse the ISO logic as it basically has no ISO specific code.
	// Provide the arm image in the iso_url fields.
	packer_common_commonsteps.ISOConfig `mapstructure:",squash"`

	// Lets you prefix all builder commands, such as with ssh for a remote build host. Defaults to "".
	// Copied from other builders :)
	CommandWrapper string `mapstructure:"command_wrapper"`

	// Output directory, where the final image will be stored.
	// Deprecated - Use OutputFile instead
	OutputDir string `mapstructure:"output_directory"`

	// Output filename, where the final image will be stored
	OutputFile string `mapstructure:"output_filename"`

	// Image type. this is used to deduce other settings like image mounts and qemu args.
	// If not provided, we will try to deduce it from the image url. (see autoDetectType())
	// For list of valid values, see: pkg/image/utils/images.go
	ImageType utils.KnownImageType `mapstructure:"image_type"`

	// Where to mounts the image partitions in the chroot.
	// first entry is the mount point of the first partition. etc..
	ImageMounts []string `mapstructure:"image_mounts"`

	// The path where the volume will be mounted. This is where the chroot environment will be.
	// Will be a temporary directory if left unspecified.
	MountPath string `mapstructure:"mount_path"`

	// What directories mount from the host to the chroot.
	// leave it empty for reasonable defaults.
	// array of triplets: [type, device, mntpoint].
	ChrootMounts [][]string `mapstructure:"chroot_mounts"`

	// What directories mount from the host to the chroot, in addition to the default ones.
	// Use this instead of `chroot_mounts` if you want to add to the existing defaults instead of
	// overriding them
	// array of triplets: [type, device, mntpoint].
	// for example: `["bind", "/run/systemd", "/run/systemd"]`
	AdditionalChrootMounts [][]string `mapstructure:"additional_chroot_mounts"`

	// Can be one of: off, copy-host, bind-host, delete. Defaults to off
	ResolvConf ResolvConfBehavior `mapstructure:"resolv-conf"`

	// Should the last partition be extended? this only works for the last partition in the
	// dos partition table, and ext filesystem
	LastPartitionExtraSize uint64 `mapstructure:"last_partition_extra_size"`
	// The target size of the final image. The last partition will be extended to
	// fill up this much room. I.e. if the generated image is 256MB and TargetImageSize
	// is set to 384MB the last partition will be extended with an additional 128MB.
	TargetImageSize uint64 `mapstructure:"target_image_size"`

	// Qemu binary to use. default is qemu-arm-static
	// If this is an absolute path, it will be used. Otherwise, we will look for one in your PATH
	// and finally, try to auto fetch one from https://github.com/multiarch/qemu-user-static/
	QemuBinary string `mapstructure:"qemu_binary"`
	// Do not use embedded qemu.
	DisableEmbedded bool `mapstructure:"disable_embedded"`
	// Arguments to qemu binary. default depends on the image type. see init() function above.
	QemuArgs []string `mapstructure:"qemu_args"`

	ctx interpolate.Context
}
