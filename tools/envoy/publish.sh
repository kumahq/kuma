#!/bin/bash

# This script publishes binary passed to $1
#
# Requirements
# * Docker

set -o errexit
set -o pipefail
set -o nounset

source "$(dirname -- "${BASH_SOURCE[0]}")/../common.sh"

if [ $# -eq 0 ]; then
    echo "Usage: ./publish.sh path_to_envoy"
    echo "Example: ./publish.sh out/envoy-v1.20.0"
    exit 1
fi

BINARY_PATH=$1

PULP_HOST="https://api.pulp.konnect-prod.konghq.com"
PULP_PACKAGE_TYPE="mesh"
PULP_DIST_NAME="alpine"

[ -z "$PULP_USERNAME" ] && msg_err "PULP_USERNAME required"
[ -z "$PULP_PASSWORD" ] && msg_err "PULP_PASSWORD required"

docker run --rm \
        -e PULP_USERNAME="${PULP_USERNAME}" \
        -e PULP_PASSWORD="${PULP_PASSWORD}" \
        -e PULP_HOST="${PULP_HOST}" \
        -v "${PWD}":/files:ro -it kong/release-script \
        --file /files/"${BINARY_PATH}" \
        --package-type "${PULP_PACKAGE_TYPE}" --dist-name "${PULP_DIST_NAME}" --publish
