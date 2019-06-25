#!/bin/bash

set -eE

function get_unused_port() {
  for port in $(seq 10000 65000);
  do
    echo -ne "\035" | telnet 127.0.0.1 $port > /dev/null 2>&1;
    [[ $? -eq 1 ]] && echo "$port" && break;
  done
}

export KONVOY_STORE_POSTGRES_HOST=localhost
export KONVOY_STORE_POSTGRES_PORT=$(get_unused_port)
export KONVOY_STORE_POSTGRES_USER=konvoy
export KONVOY_STORE_POSTGRES_PASSWORD=konvoy
export KONVOY_STORE_POSTGRES_DB_NAME="konvoy_$RANDOM"

cd `dirname "$0"`

docker-compose up --build --no-start
docker-compose up -d
trap "docker-compose down" EXIT

# wait for postgres
while ! nc -z localhost ${KONVOY_STORE_POSTGRES_PORT}; do sleep 1; done;
sleep 5;

go test ../... -tags=integration -count=1

docker-compose down