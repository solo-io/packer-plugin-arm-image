//go:generate go run github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc struct-markdown
//go:generate go run github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc mapstructure-to-hcl2 -type FlashConfig

package postprocessor

import (
	"context"
	"errors"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/solo-io/packer-plugin-arm-image/pkg/flasher"
)

type FlashConfig struct {
	Device         string `mapstructure:"device"`
	NotInteractive bool   `mapstructure:"not_interactive"`
	Verify         bool   `mapstructure:"verify"`
}

type Flasher struct {
	config FlashConfig
}

func NewFlasher() packer.PostProcessor {
	return &Flasher{}
}

func (f *Flasher) ConfigSpec() hcldec.ObjectSpec {
	return f.config.FlatMapstructure().HCL2Spec()
}

func (f *Flasher) Configure(cfgs ...interface{}) error {
	err := config.Decode(&f.config, &config.DecodeOpts{
		Interpolate:       true,
		InterpolateFilter: &interpolate.RenderFilter{},
	}, cfgs...)
	if err != nil {
		return err
	}
	return nil
}

func (f *Flasher) PostProcess(ctx context.Context, ui packer.Ui, ain packer.Artifact) (a packer.Artifact, keep bool, forceOverride bool, err error) {
	inputfiles := ain.Files()
	if len(inputfiles) != 1 {
		return nil, false, false, errors.New("ambiguous images")
	}
	imageToFlash := inputfiles[0]

	flashercfg := flasher.FlashConfig{
		Image:          imageToFlash,
		Device:         f.config.Device,
		NotInteractive: f.config.NotInteractive,
		Verify:         f.config.Verify,
	}
	flasher := flasher.NewFlasher(ui, flashercfg)
	return nil, false, false, flasher.Flash(ctx)

}
