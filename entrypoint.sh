#!/bin/bash
/usr/sbin/update-binfmts --enable qemu-arm >/dev/null 2>&1

PACKER=/bin/packer

if [ ! -f $PACKER ]; then 
    # edge case: we use latest release from github
    PACKER=/bin/pkg/packer_linux_amd64
fi
echo running $PACKER

exec $PACKER "${@}"