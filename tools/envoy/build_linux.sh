#!/bin/bash

set -o errexit
set -o pipefail
set -o nounset

echo "Building Envoy for Linux"

mkdir -p "$(dirname "${BINARY_PATH}")"

SOURCE_DIR="${SOURCE_DIR}" "${KUMA_DIR:-.}/tools/envoy/fetch_sources.sh"
CONTRIB_ENABLED_MATRIX_SCRIPT=$(realpath "${KUMA_DIR:-.}/tools/envoy/contrib_enabled_matrix.py")

BAZEL_BUILD_EXTRA_OPTIONS=${BAZEL_BUILD_EXTRA_OPTIONS:-""}
read -ra BAZEL_BUILD_EXTRA_OPTIONS <<< "${BAZEL_BUILD_EXTRA_OPTIONS}"
BAZEL_BUILD_OPTIONS=(
    "--config=clang"
    "--verbose_failures"
    "${BAZEL_BUILD_EXTRA_OPTIONS[@]+"${BAZEL_BUILD_EXTRA_OPTIONS[@]}"}")
BUILD_TARGET=${BUILD_TARGET:-"//contrib/exe:envoy-static"}

pushd "${SOURCE_DIR}"
CONTRIB_ENABLED_ARGS=$(python "${CONTRIB_ENABLED_MATRIX_SCRIPT}")
popd

BUILD_CMD=${BUILD_CMD:-"bazel build ${BAZEL_BUILD_OPTIONS[@]} -c opt ${BUILD_TARGET} ${CONTRIB_ENABLED_ARGS}"}

ENVOY_BUILD_SHA=$(curl --fail --location --silent https://raw.githubusercontent.com/envoyproxy/envoy/"${ENVOY_TAG}"/.bazelrc | grep envoyproxy/envoy-build-ubuntu | sed -e 's#.*envoyproxy/envoy-build-ubuntu:\(.*\)#\1#'| uniq)
ENVOY_BUILD_IMAGE="envoyproxy/envoy-build-ubuntu:${ENVOY_BUILD_SHA}"
LOCAL_BUILD_IMAGE="envoy-builder:${ENVOY_TAG}"

echo "BUILD_CMD=${BUILD_CMD}"

docker build -t "${LOCAL_BUILD_IMAGE}" --progress=plain \
  --build-arg ENVOY_BUILD_IMAGE="${ENVOY_BUILD_IMAGE}" \
  --build-arg BUILD_CMD="${BUILD_CMD}" \
  -f "${KUMA_DIR:-.}/tools/envoy/Dockerfile.build-ubuntu" "${SOURCE_DIR}"

# copy out the binary
id=$(docker create "${LOCAL_BUILD_IMAGE}")
docker cp "$id":/envoy-sources/bazel-bin/contrib/exe/envoy "${BINARY_PATH}"
docker rm -v "$id"
