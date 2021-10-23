


check-generated:
	./tools/check_generated.sh

GORELEASER=go run github.com/goreleaser/goreleaser

build:
	API_VERSION="$(shell go run . describe 2>/dev/null | jq -r .api_version)" \
		$(GORELEASER) build --skip-validate --snapshot --rm-dist

release-snapshot:
	$(MAKE) check-generated
	API_VERSION="$(shell go run . describe 2>/dev/null | jq -r .api_version)" \
		$(GORELEASER) release --snapshot --rm-dist --skip-publish

release:
	$(MAKE) check-generated
	API_VERSION="$(shell go run . describe 2>/dev/null | jq -r .api_version)" \
		$(GORELEASER) release

ci-release-docs:
	go run github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc renderdocs -src docs -partials docs-partials/ -dst docs/
	/bin/sh -c "[ -d docs ] && zip -r docs.zip docs/"

install-local:
	go build -o packer-plugin-arm-image .
	mkdir -p $(HOME)/.packer.d/plugins
	cp packer-plugin-arm-image $(HOME)/.packer.d/plugins/

packer:
	go install github.com/hashicorp/packer@v1.7.7

testacc: install-local
	@PACKER_ACC=1 go test -count $(COUNT) -v $(TEST) -timeout=120m