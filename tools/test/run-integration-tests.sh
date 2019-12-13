#!/bin/bash

set -eE

function get_unused_port() {
  for port in $(seq 10000 65000);
  do
    nc -z localhost ${port} > /dev/null 2>&1;
    [[ $? -eq 1 ]] && echo "$port" && break;
  done
}

# Takes a path argument and returns it as an absolute path.
function to_abs_path() {
  echo "$(cd "$(dirname "$1")"; pwd)/$(basename "$1")"
}

dockerComposeSslDir="$(to_abs_path $(dirname $0)/../postgres/ssl)"

chmod 600 $dockerComposeSslDir/certs/postgres.client.key

export KUMA_STORE_POSTGRES_HOST=localhost
export KUMA_STORE_POSTGRES_PORT=$(get_unused_port)
export KUMA_STORE_POSTGRES_USER=kuma
export KUMA_STORE_POSTGRES_PASSWORD=kuma
export KUMA_STORE_POSTGRES_DB_NAME=kuma
export KUMA_STORE_POSTGRES_TLS_MODE=verifyCa
export KUMA_STORE_POSTGRES_TLS_CERT_PATH=$dockerComposeSslDir/certs/postgres.client.crt
export KUMA_STORE_POSTGRES_TLS_KEY_PATH=$dockerComposeSslDir/certs/postgres.client.key
export KUMA_STORE_POSTGRES_TLS_CA_PATH=$dockerComposeSslDir/certs/rootCA.crt

dockerCompose="$dockerComposeSslDir/docker-compose.yaml"

docker-compose -f ${dockerCompose} up --build --no-start
docker-compose -f ${dockerCompose} up -d
trap "docker-compose -f ${dockerCompose} down" EXIT

# wait for postgres
$(dirname "$0")/../postgres/wait-for-postgres.sh ${KUMA_STORE_POSTGRES_PORT}

# run tests
eval $1

docker-compose -f ${dockerCompose} down