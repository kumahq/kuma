#!/bin/bash

set -e

BASEDIR=$(dirname "$0")

[ -z "${NUM_CPUS}" ] && export NUM_CPUS=`sysctl -n hw.ncpu`
[ -z "${ENVOY_SRCDIR}" ] && export ENVOY_SRCDIR="${BASEDIR}/.."

# overwrite https://github.com/envoyproxy/envoy/blob/v1.10.0/ci/build_setup.sh#L67
export BAZEL_BUILD_EXTRA_OPTIONS="${BAZEL_BUILD_EXTRA_OPTIONS} --linkopt=-fuse-ld="

# code coverage
[ -z "${TOOL_CP}" ] && export TOOL_CP=gcp
[ -z "${TOOL_GREP}" ] && export TOOL_GREP=pcregrep

. "${BASEDIR}/do_ci.sh" $*
