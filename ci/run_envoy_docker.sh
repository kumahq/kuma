#!/bin/bash

set -e

KONVOY_DIR=`pwd`

cd envoy/ && PWD=$KONVOY_DIR source ci/run_envoy_docker.sh "./ci/do_ci.sh $*"
