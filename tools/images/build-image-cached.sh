#!/usr/bin/env bash

APP="$1"
DOCKER_REGISTRY="$2"
GOARCH="$3"
DOCKER_BUILD_ARGS="$4"
DOCKERFILE_PATH="$5"

function usage {
  echo "Usage: $0 app docker_registry goarch docker_build_args dockerfile_path"
  echo "app - name of the app, eg: kuma-cp"
  echo "docker_registry - eg: kumahq"
  echo "goarch - eg: arm64"
  echo "docker_build_args - eg: --build-arg ENVOY_VERSION=1.22.0"
  echo "dockerfile_path - eg: tools/releases/dockerfiles/Dockerfile.kuma-cp"
  exit 0
}

if [ "$#" -ne 5 ]; then
    echo "Wrong number of parameters"
    usage
    exit 1
fi

function image_tag_by_content {
  local NAME=${1}
  local TAG=${2}
  local ARCH=${3}

  if [ -z "$IMAGE_ARCH_TAG_ENABLED" ]; then
    echo "$DOCKER_REGISTRY"/"$NAME":"$TAG"
  else
    echo "$DOCKER_REGISTRY"/"$NAME":"$TAG"-"$ARCH"
  fi
}

KUMA_DOCKER_IMAGE_CONTENTS_TAG=$(tools/images/content-version.sh "$DOCKERFILE_PATH".dockerignore)
KUMA_DOCKER_IMAGE_CONTENTS=$(image_tag_by_content "$APP" "$KUMA_DOCKER_IMAGE_CONTENTS_TAG" "$GOARCH")

echo "checking $KUMA_DOCKER_IMAGE_CONTENTS"
if docker image history "$KUMA_DOCKER_IMAGE_CONTENTS" > /dev/null; then
  echo "already built - skipping rebuilding"
  exit 0
fi

docker build -t "$KUMA_DOCKER_IMAGE" -t "$KUMA_DOCKER_IMAGE_CONTENTS" "$DOCKER_BUILD_ARGS" --build-arg ARCH="$GOARCH" --build-arg BASE_IMAGE_ARCH="$GOARCH" -f "$DOCKERFILE_PATH" .
