#!/bin/bash

DESTINATION="/usr/bin"
if [ -d "/usr/local/bin" ]; then
  DESTINATION="/usr/local/bin"
fi

# Handle sudo requirement on default install location
if [ $(id -u) -ne 0 ]; then
  echo "By default, Isaiah attempts to install its binary in /usr/bin/"
  echo "but that requires root permission. You can either restart the"
  echo "install script using sudo, or provide a new installation directory."


  read -r -p "New installation directory: " DESTINATION
  if [ ! -d $DESTINATION ]; then
    echo "Error: No such directory"
    exit
  fi

  # Remove trailing slash if any
  DESTINATION=${DESTINATION%/}
fi


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


mv isaiah $DESTINATION
chmod 755 $DESTINATION/isaiah
rm isaiah.tar.gz
