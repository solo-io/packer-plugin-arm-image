source "arm-image" "raspberry_pi_os" {
  iso_checksum = "008d7377b8c8b853a6663448a3f7688ba98e2805949127a1d9e8859ff96ee1a9"
  iso_url      = "https://downloads.raspberrypi.org/raspios_lite_armhf/images/raspios_lite_armhf-2021-11-08/2021-10-30-raspios-bullseye-armhf-lite.zip"
}

build {
  sources = ["source.arm-image.raspberry_pi_os"]

  provisioner "ansible" {
    playbook_file    = "/vagrant/samples/raspbian_ansible_chroot.yml"
    ansible_env_vars = [
      "ANSIBLE_FORCE_COLOR=1",
      "PYTHONUNBUFFERED=1",
    ]
    extra_arguments  = [
      # The following arguments are required for running Ansible within a chroot
      # See https://www.packer.io/plugins/provisioners/ansible/ansible#chroot-communicator for details
      "--connection=chroot",
      "--become-user=root",
      #  Ansible needs this to find the mount path
      "-e ansible_host=${build.MountPath}"
    ]
  }
}
