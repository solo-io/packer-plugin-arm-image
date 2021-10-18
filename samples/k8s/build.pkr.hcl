
build {
  sources = [
    "source.arm-image.k8s"
  ]
  # install k8s and docker keys
  # install utilities.
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

  provisioner "shell" {
    inline = [
      "lsb_release -cs", # informational only
      "curl -fsSLo /usr/share/keyrings/kubernetes-archive-keyring.gpg https://packages.cloud.google.com/apt/doc/apt-key.gpg",
      "curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg",
      "apt-get update",
      "apt-get full-upgrade -y",
      "apt-get install -y bridge-utils"
    ]
  }

  # install stuff for kubeadm and then insall kubeadm
  provisioner "shell" {
    inline = [
      "apt-get install -y apt-transport-https ca-certificates curl",
      "apt-get install -y kubelet kubeadm kubectl",
      "apt-mark hold kubelet kubeadm kubectl",
    ]
  }

  # install docker
  provisioner "shell" {
    inline = [
      "apt-get install -y docker-ce docker-ce-cli containerd.io",
      "systemctl enable docker",
    ]
  }

  provisioner "shell" {
    inline = [
      # remove swap, as k8s will not work with swap on... (see: https://raspberrypi.stackexchange.com/questions/84390/how-to-permanently-disable-swap-on-raspbian-stretch-lite)
      "update-rc.d dphys-swapfile remove",
      "apt purge -y dphys-swapfile",
      # adjust cgroups (see: https://github.com/kubernetes/kubernetes/issues/67310):
      "sed -i '1s/$/ cgroup_enable=cpuset cgroup_enable=memory/' /boot/cmdline.txt",
    ]
  }

  # these fail in packer; we can figure out a way to enable these in a later time.
  # you should run them once the pi boots.
  # # run preflight checks
  # provisioner "shell" {
  #   inline = [
  #     "kubeadm init phase preflight",
  #   ]
  # }

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
