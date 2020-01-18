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

# Add the golang repo
sudo add-apt-repository --yes ppa:longsleep/golang-backports
sudo apt-get update

# Install required packages
sudo apt-get install -y \
    kpartx \
    qemu-user-static \
    git \
    wget \
    curl \
    vim \
    unzip \
    golang-go \
    gcc

# Set GO paths for vagrant user
echo 'export GOROOT=/usr/lib/go-1.13
export GOPATH=$HOME/work
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin' | tee -a /home/vagrant/.profile

# Also set them while we work:
export GOROOT=/usr/lib/go-1.13
export GOPATH=$HOME/work
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin

# Download and install packer
[[ -e /tmp/packer ]] && rm /tmp/packer
wget https://releases.hashicorp.com/packer/1.4.5/packer_1.4.5_linux_amd64.zip \
    -q -O /tmp/packer_1.4.5_linux_amd64.zip
cd /tmp
unzip -u packer_1.4.5_linux_amd64.zip
sudo cp packer /usr/local/bin
cd ..
