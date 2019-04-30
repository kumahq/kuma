#!/bin/bash

set -e

BASEDIR=$(dirname "$0")

# environment variables set by TravisCI
TRAVIS_ENV=$(printenv | awk -F= '{print $1}' | grep -e '^TRAVIS_' -e '^CI$' -e '^CONTINUOUS_INTEGRATION$' | sort | awk '{print "-e "$0}' | tr '\n' ' ')

# manually set variables must be propagated explicitly
DOCKER_ENV="${TRAVIS_ENV} -e BAZEL_EXTRA_TEST_OPTIONS"

# abuse 'http_proxy' environment variable to pass extra arguments into envoy/ci/run_envoy_docker.sh
export http_proxy=" ${DOCKER_ENV}"

"${BASEDIR}"/run_envoy_docker.sh $*
