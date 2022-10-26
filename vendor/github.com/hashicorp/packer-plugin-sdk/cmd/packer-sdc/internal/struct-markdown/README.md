## `struct-markdown`

`struct-markdown` will read all go structs that have fields with a mapstructure
tag, for Example:


```
//go:generate packer-sdc struct-markdown

// Config helps configure things
type Config struct {

	// The JSON file containing your account credentials. Not required if you
	// run Packer on a HappyCloud instance with a service account. Instructions for
	// creating the file or using service accounts are above.
	AccountFile string `mapstructure:"account_file" required:"false"`

	// Foo is an example field.
	Foo string `mapstructure:"foo" required:"true"`
```

This will generate a `Config.mdx` file containing the header docs of the Config
struct, a `Config-not-required.mdx` file containing the docs for the the
`account_file` field and a `Config-required.mdx` file containing the docs for
the `foo` field. This is quite helpful in the sense that the code now becomes
the single source of truth for docs. In Packer, many common structs are reused
in internal and external plugins, this binary makes it possible to use these the
docs too. See the documentation for `renderdocs` for further info.
