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
  -v ${PWD}:/build:ro \
  -v ${PWD}/packer_cache:/build/packer_cache \
  -v ${PWD}/output-arm-image:/build/output-arm-image \
  -v ${HOME}/.ssh/id_rsa.pub:/config/id_rsa.pub:ro \
  -e PACKER_CACHE_DIR=/build/packer_cache \
  -w /build/hostapd \
  quay.io/solo-io/packer-builder-arm-image:v0.1.5 build -var wifi_ssid=wifi_extender -var wifi_psk=$PASSWORD -var local_ssh_public_key=/config/id_rsa.pub .
```