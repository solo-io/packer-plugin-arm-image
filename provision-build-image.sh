#!/bin/bash
#
# @script          provision.sh
# @description     provisioning script that builds environment for
#                  https://github.com/solo-io/packer-builder-arm-image
#
#                 By default, sets up environment, builds the plugin, and image
##
set -x
set -e
# Set to false to disable auto building
export PACKERFILE=${PACKERFILE:-example.json}

mkdir -p /home/vagrant/.packer.d/plugins
cp /vagrant/packer-builder-arm-image /home/vagrant/.packer.d/plugins/

# Now build the image
if [[ ! -f /home/vagrant/.packer.d/plugins/packer-builder-arm-image ]]; then {
    echo "Error: Plugin not found. Retry build."
    exit
} else {
    echo "Attempting to build image"

    # If there is a custom json, try that one
    # otherwise go with the default
    if [[ -f /vagrant/${PACKERFILE} ]]; then {
        sudo packer build /vagrant/${PACKERFILE}
    } else {
        if [[ -f $GOPATH/src/github.com/solo-io/packer-builder-arm-image/${PACKERFILE} ]]; then {
            sudo packer build $GOPATH/src/github.com/solo-io/packer-builder-arm-image/${PACKERFILE}
        } else {
            echo "Error: packer build definition ${PACKERFILE} not found."
            exit
        }; fi
    }; fi

    # If the new image is there, copy it out or throw an error
    if [[ -f /home/vagrant/output-arm-image/image ]]; then {
        sudo cp /home/vagrant/output-arm-image/image \
            /vagrant/raspbian-stretch-modified.img
    } else {
        echo "Error: Unable to find build artifact."
        exit
    }; fi
}; fi

