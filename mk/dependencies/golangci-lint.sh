#!/bin/bash
set -e

OUTPUT_BIN_DIR=$1/bin
VERSION="1.51.2"

golangcilint="${OUTPUT_BIN_DIR}"/golangci-lint

if [ -e "${golangcilint}" ] && [ "$(${golangcilint} version --format short)" == "${VERSION}" ]; then
  echo "golangci-lint ${VERSION} is already installed at ${OUTPUT_BIN_DIR}"
  exit
fi
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "${OUTPUT_BIN_DIR}" "v${VERSION}"
