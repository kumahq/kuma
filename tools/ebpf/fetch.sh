#!/usr/bin/env bash

# This script fetches ebpf programs
#
# Requires:
# - $OUTPUT_PATH, path where ebpf programs will be fetched, i.e. 'out/ebpf'
# - $MERBRIDGE_TAG, name of the release tag to use to download ebpf programs
# Optional:
# - $RELEASE_REPO, repository which contains released merbridge ebpf programs
#   (default: https://github.com/kumahq/merbridge)
# - $TARBALL_NAME, name of the released tarball with all ebpf programs
#   (default: all.tar.gz)

set -o errexit
set -o pipefail
set -o nounset

CURRENT_RELEASE_REPO="${RELEASE_REPO:-https://github.com/kumahq/merbridge}"
CURRENT_TARBALL_NAME="${TARBALL_NAME:-all.tar.gz}"

# shellcheck source=./../common.sh
source "$(dirname -- "${BASH_SOURCE[0]}")/../common.sh"

function download_ebpf() {
  local release_name
  release_name=${1}

  local output_path
  output_path=${2}

  local release_repo
  release_repo=${3:-${CURRENT_RELEASE_REPO}}

  local tarball_name
  tarball_name=${4:-${CURRENT_TARBALL_NAME}}

  local url
  url="${release_repo}/releases/download/${release_name}/${tarball_name}"

  echo "Downloading ${release_name} from ${url}"

  if [ ! -d "${output_path}" ]; then
    mkdir -p "${output_path}"
  fi

  local tar_path="${output_path}/ebpf.tar.gz"

  local status
  status=$(curl -# --location --output "${tar_path}" --write-out '%{http_code}' \
    "${url}")

  if [ "$status" -ne "200" ]; then
    msg_err "Error: failed downloading ebpf programs: ${status} error"
  fi

  tar -C "${output_path}" -xzvf "${tar_path}"
  rm "${tar_path}"
}

if [[ -n "${MERBRIDGE_TAG}" ]]; then
  if [[ -z "${OUTPUT_PATH}" ]]; then
    msg_err "Error: \${OUTPUT_PATH} has to be specified"
  fi

  download_ebpf \
    "${MERBRIDGE_TAG}" \
    "${OUTPUT_PATH}" \
    "${CURRENT_RELEASE_REPO}" \
    "${CURRENT_TARBALL_NAME}"

  exit 0
fi
