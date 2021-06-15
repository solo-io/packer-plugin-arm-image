
build {
  sources = [
    "source.arm-image.hostapd"
  ]
  # install hostapd, bridge utils, and other utilities.
  provisioner "shell" {
    inline = [
      "apt-get update",
      "apt-get install -y hostapd dnsmasq vlan bridge-utils"
    ]
  }
  # upload all our configuration files.
  provisioner "file" {
    source = "overlay"
    destination = "/tmp/"
  }
  # the configuration folder is structured so that copying it to / will
  # place the files in the correct location. so do just that.
  # among other things, this will setup the bridge between the wifi and ethernet interfaces.
  provisioner "shell" {
    inline = [
      "cp -r /tmp/overlay/* /",
      "rm -rf /tmp/overlay"
    ]
  }

  # configure hostapd with the user parameters
  provisioner "shell" {
    inline = [
      "sed -i s/NameOfNetwork/${var.wifi_ssid}/ /etc/hostapd/hostapd.conf",
      "sed -i s/PasswordOfNetwork/${var.wifi_psk}/ /etc/hostapd/hostapd.conf"
    ]
  }

  # we will use systemd networking, so remove these packages.
  provisioner "shell" {
    inline = [
      "apt-get update",
      "apt-get --autoremove purge -y openresolv ifupdown dhcpcd5 isc-dhcp-client isc-dhcp-common"
    ]
  }

  # enable systemd networking
  provisioner "shell" {
    inline = [
      "systemctl enable systemd-networkd",
      "ln -sf /run/systemd/resolve/resolv.conf /etc/resolv.conf",
      "systemctl enable systemd-resolved",
      "systemctl enable systemd-timesyncd",
    ]
  }

  # setup secure ssh access.
  provisioner "shell" {
    inline = [
      "mkdir ${var.image_home_dir}/.ssh",
    ]
  }
  # enable ssh in the pi.
  # if you don't know or don't care about ssh, delete these steps.
  provisioner "shell" {
    inline = ["touch /boot/ssh"]
  }
  # upload our public key as an authorized key.
  provisioner "file" {
    source = "${local.ssh_key}"
    destination = "${var.image_home_dir}/.ssh/authorized_keys"
  }
  # set permissions for the authorized keys;
  # disabled password authentication
  provisioner "shell" {
    inline = [
        "chown pi:pi ${var.image_home_dir}/.ssh/authorized_keys",
        "sed '/PasswordAuthentication/d' -i /etc/ssh/sshd_config",
        "echo >> /etc/ssh/sshd_config",
        "echo 'PasswordAuthentication no' >> /etc/ssh/sshd_config",
      ]
  }
  # end of ssh setup
}
