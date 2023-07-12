#!/bin/bash

# kumactl generate dataplane-token --name=redis --valid-for 8760h > kuma-token-redis

# start redis-server brew services start redis

DIR=`dirname $0`
echo $DIR

DP_BIN=$DIR/../../build/artifacts-darwin-amd64/kuma-dp/kuma-dp
echo $DP_BIN

CMD=`$DP_BIN run \
          --cp-address=https://localhost:5678/ \
          --dns-enabled=false \
          --dataplane-token-file=$DIR/kuma-token-redis \
          --dataplane-file=$DIR/dataplane-file-redis.yaml`

echo $CMD
