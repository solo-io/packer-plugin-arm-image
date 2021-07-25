## `renderdocs`

`renderdocs` is meant to be used in Packer plugins. It renders the docs by
replacing any `@include 'partial.mdx'` call with its actual content, for
example:

`packer-sdc renderdocs -src ./docs-src -dst ./docs-rendered -partials ./docs-partials`

Will first copy the contents of the `./docs-src` dir into the `./docs-rendered`
dir (any file in `./docs-rendered` that is not present it `./docs-src` will be
removed), then each file in `./docs-rendered` will be parsed and any 
`@include 'partial.mdx'` call will be replaced by the contents of the 
`partial.mdx` file.
