# First Boot Provisioning / Setup for Raspberry Pi using Cloud-Init / Solo-IO Packer Plugin

[Cloud-Init][1] is a prominent tool not just for cloud images but for configuring / setting up
Raspberry Pi images on first-boot. Cloud-Init just requires some YAML configuration files to be 
able to setup the Pi without any manual intervention.

## How does it all work?

For this example we will setup an Ubuntu 22.04 ARM64 Image with:

1. a user called `admin` with password `solorpi` which will have `sudo` rights escalation
2. hostname for the pi to be set as `solorpi`
3. Timezone for the image set to `Europe/Berlin`
4. Install Image with APT package `cowsay`

the `user-data` file with the above mentioned configuration has to persist in a the `/boot/firmware`
partition for the raspberry pi image for the cloud-init installer to set the image up. The installer will
also require an empty `meta-data` as well as `vendor-data` (this file is optional) file also in the `/boot/firmware`
partition. These tasks will be achieved using `file` provisioner from Packer.

### Cloud-Init: `user-data` file

```yaml
#cloud-config
ssh_pwauth: true
preserve_hostname: false
hostname: solorpi
package_upgrade: true
timezone: Europe/Berlin
users:
  - name: admin
    gecos: "SoloIO Admin"
    passwd: $6$xqBmC/BkZMzbERcn$YmzoBQ70fap9wYh1A8sYrj3An8isBLda4FW2KX.NLPFcV0jLo6ys5jkY1uTkuNj5T76DkpKkGcOCOLBQSbEYb0
    no_user_group: true
    lock_passwd: false
    groups: sudo, adm, lxd, dip, plugdev
    sudo: ALL=(ALL) NOPASSWD:ALL
    shell: /bin/bash
```
The password is generated using:

```bash
openssl passwd -6 # password is solorpi
```

Check Cloud-Init schema validity using:

```bash
cloud-init schema -c cloud-init/user-data
```

### Cloud-Init: `network-config` file

If you Pi needs to connect to a DHCP Server, Please adapt a `network-config` file according to your networking configuration and
uncomment the respective line in the `ubuntu.pkr.hcl` template file. Refer to [Netplan.io][2]

Example `network-config` file (replace your IP addresses / credentials accordingly):

```yaml
version: 2
ethernets:
  eth0:
    dhcp4: true
    optional: true
wifis:
  wlan0:
    dhcp4: false
    dhcp6: false
    addresses: [${STATIC_IP}/24]
    gateway4: ${NETWORK_IP}
    nameservers:
      addresses: [${NETWORK_IP}]
    access-points:
      "${WIFI_SSID}":
        password: "${WIFI_PASSWORD}"
```

### Usage

Install the Packer Plugin

```bash
packer init ubuntu.pkr.hcl
```

Validate the Packer template:

```bash
packer validate ubuntu.pkr.hcl
```

Build the Image using (might require sudo privileges to access `/dev/loop` devices):

```bash
packer build ubuntu.pkr.hcl
```

Once the build is done, flash the image on a SD-Card, power the Pi up and additionally, connect a monitor
to the Pi to see the Cloud-Init Logs. The process will install the security APT packages, install the Device Tree,
and `cowsay` package.

Upon prompt, the user `admin` (with password `solorpi`) will be available on your Pi.


### Build / Test Environment

Build tested on:

- __Manjaro Linux__ Distribution
- `qemu-system-aarch64 --version`: _7.0.0_
- `packer --version`: _1.8.3_
- `cloud-init --version`: _22.3.1_


Tested on:

- __Raspberry Pi 4 Model B Rev 1.1__

[1]: https://cloud-init.io
[2]: https://netplan.io/reference