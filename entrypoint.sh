#!/bin/bash
/usr/sbin/update-binfmts --enable qemu-arm >/dev/null 2>&1
/bin/packer "${@}"