#!/usr/bin/env bash

# This script fetches Envoy binary from download.konghq.com
#
# Requires:
# - $BINARY_PATH, path where binary will be fetched, for example 'out/envoy'
# - $ENVOY_DISTRO, name of the distributive (i.e darwin, linux)
#
# Optional:
# - $ENVOY_TAG, git tag to reference specific revision
# - $ENVOY_COMMIT_HASH, hash of the git commit. If specified, then $ENVOY_TAG will be ignored
#
# at least one of $ENVOY_TAG or $ENVOY_COMMIT_HASH should be specified

set -o errexit
set -o pipefail
set -o nounset

function msg_red() {
  builtin echo -en "\033[1;31m" >&2
  echo "$@" >&2
  builtin echo -en "\033[0m" >&2
}

function msg_err() {
  msg_red "$@"
  exit 1
}

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

ENVOY_TAG=${ENVOY_TAG:-}
ENVOY_COMMIT_HASH=${ENVOY_COMMIT_HASH:-}
[[ -z "${ENVOY_TAG}" ]] && [[ -z "${ENVOY_COMMIT_HASH}" ]] && msg_err "Error: either ENVOY_TAG or ENVOY_COMMIT_HASH should be specified"

if [ "${ENVOY_DISTRO}" == "linux" ]; then
  ENVOY_DISTRO="alpine"
fi

if [ "${ENVOY_DISTRO}" == "centos7" ]; then
  ENVOY_DISTRO="centos"
fi

if [[ -n "${ENVOY_COMMIT_HASH}" ]]; then
  ENVOY_SHORT_HASH=${ENVOY_COMMIT_HASH:0:8}

  BINARY_NAME=$(curl --silent https://download.konghq.com/mesh-alpine/ \
    | { grep "${ENVOY_SHORT_HASH}" || true; } \
    | { grep "${ENVOY_DISTRO}" || true; } \
    | sed -e 's#.*<li><a href=".*">\(.*\)</a></li>#\1#')

  [[ -z "${BINARY_NAME}" ]] && msg_err "failed to resolve binary name by ENVOY_COMMIT_HASH=${ENVOY_COMMIT_HASH}"

  download_envoy "${BINARY_NAME}"
  exit 0
fi

if [[ -n "${ENVOY_TAG}" ]]; then
  ENVOY_VERSION=${ENVOY_TAG:1}
  BINARY_NAME="envoy-${ENVOY_VERSION}-${ENVOY_DISTRO}"

  download_envoy "${BINARY_NAME}"
  exit 0
fi
