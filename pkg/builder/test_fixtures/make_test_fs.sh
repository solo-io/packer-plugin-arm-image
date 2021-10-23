#!/bin/bash

# mount must be run under root
if [ $UID -ne 0 ]; then
    echo "Must be root to run this script."
    exit 1
fi

TEST_FS=img.bin
TEST_FS_SIZE=30
dd bs=1M if=/dev/zero of=${TEST_FS} seek=${TEST_FS_SIZE} count=0

sfdisk ${TEST_FS} < part-layout

LO=$(losetup --show -f -P ${TEST_FS})
ls ${LO}*

DEV=${LO}p1

mkfs.ext4 ${DEV}

mkdir tmpmnt
mount ${DEV} tmpmnt
curl -sSL https://dl-cdn.alpinelinux.org/alpine/v3.14/releases/armv7/alpine-minirootfs-3.14.2-armv7.tar.gz | tar -C tmpmnt -xzf -
umount tmpmnt
rmdir tmpmnt

losetup -d ${LO}

# gzip to save space
gzip img.bin