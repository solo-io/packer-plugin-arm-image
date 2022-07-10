# Packer plugin for ARM images

This plugin lets you take an existing ARM image and modify it on your x86 machine.
It is optimized for raspberry pi use case - MBR partition table, with the file system partition
being the last partition.

With this plugin, you can:

- Provision new ARM images from existing ones
- Use ARM binaries for provisioning (`apt-get install` for example)
- Resize the last partition (the filesystem partition in the raspberry pi) in case you need more
  space than the default.

Tested for Raspbian images built on Ubuntu 19.10. It is based partly on the chroot AWS
provisioner, though the code was copied to prevent AWS dependencies.

## How it works

The plugin runs the provisioners in a chroot environment.  Binary execution is done using
`qemu-arm-static`, via `binfmt_misc`.

### Dependencies

This builder uses the following shell commands:

- `qemu-user-static` - Executing arm binaries. This is optional as the released binary can use embedded versions of `qemu-aarch64-static` and `qemu-arm-static`. If you have one installed, it will be used instead of the embedded ones.
- `losetup` - To mount the image. This command is pre-installed in most distributions.

To install the needed binaries on derivatives of the Debian Linux variant:

```shell
sudo apt install qemu-user-static
```

Fedora:

```shell
sudo dnf install qemu-user-static
```

Archlinux:

```shell
pacman -S qemu-arm-static
```

Other commands that are used are (that should already be installed) : mount, umount, cp, ls, chroot.

To resize the filesystem, the following commands are used:

- `e2fsck`
- `resize2fs`

To provide custom arguments to `qemu-arm-static` using the `qemu_args` config, `gcc` is required (to compile a C wrapper).

Note: resizing is only supported for the last active
partition in an MBR partition table (as there is no need to move things).

This plugin uses the following kernel feature:

- support for `/proc/sys/fs/binfmt_misc` so that ARM binaries are automatically executed with qemu

### Operation

This provisioner allows you to run packer provisioners on your ARM image locally. It does so by mounting the image on to the local file system, and then using `chroot` combined with `binfmt_misc` to the provisioners in a simulated ARM environment.

## Configuration

To use, you need to provide an existing image that we will then modify. We re-use packer's support
for downloading ISOs (though the image should not be an ISO file).
Supporting also zipped images (enabling you downloading official raspbian images directly).

See [raspbian_golang.json](samples/raspbian_golang.json) and [config.go](pkg/builder/config.go) for details.
For configuration reference, see the [builder doc](docs/builders/arm-image.mdx).

*Note* if your image is arm64, set `qemu_binary` to `qemu-aarch64-static` in your configuration json file.

## Compiling and Testing

### Building

As this tool performs low-level OS manipulations - consider using a VM to run this code for isolation. While this is recommended, it is not mandatory.

