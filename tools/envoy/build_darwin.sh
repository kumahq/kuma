#!/bin/bash

set -e

echo "Building Envoy for Darwin"

BINARY_PATH=${BINARY_PATH:-"out/envoy"}
BINARY_DIR=$(dirname ${BINARY_PATH})
mkdir -p "${BINARY_DIR}"

SOURCE_DIR=${SOURCE_DIR:-"out/envoy-sources"}
SOURCE_DIR="${SOURCE_DIR}" ./tools/envoy/fetch_sources.sh

pushd "${SOURCE_DIR}"

BAZEL_BUILD_EXTRA_OPTIONS=${BAZEL_BUILD_EXTRA_OPTIONS:-""}
read -ra BAZEL_BUILD_EXTRA_OPTIONS <<< "${BAZEL_BUILD_EXTRA_OPTIONS}"
BAZEL_BUILD_OPTIONS=(
    "--curses=no"
    --show_task_finish
    --verbose_failures
    "--action_env=PATH=/usr/local/bin:/opt/local/bin:/usr/bin:/bin"
    "--define" "wasm=disabled"
    "${BAZEL_BUILD_EXTRA_OPTIONS[@]}")
bazel build "${BAZEL_BUILD_OPTIONS[@]}" -c opt //source/exe:envoy-static

popd

cp ${SOURCE_DIR}/bazel-bin/source/exe/envoy-static ${BINARY_PATH}

