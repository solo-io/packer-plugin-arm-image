# packer-sdc

The packer software development command is meant for plugin maintainers and
Packer maintainers, it helps generate docs and code. For Packer and for Packer
Plugins.

At the top of many Packer go files, you can see the following:

```go
//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type Config,CustomerEncryptionKey
```
This will generate multiple files.

See docs for subcommands for further info by reading their readme or using their
specific help commands:

* `packer-sdc mapstructure-to-hcl2 -h`
* `packer-sdc struct-markdown -h`
* `packer-sdc renderdocs -h`
