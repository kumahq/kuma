#!/usr/bin/env bash

set -e

OUTPUT_DIR=$1/bin
VERSION="5.4.7"

if [[ $2 == "get-version" ]]; then
  echo ${VERSION}
else
  curl --fail --location -s https://raw.githubusercontent.com/rancher/k3d/4709d6adb24b23721f471e667e7301fa673b5efc/install.sh | \
            PATH=${OUTPUT_DIR}:${PATH} TAG=v${VERSION} USE_SUDO="false" K3D_INSTALL_DIR="${OUTPUT_DIR}" bash
fi
