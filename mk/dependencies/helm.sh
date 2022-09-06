#!/bin/bash

OUTPUT_DIR=$1/bin
VERSION="3.8.2"
curl --fail --location -s https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | \
	HELM_INSTALL_DIR=${OUTPUT_DIR} DESIRED_VERSION=v${VERSION} USE_SUDO=false bash
