#!/usr/bin/env bash

set -e

OUTPUT_DIR=$1/bin
VERSION="2.12.0"
if [ "$ARCH" == "amd64" ]; then
  ARCH="x86_64"
fi
if [ "$OS" == "darwin" ]; then
  OS="Darwin"
  # Darwin does not have arm builds so we will use x86_64 via rosetta
  ARCH="x86_64"
elif [ "$OS" == "linux" ]; then
  OS="Linux"
fi
VERSION_NAME="hadolint-${OS}-${ARCH}"
hadolint=${OUTPUT_DIR}/hadolint
if [ -e "${hadolint}" ] && [ "v$(${hadolint} --version | cut -d' ' -f4)" == v${VERSION} ]; then
  echo "hadolint is already installed at ${OUTPUT_DIR}"
  exit
fi
echo "Installing hadolint ${hadolint}"
set -x
curl --output "$hadolint" --fail --location -s https://github.com/hadolint/hadolint/releases/download/v${VERSION}/"${VERSION_NAME}"
chmod +x "$hadolint"
set +x
echo "hadolint $hadolint has been installed at $OUTPUT_DIR"
