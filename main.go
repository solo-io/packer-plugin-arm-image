package main

import (
	"github.com/hashicorp/packer/packer/plugin"
	"github.com/solo-io/packer-builder-arm-image/pkg/builder"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterBuilder(builder.NewBuilder())
	server.Serve()
}
