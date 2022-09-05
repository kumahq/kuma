#!/bin/bash

# see https://book.kubebuilder.io/quick-start.html#installation
OUTPUT_DIR=$1/bin
VERSION="2.3.2"
KUBEBUILDER="${OUTPUT_DIR}"/kubebuilder
VERSION_NAME=kubebuilder_"${VERSION}"_"${OS}"_"${ARCH}"
if [ -e "${KUBEBUILDER}" ] && [ "v$("${KUBEBUILDER}" version  | sed -E 's/.*KubeBuilderVersion:"([0-9\.]+)".*/\1/')" == "v${VERSION}" ]; then
  echo "kubebuilder version ${VERSION} is already installed at ${KUBEBUILDER}"
  exit
fi
echo "Installing Kubebuilder ${CI_KUBEBUILDER_VERSION} ..."
rm -rf "${KUBEBUILDER}"
set -x
curl --location --fail -s https://github.com/kubernetes-sigs/kubebuilder/releases/download/v"${VERSION}"/"${VERSION_NAME}".tar.gz \
  | tar --strip-components=2 -xz -C "${OUTPUT_DIR}" "${VERSION_NAME}"/bin/kubebuilder
set +x
echo "kubebuilder ${CI_KUBEBUILDER_VERSION} and dependencies has been installed at ${KUBEBUILDER}"
