#!/bin/bash
OUTPUT_DIR=$1/bin
VERSION="0.8.0"
if [ "$ARCH" == "amd64" ]; then
  ARCH="x86_64"
elif [ "$ARCH" == "arm64" ]; then
  if [ "$OS" == "linux" ]; then
    ARCH="aarch64"
  else
    ARCH="x86_64"
  fi
fi
VERSION_NAME="shellcheck-v${VERSION}.${OS}.${ARCH}"
SHELLCHECK=${OUTPUT_DIR}/shellcheck
if [ -e "${SHELLCHECK}" ] && [ "v$(${SHELLCHECK} --version | grep version: | cut -d' ' -f2)" == v${VERSION} ]; then
  echo "Shellcheck is already installed at ${OUTPUT_DIR}"
  exit
fi
  echo "Installing shellcheck ${SHELLCHECK}"
  set -x
  mkdir /tmp/shellcheck
  curl --fail --location -s https://github.com/koalaman/shellcheck/releases/download/v${VERSION}/"${VERSION_NAME}".tar.xz \
    | tar --no-same-owner --strip-component=1 -C /tmp/shellcheck -xJ
  mv /tmp/shellcheck/shellcheck "$SHELLCHECK"
  rm -rf /tmp/shellcheck
  set +x
  echo "Shellcheck $SHELLCHECK has been installed at $OUTPUT_DIR"
