
build {
  sources = [
    "source.arm-image.hostapd"
  ]

  provisioner "shell" {
    inline = [
      "apt-get update",
      "apt-get install -y hostapd dnsmasq vlan bridge-utils"
    ]
  }
  provisioner "file" {
    source = "overlay"
    destination = "/tmp/"
  }
  provisioner "shell" {
    inline = [
      "cp -r /tmp/overlay/* /",
      "rm -rf /tmp/overlay"
    ]
  }

  provisioner "shell" {
    inline = [
      "apt-get update",
      "apt-get --autoremove purge -y openresolv ifupdown dhcpcd5 isc-dhcp-client isc-dhcp-common"
    ]
  }

  provisioner "shell" {
    inline = [
      "sed -i s/NameOfNetwork/${var.wifi_ssid}/ /etc/hostapd/hostapd.conf",
      "sed -i s/PasswordOfNetwork/${var.wifi_psk}/ /etc/hostapd/hostapd.conf"
    ]
  }

  provisioner "shell" {
    inline = [
      "systemctl enable systemd-networkd",
      "ln -sf /run/systemd/resolve/resolv.conf /etc/resolv.conf",
      "systemctl enable systemd-resolved",
      "systemctl enable systemd-timesyncd"
    ]
  }
  provisioner "shell" {
    inline = [
      "mkdir ${var.image_home_dir}/.ssh"
    ]
  }
  provisioner "shell" {
    inline = ["touch /boot/ssh"]
  }
  provisioner "file" {
    source = "${local.ssh_key}"
    destination = "${var.image_home_dir}/.ssh/authorized_keys"
  }
  provisioner "shell" {
    inline = [
        "sed '/PasswordAuthentication/d' -i /etc/ssh/sshd_config",
        "echo >> /etc/ssh/sshd_config",
        "echo 'PasswordAuthentication no' >> /etc/ssh/sshd_config",
      ]
  }
}
