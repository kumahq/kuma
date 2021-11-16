#!/bin/bash

# This script fetches Envoy source code to $SOURCE_DIR
#
# Requires:
# - $SOURCE_DIR, a directory where sources will be placed
#
# Optional:
# - $ENVOY_TAG, git tag to reference specific revision
# - $ENVOY_COMMIT_HASH, hash of the git commit. If specified, then $ENVOY_TAG will be ignored
#
# at least one of $ENVOY_TAG or $ENVOY_COMMIT_HASH should be specified

set -o errexit
set -o pipefail
set -o nounset

source "$(dirname -- "${BASH_SOURCE[0]}")/../common.sh"

ENVOY_TAG=${ENVOY_TAG:-}
ENVOY_COMMIT_HASH=${ENVOY_COMMIT_HASH:-}
[[ -z "${ENVOY_TAG}" ]] && [[ -z "${ENVOY_COMMIT_HASH}" ]] && msg_err "Error: either ENVOY_TAG or ENVOY_COMMIT_HASH should be specified"

# clone Envoy repo if not exists
if [[ ! -d "${SOURCE_DIR}" ]]; then
  mkdir -p "${SOURCE_DIR}"
  (
    cd "${SOURCE_DIR}"
    git init .
    git remote add origin https://github.com/envoyproxy/envoy.git
  )
else
  echo "Envoy source directory already exists, just fetching"
  pushd ${SOURCE_DIR} && git fetch --all && popd
fi

pushd ${SOURCE_DIR}

git fetch origin --depth=1 "${ENVOY_COMMIT_HASH:-${ENVOY_TAG}}"
git reset --hard FETCH_HEAD

echo "ENVOY_TAG=${ENVOY_TAG}"
echo "ENVOY_COMMIT_HASH=${ENVOY_COMMIT_HASH}"

popd
