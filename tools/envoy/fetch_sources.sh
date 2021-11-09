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
set -e

function msg_red() {
  builtin echo -en "\033[1;31m" >&2
  echo "$@" >&2
  builtin echo -en "\033[0m" >&2
}

function msg_err() {
  msg_red "$@"
  exit 1
}

[[ -z "${SOURCE_DIR}" ]] && msg_err "Error: SOURCE_DIR is not specified"
[[ -z "${ENVOY_TAG}" ]] && [[ -z "${ENVOY_COMMIT_HASH}" ]] && msg_err "Error: either ENVOY_TAG or ENVOY_COMMIT_HASH should be specified"

# clone Envoy repo if not exists
if [[ ! -d "${SOURCE_DIR}" ]]; then
  git clone https://github.com/envoyproxy/envoy.git ${SOURCE_DIR}
else
  echo "Envoy source directory already exists, just fetching"
  pushd ${SOURCE_DIR} && git fetch --all && popd
fi

pushd ${SOURCE_DIR}

ENVOY_COMMIT_HASH=${ENVOY_COMMIT_HASH:-$(git rev-list -n 1 "${ENVOY_TAG}")}

echo "ENVOY_TAG=${ENVOY_TAG}"
echo "ENVOY_COMMIT_HASH=${ENVOY_COMMIT_HASH}"

git reset --hard ${ENVOY_COMMIT_HASH}

popd
