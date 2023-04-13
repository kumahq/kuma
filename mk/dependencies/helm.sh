#!/bin/bash

set -e

OUTPUT_DIR=$1/bin
VERSION="3.8.2"
export PATH="$OUTPUT_DIR:$PATH" # install script checks if helm is in your path
curl --fail --location -s https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | \
	HELM_INSTALL_DIR=${OUTPUT_DIR} DESIRED_VERSION=v${VERSION} USE_SUDO=false bash

CR_VERSION="1.3.0"
curl --fail --location -s "https://github.com/helm/chart-releaser/releases/download/v${CR_VERSION}/chart-releaser_${CR_VERSION}_${OS}_${ARCH}.tar.gz" | tar xvz --directory "${OUTPUT_DIR}" cr
