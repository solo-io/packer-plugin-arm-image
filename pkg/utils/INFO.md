Outputs of lsblk and udevad are presented here for usb drive, and sdcard in ro and rw mode.
The goal is to understand how to detect removable devices for flashing.

usb device:
```
/sbin/udevadm info --query=property --name /dev/sda
DEVLINKS=/dev/disk/by-uuid/2017-10-18-18-53-18-00 /dev/disk/by-id/usb-USB_2.0_USB_Flash_Driver_AA03000000011985-0:0 /dev/disk/by-label/Ubuntu\x2017.10\x20amd64 /dev/disk/by-path/pci-0000:00:14.0-usb-0:1:1.0-scsi-0:0:0:0
DEVNAME=/dev/sda
DEVPATH=/devices/pci0000:00/0000:00:14.0/usb1/1-1/1-1:1.0/host2/target2:0:0/2:0:0:0/block/sda
DEVTYPE=disk
ID_BUS=usb
ID_FS_BOOT_SYSTEM_ID=EL\x20TORITO\x20SPECIFICATION
ID_FS_LABEL=Ubuntu_17.10_amd64
ID_FS_LABEL_ENC=Ubuntu\x2017.10\x20amd64
ID_FS_TYPE=iso9660
ID_FS_USAGE=filesystem
ID_FS_UUID=2017-10-18-18-53-18-00
ID_FS_UUID_ENC=2017-10-18-18-53-18-00
ID_FS_VERSION=Joliet Extension
ID_INSTANCE=0:0
ID_MODEL=USB_Flash_Driver
ID_MODEL_ENC=USB\x20Flash\x20Driver
ID_MODEL_ID=1000
ID_PART_TABLE_TYPE=dos
ID_PART_TABLE_UUID=690a7a2e
ID_PATH=pci-0000:00:14.0-usb-0:1:1.0-scsi-0:0:0:0
ID_PATH_TAG=pci-0000_00_14_0-usb-0_1_1_0-scsi-0_0_0_0
ID_REVISION=1100
ID_SERIAL=USB_2.0_USB_Flash_Driver_AA03000000011985-0:0
ID_SERIAL_SHORT=AA03000000011985
ID_TYPE=disk
ID_USB_DRIVER=usb-storage
ID_USB_INTERFACES=:080650:
ID_USB_INTERFACE_NUM=00
ID_VENDOR=USB_2.0
ID_VENDOR_ENC=USB\x202.0\x20
ID_VENDOR_ID=090c
MAJOR=8
MINOR=0
SUBSYSTEM=block
TAGS=:systemd:
USEC_INITIALIZED=642597442277
```
```
lsblk -b --output NAME,SIZE,RO,RM,MODEL,UUID --json /dev/sda
{
   "blockdevices": [
      {"name": "sda", "size": "8039432192", "ro": "0", "rm": "1", "model": "USB Flash Driver", "uuid": "2017-10-18-18-53-18-00",
         "children": [
            {"name": "sda1", "size": "1501102080", "ro": "0", "rm": "1", "model": null, "uuid": "2017-10-18-18-53-18-00"},
            {"name": "sda2", "size": "2359296", "ro": "0", "rm": "1", "model": null, "uuid": "2D90-0993"}
         ]
      }
   ]
}
```

for eMMC R/O:
```
lsblk -b --output NAME,SIZE,RO,RM,MODEL,UUID --json  /dev/mmcblk0
{
   "blockdevices": [
      {"name": "mmcblk0", "size": "7990149120", "ro": "1", "rm": "0", "model": null, "uuid": null,
         "children": [
            {"name": "mmcblk0p1", "size": "66060288", "ro": "1", "rm": "0", "model": null, "uuid": "70CD-BC89"},
            {"name": "mmcblk0p2", "size": "7919894528", "ro": "1", "rm": "0", "model": null, "uuid": "8a9074c8-46fe-4807-8dc9-8ab1cb959010"}
         ]
      }
   ]
}
```
```
 /sbin/udevadm info --query=property --name /dev/mmcblk0DEVLINKS=/dev/disk/by-path/pci-0000:03:00.0-platform-rtsx_pci_sdmmc.0 /dev/disk/by-id/mmc-SD8G_0x00001ee8
DEVNAME=/dev/mmcblk0
DEVPATH=/devices/pci0000:00/0000:00:1c.1/0000:03:00.0/rtsx_pci_sdmmc.0/mmc_host/mmc0/mmc0:0001/block/mmcblk0
DEVTYPE=disk
ID_DRIVE_FLASH_SD=1
ID_DRIVE_MEDIA_FLASH_SD=1
ID_NAME=SD8G
ID_PART_TABLE_TYPE=dos
ID_PART_TABLE_UUID=529fc551
ID_PATH=pci-0000:03:00.0-platform-rtsx_pci_sdmmc.0
ID_PATH_TAG=pci-0000_03_00_0-platform-rtsx_pci_sdmmc_0
ID_SERIAL=0x00001ee8
MAJOR=179
MINOR=0
SUBSYSTEM=block
TAGS=:systemd:
USEC_INITIALIZED=642767454485
```

# for mmc r/w:
```
 /sbin/udevadm info --query=property --name /dev/mmcblk0DEVLINKS=/dev/disk/by-path/pci-0000:03:00.0-platform-rtsx_pci_sdmmc.0 /dev/disk/by-id/mmc-SD8G_0x00001ee8
DEVNAME=/dev/mmcblk0
DEVPATH=/devices/pci0000:00/0000:00:1c.1/0000:03:00.0/rtsx_pci_sdmmc.0/mmc_host/mmc0/mmc0:0001/block/mmcblk0
DEVTYPE=disk
ID_DRIVE_FLASH_SD=1
ID_DRIVE_MEDIA_FLASH_SD=1
ID_NAME=SD8G
ID_PART_TABLE_TYPE=dos
ID_PART_TABLE_UUID=529fc551
ID_PATH=pci-0000:03:00.0-platform-rtsx_pci_sdmmc.0
ID_PATH_TAG=pci-0000_03_00_0-platform-rtsx_pci_sdmmc_0
ID_SERIAL=0x00001ee8
MAJOR=179
MINOR=0
SUBSYSTEM=block
TAGS=:systemd:
USEC_INITIALIZED=642767454485
```

```
lsblk -b --output NAME,SIZE,RO,RM,MODEL,UUID --json  /dev/mmcblk0
{
   "blockdevices": [
      {"name": "mmcblk0", "size": "7990149120", "ro": "0", "rm": "0", "model": null, "uuid": null,
         "children": [
            {"name": "mmcblk0p1", "size": "66060288", "ro": "0", "rm": "0", "model": null, "uuid": "70CD-BC89"},
            {"name": "mmcblk0p2", "size": "7919894528", "ro": "0", "rm": "0", "model": null, "uuid": "8a9074c8-46fe-4807-8dc9-8ab1cb959010"}
         ]
      }
   ]
}
```