This project uses [go modules](https://github.com/golang/go/wiki/Modules) for dependencies introduced in Go 1.11.
To build:

```bash
git clone https://github.com/solo-io/packer-plugin-arm-image
cd packer-plugin-arm-image
go mod download
go build
```

### Running with Vagrant

This project includes a Vagrant file and helper script that build a VM run time environment. The run time environment has
custom provisions to build an image in an iterative fashion (thanks to @tommie-lie for adding this feature).

To use the Vagrant environment, run the following commands:

```shell
git clone https://github.com/solo-io/packer-plugin-arm-image
cd packer-plugin-arm-image
vagrant up
```

To build an image edit [samples/raspbian_golang.json](samples/raspbian_golang.json) (or set `PACKERFILE` to point to your json config), and use `vagrant provision` like so:

```shell
vagrant provision --provision-with build-image
```

The example config produces an image with go installed and extends the filesystem by 1GB.

That's it! Flash it and run!

### Running locally

This builder requires root permissions as it performs low level machine operations. To run it locally,
you can set `PACKER_CONFIG_DIR` back to your local home before sudo-ing to packer. For example:

```shell
PACKER_CONFIG_DIR=$HOME sudo -E $(which packer) build .
```

### Running with Docker

#### Prerequisites

Docker needs capability of creating new devices on host machine, so it can create `/dev/loop*` and mount image into it. While it may be possible to accomplish with multiple `--device-cgroup-rule` and `--add-cap`, it's much easier to use `--privileged` flag to accomplish that. Even so, it is considered bad practice to do so, do it with extra precautions. Also because of those requirements rootless will not work for this container.

#### Option 1: Clone this repo and build the Docker image locally

Build the Docker image locally

```shell
docker build -t packer-builder-arm .
```

Build the `samples/raspbian_golang.json` Packer image

```shell
docker run \
  --rm \
  --privileged \
  -v /dev:/dev \
  -v ${PWD}:/build:ro \
  -v ${PWD}/packer_cache:/build/packer_cache \
  -v ${PWD}/output-arm-image:/build/output-arm-image \
  -e PACKER_CACHE_DIR=/build/packer_cache \
  packer-builder-arm build samples/raspbian_golang.json
```

#### Option 2: Run the published Docker image

Alternatively, you can use the `ghcr.io/solo-io/packer-plugin-arm-image` that's built off latest master without needing to clone this repository.

```shell
docker run \
  --rm \
  --privileged \
  -v /dev:/dev \
  -v ${PWD}:/build:ro \
  -v ${PWD}/packer_cache:/build/packer_cache \
  -v ${PWD}/output-arm-image:/build/output-arm-image \
  ghcr.io/solo-io/packer-plugin-arm-image build samples/raspbian_golang.json
```

That's it, flash it and run!

### Running Standalone

```shell
packer build samples/raspbian_golang.json
```

## Flashing

We have a post-processor stage for flashing.

### Golang flasher

```shell
go build cmd/flasher/main.go
```

It will auto-detect most things and guides you with questions.

### dd

(Tested on MacOS)

```shell
# find the identifier of the device you want to flash
diskutil list

# un-mount the disk
diskutil unmountDisk /dev/disk2

# flash the image, go for a coffee
sudo dd bs=4m if=output-arm-image/image of=/dev/disk2

# eject the disk
diskutil eject /dev/disk2
```

## Cookbook

### Raspberry Pi Provisioners

#### Enable ssh

```json
{
  "type": "shell",
  "inline": ["touch /boot/ssh"]
}
```

#### Set WiFi password

set the user variables name `wifi_name` and `wifi_password`, then:

```json
{
  "type": "shell",
  "inline": [
    "echo 'network={' >> /etc/wpa_supplicant/wpa_supplicant.conf",
    "echo '    ssid=\"{{user `wifi_name`}}\"' >> /etc/wpa_supplicant/wpa_supplicant.conf",
    "echo '    psk=\"{{user `wifi_password`}}\"' >> /etc/wpa_supplicant/wpa_supplicant.conf",
    "echo '}' >> /etc/wpa_supplicant/wpa_supplicant.conf"
    ]
}
```

#### Add ssh key to authorized keys, enable ssh, disable password login

This example locks down the image to only use your
current ssh key. Disabling password login makes it extra secure for networked environments.

Note: this example requires you to run the plugin without a VM, as it copies your local ssh key.

```json
{
  "variables": {
    "ssh_key_src": "{{env `HOME`}}/.ssh/id_rsa.pub",
    "image_home_dir": "/home/pi"
  },
  "builders": [
    {
      "type": "arm-image",
      "iso_url": "https://downloads.raspberrypi.org/raspbian_lite/images/raspbian_lite-2017-12-01/2017-11-29-raspbian-stretch-lite.zip",
      "iso_checksum": "sha256:e942b70072f2e83c446b9de6f202eb8f9692c06e7d92c343361340cc016e0c9f"
    }
  ],
  "provisioners": [
    {
      "type": "shell",
      "inline": [
        "mkdir {{user `image_home_dir`}}/.ssh"
      ]
    },
    {
      "type": "file",
      "source": "{{user `ssh_key_src`}}",
      "destination": "{{user `image_home_dir`}}/.ssh/authorized_keys"
    },
    {
      "type": "shell",
      "inline": [
        "touch /boot/ssh"
      ]
    },
    {
      "type": "shell",
      "inline": [
        "sed '/PasswordAuthentication/d' -i /etc/ssh/sshd_config",
        "echo >> /etc/ssh/sshd_config",
        "echo 'PasswordAuthentication no' >> /etc/ssh/sshd_config"
      ]
    }
  ]
}
```

### A complete example

See everything included in here: [samples/pi-secure-wifi-ssh.json](samples/pi-secure-wifi-ssh.json). Build like so:

```shell
sudo packer build  -var wifi_name=SSID -var wifi_password=PASSWORD samples/pi-secure-wifi-ssh.json
# or  if running from vagrant ssh:
sudo packer build  -var wifi_name=SSID -var wifi_password=PASSWORD /vagrant/samples/pi-secure-wifi-ssh.json
```
