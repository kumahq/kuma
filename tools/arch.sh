#!/bin/bash

ARCH=$1
OS=$2

if [ "$ARCH" == "amd64" ]; then
    ARCH="x86_64"
elif [ "$ARCH" == "arm64" ]; then
    ARCH="aarch64"
fi

echo $ARCH
