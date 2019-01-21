#!/bin/bash -ex

if [ -z "$TAG_NAME" ]; then
    exit 0
fi

./hack/github-release.sh owner=solo-io repo=packer-builder-arm-image tag=$TAG_NAME
./hack/upload-github-release-asset.sh owner=solo-io repo=packer-builder-arm-image tag=$TAG_NAME filename=./packer-builder-arm-image
./hack/upload-github-release-asset.sh owner=solo-io repo=packer-builder-arm-image tag=$TAG_NAME filename=./cmd/flasher/flasher