#!/bin/bash

# Retrieve the system's architecture
ARCH=$(uname -m)
case $ARCH in
    i386|i686) ARCH=i386 ;;
    armv6*) ARCH=armv6 ;;
    armv7*) ARCH=armv7 ;;
    aarch64*) ARCH=arm64 ;;
esac

# Prepare the download URL
GITHUB_LATEST_VERSION=$(curl -L -s -H 'Accept: application/json' https://github.com/will-moss/isaiah/releases/latest | sed -e 's/.*"tag_name":"\([^"]*\)".*/\1/')
GITHUB_FILE="isaiah_${GITHUB_LATEST_VERSION//v/}_$(uname -s)_${ARCH}.tar.gz"
GITHUB_URL="https://github.com/will-moss/isaiah/releases/download/${GITHUB_LATEST_VERSION}/${GITHUB_FILE}"

# Install/Update the local binary
curl -L -o isaiah.tar.gz $GITHUB_URL
tar xzvf isaiah.tar.gz isaiah
mv isaiah /usr/bin/
chmod 755 /usr/bin/isaiah
rm isaiah.tar.gz
