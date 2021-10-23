#!/bin/bash

cd "$(dirname "$0")"

FILES=$(cut -c67- sha256sums)

NEED_VERIFY=0
for f in $FILES; do
    if [ ! -f "$f.gz" ]; then
        NEED_VERIFY=1
    fi
done

if [ $NEED_VERIFY -eq 0 ]; then
    exit 0
fi

echo "Downloading binaries..."

for f in $FILES; do
    echo "Downloading $f"
    curl -L -O https://github.com/multiarch/qemu-user-static/releases/download/v6.1.0-6/$f
done

    # verify checksum:
if sha256sum --check sha256sums; then
    echo "Checksum OK"
    for f in $FILES; do
        echo "Gzipping $f"
        gzip $f
    done
else
    echo "Checksum FAILED"
    for f in $FILES; do
        echo "Removing $f"
        rm $f
    done
    exit 1
fi
