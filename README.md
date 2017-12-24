# Packer plugin for ARM images.

## Overview

This plugin lets you take an existing ARM image, and modify it on your x86 machine.
It is optimized for raspberry pi use case - MBR partition table, with the file system partition 
being the last partition.

With this plugin, you can:
- Provision new ARM images from existing ones.
- Use ARM binaries for provisioning ('apt-get install' for example)
- Resize the last partition (the filesystem partition in the raspberry pi) in case you need more
  space than the default.

Tested for raspbian images on built on Ubuntu 17.10. It is based partly on the chroot aws 
provisioner, though the code was copied to prevent aws dependenceies.

# How it works?

The plugin runs the provisioners in a chroot envrionment. binary execution is done using
qemu-arm-static, via binfmt_misc.


## Dependencies:
This builder uses the following shell commands:
- kpartx - mapping the partitons to mountable devices
- qemu-user-static - Executing arm binaries

To install the needed binaries:
```
sudo apt install kpartx qemu-user-static
```
Other commands that are used are (that should already be installed) : mount, umount, cp, ls, chroot.

To resize the filesystem, the following commands are used:
- e2fsck
- resize2fs

Note: resizing is only supported for the last active
partition in an MBR partition table (as there is no need to move things).

# Configuration
To use, you need to provide an existing image that we will then modify. We re-use packer's support 
for downloading ISOs (though the image should not be an ISO file).
Supporting also zipped images (enabling you downloading official raspbian images directly).

See [example.json](example.json) and [builder.go](pkg/builder/builder.go) for details.

# Compiling and Testing
## Building
As this is an alpha release - consider using a vm to run this code for isolation.

This project uses go dep tool for dependencies.
To build:
```bash
mkdir -p $GOPATH/src/github.com/solo-io/
cd $GOPATH/src/github.com/solo-io/
git clone https://github.com/solo-io/packer-builder-arm-image
cd packer-builder-arm-image
dep ensure
go build
```

## Test VM
Testing in a VM (vs directly on your laptop) is highly recommended, though not manadatory. To test in a VM:

Download a linux machine. I personally used Ubuntu 17.10, but other versions \ distributions should work as well (as long as the shell commands use are similary. namely kpartx who's output is parsed)

Then I used virt-install to help install it:
```bash
virt-install -n devel -r 4096 --disk path=$PWD/devel.img,bus=virtio,size=40 -c ~/Downloads/ubuntu-17.10-desktop-amd64.iso --network network=default,model=virtio
```

## Testing
Once it is installed, you can build the plugin locally and copy it to the machine:
```bash
cd $GOPATH/src/github.com/solo-io/packer-builder-arm-image
go build && scp packer-builder-arm-image example.json USER@IP-OF-VM:
```
(note the colon after the ip of the vm)

Then, (install packer)[https://www.packer.io/docs/install/index.html] in the vm, the ssh to the vm and test:
```
ssh USER@IP-OF-VM
sudo packer build -debug example.json
```

The example config produces an image with go installed and extends the filesystem by 1GB.

That's it! Flash it and run!
