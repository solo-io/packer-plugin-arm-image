package main

import (
	"github.com/hashicorp/packer-plugin-sdk/plugin"
	"github.com/solo-io/packer-plugin-arm-image/pkg/builder"
	"github.com/solo-io/packer-plugin-arm-image/pkg/postprocessor"
	"github.com/solo-io/packer-plugin-arm-image/version"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterBuilder(plugin.DEFAULT_NAME, builder.NewBuilder())
	pps.RegisterPostProcessor(plugin.DEFAULT_NAME, postprocessor.NewFlasher())
	pps.SetVersion(version.PluginVersion)
	err := pps.Run()
	if err != nil {
		panic(err)
	}
}
