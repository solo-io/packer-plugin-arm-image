package main

import (
	"github.com/hashicorp/packer-plugin-sdk/plugin"
	"github.com/solo-io/packer-builder-arm-image/pkg/builder"
	"github.com/solo-io/packer-builder-arm-image/pkg/postprocessor"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterBuilder(builder.NewBuilder())
	server.RegisterPostProcessor(postprocessor.NewFlasher())
	server.Serve()
}
