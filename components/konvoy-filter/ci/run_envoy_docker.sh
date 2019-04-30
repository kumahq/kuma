#!/bin/bash

set -e

BASEDIR=$(dirname "$0")

KONVOY_REPO_DIR=$(git rev-parse --show-toplevel)
ENVOY_SRCDIR=/source/components/konvoy-filter

ENVOY_RUN_DOCKER_SCRIPT="${BASEDIR}"/../envoy/ci/run_envoy_docker.sh
PATCHED=".patched"

# patch the original `envoy/ci/run_envoy_docker.sh` script to support non-standard directory structure
function patch_run_envoy_docker() {
  sed -e 's/-v "$PWD":\/source /-v "$SOURCE_DIR":\/source -e ENVOY_SRCDIR /' \
      -e 's/^export TEST_TMPDIR=\/build\/tmp$/export TEST_TMPDIR="${BUILD_DIR}\/tmp"/' \
      "${ENVOY_RUN_DOCKER_SCRIPT}" > "${ENVOY_RUN_DOCKER_SCRIPT}${PATCHED}"
}

patch_run_envoy_docker

cd envoy/ && SOURCE_DIR=${KONVOY_REPO_DIR} ENVOY_SRCDIR=${ENVOY_SRCDIR} source ci/run_envoy_docker.sh${PATCHED} "cd \${ENVOY_SRCDIR} && ./ci/do_ci.sh $*"
