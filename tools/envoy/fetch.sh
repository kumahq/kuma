#!/usr/bin/env bash

# This script fetches Envoy binary from download.konghq.com
#
# Requires:
# - $BINARY_PATH, path where binary will be fetched, for example 'out/envoy'
# - $ENVOY_TAG, tag of the envoy binary to fetch
# - $ENVOY_DISTRO, name of the distributive (i.e darwin, alpine)
# - $ENVOY_ARTIFACT_EXT, optional artifact suffix (i.e. -extended)

set -o errexit
set -o pipefail
set -o nounset

source "$(dirname -- "${BASH_SOURCE[0]}")/../common.sh"

function download_envoy() {
    local binary_name=$1
    local bin_dir
    bin_dir=$(dirname "${BINARY_PATH}")
    echo "Downloading ${binary_name}"

    if [ ! -d "$(dirname "${BINARY_PATH}")" ]; then
      mkdir -p "$(dirname "${BINARY_PATH}")"
    fi

    local tar_path="${BINARY_PATH}.tar.gz"

    local status
    status=$(curl -# --location --output "${tar_path}" --write-out '%{http_code}' \
    "https://github.com/kumahq/envoy-builds/releases/download/${ENVOY_TAG}/${binary_name}.tar.gz")

    if [ "$status" -ne "200" ]; then
        msg_err "Error: failed downloading Envoy: ${status} error"
    fi

    tar -C "$bin_dir" -xzvf "${tar_path}"
    rm "$tar_path"
    mv "${bin_dir}/envoy-${ENVOY_DISTRO}" "${BINARY_PATH}"

    [ -f "${BINARY_PATH}" ] && chmod +x "${BINARY_PATH}"
}

if [[ -n "${ENVOY_TAG}" ]]; then
  BINARY_NAME="envoy-${GOOS}-${GOARCH}-${ENVOY_TAG}-${ENVOY_DISTRO}-opt${ENVOY_ARTIFACT_EXT}"
  download_envoy "${BINARY_NAME}"
  exit 0
fi
