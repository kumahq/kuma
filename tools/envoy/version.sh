#!/bin/bash

set -o errexit
set -o pipefail
set -o nounset

# Returns Envoy version by ENVOY_TAG:
# - if ENVOY_TAG is a real git tag like 'v1.20.0' then the version is equal to '1.20.0' (without the first letter 'v').
# - if ENVOY_TAG is a commit hash then the version will look like '1.20.1-dev-b16d390f'

ENVOY_TAG=${ENVOY_TAG:-"v1.22.1"}
ENVOY_VERSION=$(curl --silent --location "https://raw.githubusercontent.com/envoyproxy/envoy/${ENVOY_TAG}/VERSION.txt")

# for envoy versions older than v1.22.0 file 'VERSION.txt' used to be called 'VERSION'
if [[ "${ENVOY_VERSION}" == "404: Not Found" ]]; then
  ENVOY_VERSION=$(curl --silent --location --fail "https://raw.githubusercontent.com/envoyproxy/envoy/${ENVOY_TAG}/VERSION")
fi

if [[ "${ENVOY_TAG}" =~ ^v[0-9]*\.[0-9]*\.[0-9]*$ ]]; then
  echo "${ENVOY_VERSION}"
else
  echo "${ENVOY_VERSION}-${ENVOY_TAG:0:8}"
fi
