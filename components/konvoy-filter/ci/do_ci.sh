#!/bin/bash

set -e

BASEDIR=$(dirname "$0")

ENVOY_BUILD_SETUP_SCRIPT="${BASEDIR}"/../envoy/ci/build_setup.sh
ENVOY_BUILD_SETUP_SCRIPT_PATCHED="${ENVOY_BUILD_SETUP_SCRIPT}.patched"

# patch the original `envoy/ci/build_setup.sh` script to support builds on MacOS
function patch_build_setup () {
  sed -e 's/^export BUILD_DIR=\/build$/[ -z "${BUILD_DIR}" ] \&\& export BUILD_DIR=\/build/' \
      -e 's/^export TEST_TMPDIR=\/build\/tmp$/export TEST_TMPDIR="${BUILD_DIR}\/tmp"/' \
      "${ENVOY_BUILD_SETUP_SCRIPT}" > "${ENVOY_BUILD_SETUP_SCRIPT_PATCHED}"
}

patch_build_setup

# setup common variables
. "${ENVOY_BUILD_SETUP_SCRIPT_PATCHED}" "-nofetch"

echo "building using ${NUM_CPUS} CPUs"

BAZEL_BUILD_TARGET="//:konvoy"
BAZEL_BUILD_OUTPUT_PATH="konvoy"
DELIVERY_PATH="konvoy"

function do_build () {
  echo "bazel release build..."

  bazel build ${BAZEL_BUILD_OPTIONS} -c opt ${BAZEL_BUILD_TARGET}

  # Copy the envoy-static binary somewhere that we can access outside of the
  # container.
  cp -f \
    "${ENVOY_SRCDIR}"/bazel-bin/"${BAZEL_BUILD_OUTPUT_PATH}" \
    "${ENVOY_DELIVERY_DIR}"/"${DELIVERY_PATH}"

  echo "Copying release binary for image build..."
  mkdir -p "${ENVOY_SRCDIR}"/_build/build_release
  cp -f "${ENVOY_DELIVERY_DIR}"/"${DELIVERY_PATH}" "${ENVOY_SRCDIR}"/_build/build_release
  mkdir -p "${ENVOY_SRCDIR}"/_build/build_release_stripped
  strip "${ENVOY_DELIVERY_DIR}"/"${DELIVERY_PATH}" -o "${ENVOY_SRCDIR}"/_build/build_release_stripped/"${DELIVERY_PATH}"
}

function do_test() {
    echo "bazel test..."

    bazel build ${BAZEL_BUILD_OPTIONS} -c opt //source/... //test/...

    bazel test ${BAZEL_TEST_OPTIONS} -c opt //test/...
}

function do_coverage() {
  echo "bazel coverage build with tests..."

  # gcovr is a pain to run with `bazel run`, so package it up into a
  # relocatable and hermetic-ish .par file.
  bazel build @com_github_gcovr_gcovr//:gcovr.par
  export GCOVR="/tmp/gcovr.par"
  cp -f "${ENVOY_SRCDIR}/bazel-bin/external/com_github_gcovr_gcovr/gcovr.par" ${GCOVR}

  "$(dirname "$0")"/../test/run_envoy_bazel_coverage.sh
}

function do_shell() {
  echo "launching interactive shell..."
  bash
}

if [[ $(uname -s) == Linux ]]; then
  setup_gcc_toolchain
elif [[ $(uname -s) == Darwin ]]; then
  setup_clang_toolchain
fi

case "$1" in
  build)
    do_build
  ;;
  test)
    do_test
  ;;
  coverage)
    do_coverage
  ;;
  shell)
    do_shell
  ;;
  *)
    echo "must be one of [build,test,coverage,shell]"
    exit 1
  ;;
esac
