#!/bin/bash

set -e

KONVOY_DIR=`pwd`

cd envoy/ && PWD=$KONVOY_DIR source ci/run_envoy_docker.sh '/bin/bash' && cd = || cd -  
