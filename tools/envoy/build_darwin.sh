#!/bin/bash

set -o errexit
set -o pipefail
set -o nounset

echo "Building Envoy for Darwin"

mkdir -p "$(dirname "${BINARY_PATH}")"

SOURCE_DIR="${SOURCE_DIR}" "${KUMA_DIR:-.}/tools/envoy/fetch_sources.sh"
CONTRIB_ENABLED_MATRIX_SCRIPT=$(realpath "${KUMA_DIR:-.}/tools/envoy/contrib_enabled_matrix.py")

pushd "${SOURCE_DIR}"

BAZEL_BUILD_EXTRA_OPTIONS=${BAZEL_BUILD_EXTRA_OPTIONS:-""}
read -ra BAZEL_BUILD_EXTRA_OPTIONS <<< "${BAZEL_BUILD_EXTRA_OPTIONS}"
BAZEL_BUILD_OPTIONS=(
    "--curses=no"
    --show_task_finish
    --verbose_failures
    --//contrib/vcl/source:enabled=false
    "--action_env=PATH=/usr/local/bin:/opt/local/bin:/usr/bin:/bin:/opt/homebrew/bin"
    "--define" "wasm=disabled"
    "${BAZEL_BUILD_EXTRA_OPTIONS[@]+"${BAZEL_BUILD_EXTRA_OPTIONS[@]}"}")

CONTRIB_ENABLED_ARGS=$(python "${CONTRIB_ENABLED_MATRIX_SCRIPT}")

# shellcheck disable=SC2086
bazel build "${BAZEL_BUILD_OPTIONS[@]}" -c opt //contrib/exe:envoy-static ${CONTRIB_ENABLED_ARGS}

popd

cp "${SOURCE_DIR}"/bazel-bin/contrib/exe/envoy-static "${BINARY_PATH}"

