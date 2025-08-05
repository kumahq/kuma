#!/usr/bin/env bash

set -e

OUTPUT_DIR=$1/bin
VERSION="1.15.0"
NAME="container-structure-test"
BASE_URL="https://github.com/GoogleContainerTools/container-structure-test/releases/download"
CONTAINER_STRUCTURE_TEST="${OUTPUT_DIR}/${NAME}"

if [ "${OS}" == "darwin" ]; then
  ARCH='amd64'
fi
if [ -e "${CONTAINER_STRUCTURE_TEST}" ] && [ "$("${CONTAINER_STRUCTURE_TEST}" version)" == "v${VERSION}" ]; then
  echo "${NAME} v${VERSION} is already installed at ${OUTPUT_DIR}"
  exit
fi

echo "Installing ${NAME} ${VERSION}..."

curl --fail --location --silent \
  --output "${CONTAINER_STRUCTURE_TEST}" \
  "${BASE_URL}/v${VERSION}/${NAME}-${OS}-${ARCH}"

chmod u+x "${CONTAINER_STRUCTURE_TEST}"
