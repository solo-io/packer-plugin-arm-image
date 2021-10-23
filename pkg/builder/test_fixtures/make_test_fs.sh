#!/bin/bash

# mount must be run under root
if [ $UID -ne 0 ]; then
    echo "Must be root to run this script."
    exit 1
fi

TEST_FS=img.bin
TEST_FS_SIZE=10
dd bs=1M if=/dev/zero of=${TEST_FS} seek=${TEST_FS_SIZE} count=0
mkfs.ext4 ${TEST_FS}

mkdir tmpmnt
mount -o loop ${TEST_FS} tmpmnt
curl -sSL https://dl-cdn.alpinelinux.org/alpine/v3.14/releases/armv7/alpine-minirootfs-3.14.2-armv7.tar.gz | tar -C tmpmnt -xzf -
umount tmpmnt
rmdir tmpmnt

# gzip to save space
gzip img.bin