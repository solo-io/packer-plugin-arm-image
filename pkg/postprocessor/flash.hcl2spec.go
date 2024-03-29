// Code generated by "packer-sdc mapstructure-to-hcl2"; DO NOT EDIT.

package postprocessor

import (
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/zclconf/go-cty/cty"
)

// FlatFlashConfig is an auto-generated flat version of FlashConfig.
// Where the contents of a field with a `mapstructure:,squash` tag are bubbled up.
type FlatFlashConfig struct {
	Device         *string `mapstructure:"device" cty:"device" hcl:"device"`
	NotInteractive *bool   `mapstructure:"not_interactive" cty:"not_interactive" hcl:"not_interactive"`
	Verify         *bool   `mapstructure:"verify" cty:"verify" hcl:"verify"`
}

// FlatMapstructure returns a new FlatFlashConfig.
// FlatFlashConfig is an auto-generated flat version of FlashConfig.
// Where the contents a fields with a `mapstructure:,squash` tag are bubbled up.
func (*FlashConfig) FlatMapstructure() interface{ HCL2Spec() map[string]hcldec.Spec } {
	return new(FlatFlashConfig)
}

// HCL2Spec returns the hcl spec of a FlashConfig.
// This spec is used by HCL to read the fields of FlashConfig.
// The decoded values from this spec will then be applied to a FlatFlashConfig.
func (*FlatFlashConfig) HCL2Spec() map[string]hcldec.Spec {
	s := map[string]hcldec.Spec{
		"device":          &hcldec.AttrSpec{Name: "device", Type: cty.String, Required: false},
		"not_interactive": &hcldec.AttrSpec{Name: "not_interactive", Type: cty.Bool, Required: false},
		"verify":          &hcldec.AttrSpec{Name: "verify", Type: cty.Bool, Required: false},
	}
	return s
}
