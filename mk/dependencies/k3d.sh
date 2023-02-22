#!/bin/bash

set -e

OUTPUT_DIR=$1/bin
VERSION="5.4.7"
# see https://raw.githubusercontent.com/rancher/k3d/main/install.sh
curl --fail --location -s https://raw.githubusercontent.com/rancher/k3d/main/install.sh | \
          PATH=${OUTPUT_DIR}:${PATH} TAG=v${VERSION} USE_SUDO="false" K3D_INSTALL_DIR="${OUTPUT_DIR}" bash
