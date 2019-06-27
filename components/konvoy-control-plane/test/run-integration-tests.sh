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

rootDir=`pwd`
cd `dirname "$0"`
dockerComposeDir=`pwd`

docker-compose up --build --no-start
docker-compose up -d
trap "cd ${dockerComposeDir} && docker-compose down" EXIT

# wait for postgres
while ! nc -z localhost ${KONVOY_STORE_POSTGRES_PORT}; do sleep 1; done;
sleep 5;

# run tests
cd ${rootDir}
eval $1

cd ${dockerComposeDir}
docker-compose down