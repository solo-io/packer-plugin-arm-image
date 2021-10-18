#!/bin/bash
/usr/sbin/update-binfmts --enable qemu-arm >/dev/null 2>&1
/usr/sbin/update-binfmts --enable qemu-aarch64 >/dev/null 2>&1

PACKER=/bin/packer

echo running $PACKER

exec $PACKER "${@}"