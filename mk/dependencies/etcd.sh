#!/bin/bash
OUTPUT_DIR=$1/bin
VERSION="3.5.3"
ETCD=${OUTPUT_DIR}/etcd
# There's no etcd for arm64 mac so let's just use the amd64
[[ ${OS} == "darwin" ]] && ARCH="amd64"

VERSION_NAME="etcd-v${VERSION}-${OS}-${ARCH}"
if [ -e "$ETCD" ] && [ "v$($ETCD --version | head -1 | cut -f3 -d ' ')" == "v${VERSION}" ]; then
  echo "$(${ETCD} --version | head -1) is already installed at ${OUTPUT_DIR}"
  exit
fi
echo "Installing etcd ${VERSION}..."
set -x
FNAME=${VERSION_NAME}.tar.gz
if [ "${OS}" != "linux" ]; then
  FNAME=${VERSION_NAME}.zip
fi
curl --location --fail -s https://github.com/etcd-io/etcd/releases/download/v${VERSION}/"${FNAME}" | tar --strip-components=1 --no-same-owner -xz -C "${OUTPUT_DIR}" "${VERSION_NAME}"/etcd
chmod u+x "${ETCD}"
set +x
