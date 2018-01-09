#!/bin/bash
#
# @script          provision.sh
# @description     provisioning script that builds environment for
#                  https://github.com/solo-io/packer-builder-arm-image
#
#                 By default, sets up environment, builds the plugin, and image
##

# Set to false to disable auto building
export RUNBUILDS = true

# Update the system
sudo apt-get update -qq

sudo DEBIAN_FRONTEND=noninteractive apt-get -y --force-yes -qq -o Dpkg::Options::="--force-confdef" -o Dpkg::Options::="--force-confold" dist-upgrade
sudo apt-get install -y software-properties-common

# Add the golang repo
sudo add-apt-repository --yes ppa:gophers/archive

# Install required packages
sudo apt-get update
sudo apt-get install -y \
    kpartx \
    qemu-user-static \
    git \
    wget \
    curl \
    vim \
    unzip \
    golang-1.9-go

# Set GO paths for vagrant user
echo "export GOROOT=/usr/lib/go-1.9
export GOPATH=$HOME/work
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin" | tee -a /home/vagrant/.profile

# Also set them while we work:
export GOROOT=/usr/lib/go-1.9
export GOPATH=$HOME/work
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin

# Install go dep
go get -u github.com/golang/dep/cmd/dep

# Download and install packer
wget https://releases.hashicorp.com/packer/1.1.3/packer_1.1.3_linux_amd64.zip \
    -O /tmp/packer_1.1.3_linux_amd64.zip
pushd /tmp
unzip packer_1.1.3_linux_amd64.zip
sudo cp packer /usr/local/bin
popd

# If RUNBUILDS is true, build the plugin, then the image
if [[ RUNBUILDS = true ]]; then{

    # Now ready to build the plugin
    mkdir -p $GOPATH/src/github.com/solo-io/
    pushd $GOPATH/src/github.com/solo-io/
    git clone https://github.com/solo-io/packer-builder-arm-image
    pushd ./packer-builder-arm-image
    dep ensure
    go build

    # Check if plugin built and copy into place
    if [[ ! -f packer-builder-arm-image ]]; then {
        echo "Error Plugin failed to build."
        exit
    } else {
        mkdir -p /home/vagrant/.packer.d/plugins
        cp packer-builder-arm-image /home/vagrant/.packer.d/plugins/
        cp example.json /home/vagrant
        popd; popd
    }; fi

    # Now build the image
    if [[ ! -f /home/vagrant/.packer.d/plugins/packer-builder-arm-image ]]; then {
        echo "Error: Plugin not found. Retry build."
        exit
    } else {
        echo "Attempting to build image"
        pushd /home/vagrant

        # If there is a custom json, try that one
        # otherwise go with the default
        if [[ -f /vagrant/example.json ]]; then {
            sudo packer build /vagrant/example.json
        } else {
            sudo packer build /home/vagrant/example.json
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
}; fi
