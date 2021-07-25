## `mapstructure-to-hcl`

`mapstructure-to-hcl` Helps generate the code necessary for any plugin,
including the core plugins to communicate the HCL2 layout of a plugin with
Packer, for Example, with the following `config.go` file:


```
//go:generate packer-sdc mapstructure-to-hcl2 -type Config

// Config helps configure things
type Config struct {
```

Will generate a `config.hcl2spec.go` file containing a `FlatConfig` struct,
which is a struct with all the mapstructure nested fields 'flattened' into the
`FlatConfig`, so nothing is nested. The `FlatConfig` struct will get a
`HCL2Spec` function that describes its HCL2 layout. This will be used to read
and validate actual HCL2 files. The `config.hcl2spec.go` will also add a
`FlatMapstructure` function to the `Config` struct. That function returns a
`FlatConfig`. These functions together define an interface meant for a plugin
component to 'speak' the HCL2 language with the Packer core.

Before HCL2, Packer JSON heavily relied on the mapstructure decoding library to
load/parse user config files, making this part of the code very tested. To go to
HCL2 this command was created. 

Here are a few differences/gaps betweens HCL2 and mapstructure:
 * in HCL2 all basic struct fields (string/int/struct) that are not pointers
  are required ( must be set ). In mapstructure everything is optional.
 * mapstructure allows to 'squash' fields
 (ex: Field CommonStructType `mapstructure:",squash"`) this allows to
 decorate structs and reuse configuration code. HCL2 parsing libs don't have
 anything similar.
