#!/bin/bash
#
# @script          provision.sh
# @description     provisioning script that builds environment for
#                  https://github.com/solo-io/packer-plugin-arm-image
#
#                 By default, sets up environment, builds the plugin, and image
##
set -x
set -e
# Set to false to disable auto building
export PACKERFILE=${PACKERFILE:-samples/raspbian_golang.json}

PLUGIN_DIR=${PLUGIN_DIR:-/root/.packer.d/plugins}
sudo mkdir -p $PLUGIN_DIR
sudo cp /vagrant/packer-plugin-arm-image "$PLUGIN_DIR/"
# Now build the image
if sudo test ! -f "$PLUGIN_DIR/packer-plugin-arm-image"; then {
    echo "Error: Plugin not found. Retry build."
    exit
} else {
    echo "Attempting to build image"

    PACKER_LOG=$(mktemp)
    export PACKER_CONFIG_DIR=/root/

    # If there is a custom json, try that one
    # otherwise go with the default
    if [[ -f /vagrant/${PACKERFILE} ]]; then {
        sudo -E packer build /vagrant/${PACKERFILE} | tee ${PACKER_LOG}
    } else {
        if [[ -f $GOPATH/src/github.com/solo-io/packer-plugin-arm-image/${PACKERFILE} ]]; then {
            sudo -E packer build $GOPATH/src/github.com/solo-io/packer-plugin-arm-image/${PACKERFILE} | tee ${PACKER_LOG}
        } else {
            echo "Error: packer build definition ${PACKERFILE} not found."
            exit
        }; fi
    }; fi

    BUILD_NAME=$(grep -Po "(?<=Build ').*(?=' finished.)" ${PACKER_LOG})
    IMAGE_PATH=$(grep -Po "(?<=--> ${BUILD_NAME}: ).*" ${PACKER_LOG})
    rm -f ${PACKER_LOG}

    # If the new image is there, copy it out or throw an error
    if [[ -f ${HOME}/${IMAGE_PATH} ]]; then {
        sudo cp ${HOME}/${IMAGE_PATH} \
            /vagrant/${IMAGE_PATH%/image}.img
    } else {
        echo "Error: Unable to find build artifact."
        exit
    }; fi
}; fi
