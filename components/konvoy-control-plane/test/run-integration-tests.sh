#!/bin/bash

set -eE

function get_unused_port() {
  for port in $(seq 10000 65000);
  do
    nc -z localhost ${port} > /dev/null 2>&1;
    [[ $? -eq 1 ]] && echo "$port" && break;
  done
}

export KONVOY_STORE_POSTGRES_HOST=localhost
export KONVOY_STORE_POSTGRES_PORT=$(get_unused_port)
export KONVOY_STORE_POSTGRES_USER=konvoy
export KONVOY_STORE_POSTGRES_PASSWORD=konvoy
export KONVOY_STORE_POSTGRES_DB_NAME=konvoy

dockerCompose="$(dirname "$0")/docker-compose.yaml"

docker-compose -f ${dockerCompose} up --build --no-start
docker-compose -f ${dockerCompose} up -d
trap "docker-compose -f ${dockerCompose} down" EXIT

# wait for postgres
while ! nc -z localhost ${KONVOY_STORE_POSTGRES_PORT}; do sleep 1; done;
sleep 5;

# run tests
eval $1

docker-compose -f ${dockerCompose} down