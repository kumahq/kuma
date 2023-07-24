#!/bin/bash

set -ex


DIR=`dirname $0`
echo $DIR

kumactl generate dataplane-token --name=grpc-client --valid-for 8760h > $DIR/kuma-token-grpc-client


kuma-dp run \
          --cp-address=https://localhost:5678/ \
          --dns-enabled=false \
          --envoy-log-level=debug \
          --dataplane-token-file=$DIR/kuma-token-grpc-client \
          --dataplane="
          type: Dataplane
          mesh: default
          name: grpc-client
          networking:
            address: 127.0.0.1
            outbound:
              - port: 8989
                tags:
                  kuma.io/service: grpc-server
                  kuma.io/protocol: http2
            inbound:
              - port: 10001
                tags:
                  kuma.io/service: grpc-client
                  kuma.io/protocol: http2
            admin:
              port: 9999
          "

