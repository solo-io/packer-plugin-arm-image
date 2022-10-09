packer {
  required_plugins {
    arm-image = {
      version = ">= 0.2.5"
      source  = "github.com/solo-io/arm-image"
    }
  }
}

source "arm-image" "ubuntu_2004_arm64" {
  image_type      = "raspberrypi"
  iso_url         = "https://cdimage.ubuntu.com/releases/22.04.1/release/ubuntu-22.04.1-preinstalled-server-arm64+raspi.img.xz"
  iso_checksum    = "sha256:5d0661eef1a0b89358159f3849c8f291be2305e5fe85b7a16811719e6e8ad5d1"
  output_filename = "images/rpi-ubuntu2204-arm64.img"
  qemu_binary     = "qemu-aarch64-static"
  image_mounts    = ["/boot/firmware","/"]
  target_image_size =  3969908736
  chroot_mounts = [
        ["proc", "proc", "/proc"],
        ["sysfs", "sysfs", "/sys"],
        ["bind", "/dev", "/dev"],
        ["devpts", "devpts", "/dev/pts"],
        ["binfmt_misc", "binfmt_misc", "/proc/sys/fs/binfmt_misc"],
        ["bind", "/run/systemd", "/run/systemd"]
  ]
}

build {
  name = "base_image"
  sources = ["source.arm-image.ubuntu_2004_arm64"]
  provisioner "shell" {
    inline = [
      "touch /boot/ssh",
      "touch /boot/firmware/meta-data",
      "touch /boot/firmware/vendor-data"
    ]
  }

# Uncomment this block with your updated network-config file
#   provisioner "file" {
#     source = "cloud-init/network-config"
#     destination = "/boot/firmware/network-config"
#   }

  provisioner "file" {
    source = "cloud-init/user-data"
    destination = "/boot/firmware/user-data"
  }

  provisioner "shell" {
    scripts = [
      "./scripts/install-cowsay.sh"
    ]
  }
}