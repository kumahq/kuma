#!/usr/bin/env bash

set -e

OUTPUT_DIR=$1/bin
VERSION="0.24.0"
KIND=${OUTPUT_DIR}/kind
if [ -e "$KIND" ] && [ "v$($KIND --version | cut -d ' ' -f3)" == v${VERSION} ]; then
  echo "$($KIND --version ) is already installed at ${OUTPUT_DIR}" ;
  exit
fi
echo "Installing kind ${VERSION} ..."
set -x
# see https://kind.sigs.k8s.io/docs/user/quick-start/#installation
curl --location --fail -s -o "${KIND}" https://github.com/kubernetes-sigs/kind/releases/download/v${VERSION}/kind-"${OS}"-"${ARCH}"
chmod +x "${KIND}"
set +x
echo "Kind $VERSION has been installed at $OUTPUT_DIR"
