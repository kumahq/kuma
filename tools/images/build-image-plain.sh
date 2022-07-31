#!/usr/bin/env bash

IMAGE_NAME="$1"
# shellcheck disable=SC2034
DOCKER_REGISTRY="$2" # unused - only kept to have interface compatibility between two build scripts
GOARCH="$3"
DOCKER_BUILD_ARGS="$4"
DOCKERFILE_PATH="$5"

# if $DOCKER_BUILD_ARGS is quoted and empty then docker complains "build expects only one argument"
# shellcheck disable=SC2086
docker build -t "$IMAGE_NAME" $DOCKER_BUILD_ARGS --build-arg ARCH="$GOARCH" --build-arg BASE_IMAGE_ARCH="$GOARCH" -f "$DOCKERFILE_PATH" .
