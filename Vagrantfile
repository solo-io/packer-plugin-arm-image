# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/focal64"
  config.vm.synced_folder "./", "/vagrant", disabled: false
  config.vm.provision "build-env", type: "shell", :path => "provision-build-env.sh", privileged: false
  config.vm.provision "packer-plugin-arm-image", type: "shell", :path => "provision-packer-plugin-arm-image.sh", privileged: false, env: {"GIT_CLONE_URL" => ENV["GIT_CLONE_URL"]}
  config.vm.provision "build-image", type: "shell", :path => "provision-build-image.sh", privileged: false, env: {"PACKERFILE" => ENV["PACKERFILE"]}
end
