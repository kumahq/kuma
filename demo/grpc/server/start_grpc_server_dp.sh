#!/bin/bash

set -ex


DIR=`dirname $0`
echo $DIR

kumactl generate dataplane-token --name=grpc-server --valid-for 8760h > $DIR/kuma-token-grpc-server


PORT=8888
SERVER_PORT=`expr $PORT + 1`
ADMIN_PORT=`expr $SERVER_PORT + 1`

#nohup $DIR/../../build/artifacts-darwin-amd64/test-server/test-server grpc server --ip 10.53.36.94 --port 2345


kuma-dp run \
          --cp-address=https://localhost:5678/ \
          --dns-enabled=false \
          --dataplane-token-file=$DIR/kuma-token-grpc-server \
          --dataplane="
          type: Dataplane
          mesh: default
          name: grpc-server
          networking:
            address: 127.0.0.1
            inbound:
              - port: 8888
                servicePort: 2345
                serviceAddress: 10.53.36.94
                tags:
                  kuma.io/service: grpc-server
                  kuma.io/protocol: http2
            admin:
              port: $ADMIN_PORT
          "

