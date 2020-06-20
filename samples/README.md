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

The pi will now create a new wifi access point, bridging it to the ethernet network.
For this to work, the pi needs to be connected to your router via an ethernet cable.

You shouldn't need to log-in to the pi, and as such, `local_ssh_public_key` is not strictly needed. It is used to allow secure remote access to the pi (disabling password login). if you don't care about that just remove the steps related to ssh from `build.pkg.hcl`.

If you don't see the wifi network, log-in to the pi. and get hostapd logs:

```
journalctl -u hostapd
```

And if you see `rfkill: WLAN soft blocked`, issue this command `sudo rfkill unblock 0`.