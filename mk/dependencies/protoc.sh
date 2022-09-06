#!/bin/bash
OUTPUT_DIR=$1/bin
VERSION="3.20.0"
PROTOC=${OUTPUT_DIR}/protoc
WKT_DIR=${1}/protos/google/protobuf
if [ "${OS}" == "darwin" ]; then
  OS="osx"
fi
if [ "${ARCH}" == "amd64" ]; then
  ARCH="x86_64"
elif [ "${ARCH}" == "arm64" ]; then
  ARCH="aarch_64"
fi

VERSION_NAME=protoc-${VERSION}-${OS}-${ARCH}
if [ -e "$PROTOC" ] && [ -e "$WKT_DIR" ] && [ "v$("$PROTOC" --version | cut -f2 -d ' ')" == "v${VERSION}" ]; then
  echo "$($PROTOC --version) is already installed at ${OUTPUT_DIR}"
  exit
fi
echo "Installing Protoc ${PROTOC} ${VERSION} ..."
rm -rf "${PROTOC}"
rm -rf "${WKT_DIR}"
set -x
mkdir -p /tmp/${VERSION_NAME}
curl --location --fail -s -o /tmp/${VERSION_NAME}.zip https://github.com/protocolbuffers/protobuf/releases/download/v${VERSION}/${VERSION_NAME}.zip
unzip /tmp/${VERSION_NAME}.zip bin/protoc 'include/*' -d /tmp/${VERSION_NAME}
cp /tmp/"${VERSION_NAME}"/bin/protoc "${PROTOC}"
mkdir -p "${WKT_DIR}"
cp -r /tmp/"${VERSION_NAME}"/include/google/protobuf/* "${WKT_DIR}"
rm -rf /tmp/"${VERSION_NAME}"*
set +x
echo "Protoc ${VERSION} has been installed at ${PROTOC}" ;
