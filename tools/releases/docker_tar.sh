#!/usr/bin/env bash

# docker_tar pulls stable Kuma docker images from official bintray and make a tar of every image
# Potential improvement - auto upload to https://bintray.com/kong/kuma-misc (as a part of CI release job)

set -e

KUMA_COMPONENTS=("kuma-cp" "kuma-dp" "kuma-injector" "kuma-tcp-echo" "kumactl" "kuma-init" "kuma-prometheus-sd")

if [ "$#" -ne 1 ]; then
  echo "Usage: $0 <Kuma version>"
  exit 1
fi

KUMA_VERSION=$1

OUT_DIR="/tmp/kuma-$KUMA_VERSION"
mkdir -p $OUT_DIR

for component in "${KUMA_COMPONENTS[@]}"; do
  IMAGE="kong-docker-kuma-docker.bintray.io/$component:$KUMA_VERSION"
  OUT_FILE="$OUT_DIR/$component-$KUMA_VERSION.docker.tgz"
  docker pull $IMAGE
  docker save --output $OUT_FILE $IMAGE
  echo "Image $IMAGE saved to $OUT_FILE"
done


