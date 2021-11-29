#!/usr/bin/env bash

# This script fetches Envoy binary from download.konghq.com
#
# Requires:
# - $BINARY_PATH, path where binary will be fetched, for example 'out/envoy'
# - $ENVOY_VERSION, version of the envoy binary to fetch
# - $ENVOY_DISTRO, name of the distributive (i.e darwin, alpine)

set -o errexit
set -o pipefail
set -o nounset

source "$(dirname -- "${BASH_SOURCE[0]}")/../common.sh"

function download_envoy() {
    local binary_name=$1
    echo "Downloading ${binary_name}"

    if [ ! -d "$(dirname "${BINARY_PATH}")" ]; then
      mkdir -p "$(dirname "${BINARY_PATH}")"
    fi

    local status=$(curl -# --location --output "${BINARY_PATH}" --write-out %{http_code} \
    "https://download.konghq.com/mesh-alpine/${binary_name}")

  [ -f "${BINARY_PATH}" ] && chmod +x "${BINARY_PATH}"
  [ "$status" -ne "200" ] && msg_err "Error: failed downloading Envoy" || true
}

if [[ -n "${ENVOY_VERSION}" ]]; then
  BINARY_NAME="envoy-${ENVOY_VERSION}-${ENVOY_DISTRO}"
  download_envoy "${BINARY_NAME}"
  exit 0
fi
