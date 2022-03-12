NAME=arm-image
BINARY=packer-plugin-${NAME}
COUNT?=1
TEST?=$(shell go list ./...)

check-generated:
	./tools/check_generated.sh

GORELEASER=go run github.com/goreleaser/goreleaser

build:
	go generate ./...
	go build -o ${BINARY} .

test:
	go test -race -count $(COUNT) $(TEST) -timeout=3m

ci-release-docs:
	rm -rf ./docs
	go run github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc renderdocs -src docs-src -partials docs-partials/ -dst docs/
	/bin/sh -c "[ -d docs ] && zip -r docs.zip docs/"

install-local: build
	go build -o ${BINARY} .
	mkdir -p $(HOME)/.packer.d/plugins
	mv ${BINARY} $(HOME)/.packer.d/plugins/

packer:
	which packer || go install github.com/hashicorp/packer@v1.8.0

testacc:
	PACKER_ACC=1 go test -count $(COUNT) -v $(TEST) -timeout=120m

testacc-sudo:
	cd pkg/builder && \
	go test -c . && \
	PACKER_ACC=1 PACKER_CONFIG_DIR=$(HOME) sudo -E bash -c "PATH=$(HOME)/go/bin:$$PATH ./builder.test" && \
	rm img.delete builder.test
 
plugin-check: build
	go run github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc plugin-check ${BINARY}

release-snapshot:
	$(MAKE) check-generated
	API_VERSION="$(shell go run . describe 2>/dev/null | jq -r .api_version)" \
		$(GORELEASER) release --snapshot --rm-dist --skip-publish

release:
	$(MAKE) check-generated
	API_VERSION="$(shell go run . describe 2>/dev/null | jq -r .api_version)" \
		$(GORELEASER) release
