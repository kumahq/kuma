#!/bin/bash


# kumactl generate dataplane-token --name=http-server --valid-for 8760h > kuma-token-http-server



DIR=`dirname $0`
echo $DIR


# test-server http

DP_BIN=$DIR/../../build/artifacts-darwin-amd64/kuma-dp/kuma-dp
echo $DP_BIN

CMD=`$DP_BIN run \
          --cp-address=https://localhost:5678/ \
          --dns-enabled=false \
          --dataplane-token-file=$DIR/kuma-token-http-server \
          --dataplane-file=$DIR/dataplane-file-redis.yaml`

echo $CMD