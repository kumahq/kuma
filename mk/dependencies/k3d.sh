#!/bin/bash

set -e

OUTPUT_DIR=$1/bin
<<<<<<< HEAD
VERSION="5.4.7"
# see https://raw.githubusercontent.com/rancher/k3d/main/install.sh
curl --fail --location -s https://raw.githubusercontent.com/rancher/k3d/main/install.sh | \
          PATH=${OUTPUT_DIR}:${PATH} TAG=v${VERSION} USE_SUDO="false" K3D_INSTALL_DIR="${OUTPUT_DIR}" bash
=======
VERSION="5.7.4"

if [[ $2 == "get-version" ]]; then
  echo ${VERSION}
else
  # see https://raw.githubusercontent.com/rancher/k3d/main/install.sh
  curl --fail --location -s https://raw.githubusercontent.com/rancher/k3d/main/install.sh | \
            PATH=${OUTPUT_DIR}:${PATH} TAG=v${VERSION} USE_SUDO="false" K3D_INSTALL_DIR="${OUTPUT_DIR}" bash
fi
>>>>>>> 529694bad (ci(k8s): download calico manifests as needed (#11851))
