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

# Update the system
sudo apt-get update -qq

sudo DEBIAN_FRONTEND=noninteractive apt-get \
  -y \
  --allow-downgrades \
  --allow-remove-essential \
  --allow-change-held-packages \
 -qq \
 -o Dpkg::Options::="--force-confdef" \
 -o Dpkg::Options::="--force-confold" \
  dist-upgrade

# Provides the add-apt-repository script
sudo apt-get install -y software-properties-common

# Install required packages
sudo apt-get install -y \
    kpartx \
    qemu-user-static \
    git \
    wget \
    curl \
    vim \
    unzip \
    gcc

# Download specific Go version
echo "Removing existing Go packages and installing Go"
[[ -e /tmp/go ]] && rm -rf /tmp/go*
sudo apt remove -y \
  'golang-*'
cd /tmp
wget https://dl.google.com/go/go1.14.3.linux-amd64.tar.gz
tar xf go1.14.3.linux-amd64.tar.gz
sudo cp -r go /usr/lib/go-1.14
rm -rf /tmp/go*

# Set GO paths for vagrant user
echo 'export GOROOT=/usr/lib/go-1.14
export GOPATH=$HOME/work
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin' | tee -a /home/vagrant/.profile

# Also set them while we work:
export GOROOT=/usr/lib/go-1.14
export GOPATH=$HOME/work
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin

# Download and install packer
[[ -e /tmp/packer ]] && rm -rf /tmp/packer*
wget https://releases.hashicorp.com/packer/1.5.2/packer_1.5.2_linux_amd64.zip \
    -q -O /tmp/packer_1.5.2_linux_amd64.zip
cd /tmp
unzip -u packer_1.5.2_linux_amd64.zip
sudo cp packer /usr/local/bin
rm -rf /tmp/packer*
cd ..
