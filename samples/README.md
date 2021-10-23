# Samples

This directory contains various samples.

## HostAPD

Allows you to use your raspberry pi as a WiFi extender.

build like so:
```
PASSWORD=$(head -c 1024 /dev/urandom | tr -dc 'a-zA-Z0-9' | cut -c -12)
docker run \
  --rm \
  --privileged \
  -v /dev:/dev \
  -v ${PWD}:/build:ro \
  -v ${PWD}/packer_cache:/build/packer_cache \
  -v ${PWD}/output-arm-image:/build/output-arm-image \
  -v ${HOME}/.ssh/id_rsa.pub:/config/id_rsa.pub:ro \
  -e PACKER_CACHE_DIR=/build/packer_cache \
  -w /build/hostapd \
  ghcr.io/solo-io/packer-plugin-arm-image:v0.2.3 build -var wifi_ssid=wifi_extender -var wifi_psk=$PASSWORD -var local_ssh_public_key=/config/id_rsa.pub .
```

The pi will now create a new wifi access point, bridging it to the ethernet network.
For this to work, the pi needs to be connected to your router via an ethernet cable.

You shouldn't need to log-in to the pi, and as such, `local_ssh_public_key` is not strictly needed. It is used to allow secure remote access to the pi (disabling password login). if you don't care about that just remove the steps related to ssh from `build.pkg.hcl`.

If you don't see the wifi network, log-in to the pi. and get hostapd logs:

```
journalctl -u hostapd
```

And if you see `rfkill: WLAN soft blocked`, issue this command `sudo rfkill unblock 0`.


## Kubernetes

Install an image that has what you need to install a k8s node:

```
docker run \
  --rm \
  --privileged \
  -v /dev:/dev \
  -v ${PWD}:/build:ro \
  -v ${PWD}/packer_cache:/build/packer_cache \
  -v ${PWD}/output-arm-image:/build/output-arm-image \
  -v ${HOME}/.ssh/id_rsa.pub:/config/id_rsa.pub:ro \
  -e PACKER_CACHE_DIR=/build/packer_cache \
  -w /build/k8s \
  ghcr.io/solo-io/packer-plugin-arm-image:v0.2.3 build -var local_ssh_public_key=/config/id_rsa.pub .
```

or, run as root:
```
PACKER_CONFIG_DIR=$HOME sudo -E $(which packer) build -var local_ssh_public_key=$HOME/.ssh/id_rsa.pub .
```

**Note**: This image doesn't result with kubernetes installed. Instead, it just sets up and image with the binaries needed to install it.
Specifically, this example automates the [Installing kubeadm](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/) section of the ["Bootstrapping clusters with kubeadm"](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/) kubernetes documentaiton.

The reason we don't go further is that installing k8s cluster is require you to make a few decisions (CNI for example), and has a different flow if it's a new install or a node joining an existing cluster.
Once running, follow the instructions here to install it: https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/