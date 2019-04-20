#!/usr/bin/env bash

#
# A wrapper script for 'test/run_envoy_bazel_coverage.sh'
# that substitutes command line tools with incompatible behaviour on MacOS.
#

set -ex # turn on command echo

BASEDIR=$(dirname "$0")

if [[ "`uname`" != "Darwin" ]]; then
  echo "This script is meant for MacOS only. Consider running 'test/run_envoy_bazel_coverage.sh' instead"
  exit 1
fi

echo "Looking for 'gcp' ..."
which gcp
exit_code="$?"
if [ "$exit_code" -ne 0 ]; then
    echo "'gcp' is missing. Consider installing it via 'brew install coreutils'"
fi

echo "Looking for 'pcregrep' ..."
which pcregrep
exit_code="$?"
if [ "$exit_code" -ne 0 ]; then
    echo "'pcregrep' is missing. Consider installing it via 'brew install pcre'"
fi

echo "Looking for 'gcovr' ..."
which gcovr
exit_code="$?"
if [ "$exit_code" -ne 0 ]; then
    echo "'gcovr' is missing. Consider installing it via 'brew install gcovr'"
fi

echo "Running test coverage build ..."
WORKSPACE=konvoy \
  TOOL_CP=gcp \
  TOOL_GREP=pcregrep\
  "${BASEDIR}"/../ci/do_mac_ci.sh coverage
