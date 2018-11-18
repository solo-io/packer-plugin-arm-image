#!/bin/bash
#
# @script          provision.sh
# @description     provisioning script that builds environment for
#                  https://github.com/solo-io/packer-builder-arm-image
#
#                 By default, sets up environment, builds the plugin, and image
##
set -x
# Set to false to disable auto building

# Now ready to build the plugin
mkdir -p $GOPATH/src/github.com/solo-io/
pushd $GOPATH/src/github.com/solo-io/
# clean up potential residual files from previous builds
rm -rf packer-builder-arm-image
if [[ -z "${GIT_CLONE_URL}" ]]; then {
    cp -a /vagrant packer-builder-arm-image
} else {
    git clone ${GIT_CLONE_URL} packer-builder-arm-image
}; fi
pushd ./packer-builder-arm-image
go build

# Check if plugin built and copy into place
if [[ ! -f packer-builder-arm-image ]]; then {
    echo "Error Plugin failed to build."
    exit
} else {
    cp packer-builder-arm-image /vagrant
}; fi

