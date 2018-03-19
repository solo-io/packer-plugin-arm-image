# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vm.box = "generic/ubuntu1604"
  config.vm.provider :virtualbox do |vb, override|
    # generic/ubuntu1604 does not come with virtualbox-guest-utils installed even
    # though a virtualbox flavor exists.
    # Therefore, override the image for virtualbox provider.
    override.vm.box = "ubuntu/xenial64"
    # disable the generation of ubuntu-xenial-16.04-cloudimg-console.log file
    # https://betacloud.io/get-rid-of-ubuntu-xenial-16-04-cloudimg-console-log/
    vb.customize [ "modifyvm", :id, "--uartmode1", "disconnected" ]
  end
  config.vm.synced_folder "./", "/vagrant", disabled: false

  config.vm.provision "build-env", type: "shell", :path => "provision-build-env.sh", privileged: false
  config.vm.provision "packer-builder-arm-image", type: "shell", :path => "provision-packer-builder-arm-image.sh", privileged: false, env: {"GIT_CLONE_URL" => ENV["GIT_CLONE_URL"]}
  config.vm.provision "build-image", type: "shell", :path => "provision-build-image.sh", privileged: false, env: {"PACKERFILE" => ENV["PACKERFILE"]}
end
