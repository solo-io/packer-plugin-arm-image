#!/bin/bash
#
# @script          provision.sh
# @description     provisioning script that builds environment for
#                  https://github.com/solo-io/packer-plugin-arm-image
#
#                 By default, sets up environment, builds the plugin, and image
##
set -x
# Set to false to disable auto building

# Now ready to build the plugin
mkdir -p $GOPATH/src/github.com/solo-io/
cd $GOPATH/src/github.com/solo-io/

# clean up potential residual files from previous builds
rm -rf packer-plugin-arm-image
if [[ -z "${GIT_CLONE_URL}" ]]; then
  cp -a /vagrant packer-plugin-arm-image
else
  git clone ${GIT_CLONE_URL} packer-plugin-arm-image
fi
cd packer-plugin-arm-image
go build

# Check if plugin built and copy into place
if [[ ! -f packer-plugin-arm-image ]]; then
  echo "Error Plugin failed to build."
  exit
else
  cp packer-plugin-arm-image /vagrant
fi
