package main

import (
	_ "embed"
	"log"
	"os"

	mapstructure_to_hcl2 "github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc/internal/mapstructure-to-hcl2"
	"github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc/internal/plugincheck"
	"github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc/internal/renderdocs"
	struct_markdown "github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc/internal/struct-markdown"
	"github.com/hashicorp/packer-plugin-sdk/version"
	"github.com/mitchellh/cli"
)

var (
	app = "packer-sdc"

	//go:embed README.md
	readme string
)

func main() {
	c := cli.NewCLI(app, version.SDKVersion.String())

	c.Args = os.Args[1:]
	c.HelpFunc = func(m map[string]cli.CommandFactory) string {
		str := cli.BasicHelpFunc(app)(m)
		return str + "\n" + readme
	}
	c.Commands = map[string]cli.CommandFactory{
		"struct-markdown": func() (cli.Command, error) {
			return &struct_markdown.Command{}, nil
		},
		"mapstructure-to-hcl2": func() (cli.Command, error) {
			return &mapstructure_to_hcl2.Command{}, nil
		},
		"renderdocs": func() (cli.Command, error) {
			return &renderdocs.Command{}, nil
		},
		"plugin-check": func() (cli.Command, error) {
			return &plugincheck.Command{}, nil
		},
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}
