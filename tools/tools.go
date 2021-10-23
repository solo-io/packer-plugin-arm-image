// +build tools

/*
	Explanation for tools pattern:
	https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
*/

package tools

// import stuff to make go mod version them
// this package should not be used.

import (
	_ "github.com/goreleaser/goreleaser"
	_ "github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc"
)
