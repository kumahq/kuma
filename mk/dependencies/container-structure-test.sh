#!/bin/bash

set -e

OUTPUT_DIR=$1/bin
VERSION="1.15.0"
NAME="container-structure-test"
BASE_URL="https://storage.googleapis.com/container-structure-test"
CONTAINER_STRUCTURE_TEST="${OUTPUT_DIR}/${NAME}"

if [ -e "${CONTAINER_STRUCTURE_TEST}" ] && [ "$("${CONTAINER_STRUCTURE_TEST}" version)" == "v${VERSION}" ]; then
  echo "${NAME} v${VERSION} is already installed at ${OUTPUT_DIR}"
  exit
fi

echo "Installing ${NAME} ${VERSION}..."

curl --fail --location --silent \
  --output "${CONTAINER_STRUCTURE_TEST}" \
  "${BASE_URL}/v${VERSION}/${NAME}-${OS}-${ARCH}"

chmod u+x "${CONTAINER_STRUCTURE_TEST}"
