#!/bin/bash
set -e

OUTPUT_BIN_DIR=$1/bin
VERSION=${GOLANGCI_LINT_VERSION}

golangcilint="${OUTPUT_BIN_DIR}"/golangci-lint
if [ "${VERSION}" == "" ]; then
  echo "No version specified for golangci-lint"
  exit 1
fi

if [ -e "${golangcilint}" ] && [ "v$(${golangcilint} version --format short)" == "${VERSION}" ]; then
  echo "golangci-lint ${VERSION} is already installed at ${OUTPUT_BIN_DIR}"
  exit
fi
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/e68d278319b6d0a68680e3389bc0576ef39ec02b/install.sh | sh -s -- -b "${OUTPUT_BIN_DIR}" "${VERSION}"
