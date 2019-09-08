#!/bin/bash

set -eE

function get_unused_port() {
  for port in $(seq 10000 65000);
  do
    nc -z localhost ${port} > /dev/null 2>&1;
    [[ $? -eq 1 ]] && echo "$port" && break;
  done
}

export KUMA_STORE_POSTGRES_HOST=localhost
export KUMA_STORE_POSTGRES_PORT=$(get_unused_port)
export KUMA_STORE_POSTGRES_USER=konvoy
export KUMA_STORE_POSTGRES_PASSWORD=konvoy
export KUMA_STORE_POSTGRES_DB_NAME=konvoy

dockerCompose="$(dirname "$0")/../postgres/docker-compose.yaml"

docker-compose -f ${dockerCompose} up --build --no-start
docker-compose -f ${dockerCompose} up -d
trap "docker-compose -f ${dockerCompose} down" EXIT

# wait for postgres
$(dirname "$0")/../postgres/wait-for-postgres.sh ${KUMA_STORE_POSTGRES_PORT}

# run tests
eval $1

docker-compose -f ${dockerCompose} down