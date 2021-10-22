


check-generated:
	./tools/check_generated.sh

release-snapshot:
	go run github.com/goreleaser/goreleaser release --snapshot --rm-dist

release:
	go run github.com/goreleaser/goreleaser release

ci-release-docs:
	go run github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc renderdocs -src docs -partials docs-partials/ -dst docs/
	/bin/sh -c "[ -d docs ] && zip -r docs.zip docs/"