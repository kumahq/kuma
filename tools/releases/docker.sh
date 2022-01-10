#!/usr/bin/env bash

set -e

source "$(dirname -- "${BASH_SOURCE[0]}")/../common.sh"

[ -z "$KUMA_DOCKER_REPO" ] && KUMA_DOCKER_REPO="docker.io"
[ -z "$KUMA_DOCKER_REPO_ORG" ] && KUMA_DOCKER_REPO_ORG=${KUMA_DOCKER_REPO}/kumahq
[ -z "$KUMA_COMPONENTS" ] && KUMA_COMPONENTS="kuma-cp kuma-dp kumactl kuma-init kuma-prometheus-sd"

function build() {
  for component in ${KUMA_COMPONENTS}; do
    msg "Building $component..."
    docker build --build-arg KUMA_ROOT="$(pwd)" -t $KUMA_DOCKER_REPO_ORG/"$component":"$KUMA_VERSION" \
      -f tools/releases/dockerfiles/Dockerfile."$component" .
    docker tag $KUMA_DOCKER_REPO_ORG/"$component":"$KUMA_VERSION" $KUMA_DOCKER_REPO_ORG/"$component":latest
    msg_green "... done!"
  done
}

function docker_login() {
  docker login -u "$DOCKER_USERNAME" -p "$DOCKER_API_KEY" $KUMA_DOCKER_REPO
}

function docker_logout() {
  docker logout $KUMA_DOCKER_REPO
}

function push() {
  docker_login

  for component in ${KUMA_COMPONENTS}; do
    msg "Pushing $component:$KUMA_VERSION ..."
    docker push $KUMA_DOCKER_REPO_ORG/"$component":"$KUMA_VERSION"
    msg_green "... done!"
  done

  docker_logout
}

function usage() {
  echo "Usage: $0 [--build | --push ] --version <Kuma version>"
  exit 0
}

function main() {
  while [[ $# -gt 0 ]]; do
    flag=$1
    case $flag in
    --help)
      usage
      ;;
    --build)
      op="build"
      ;;
    --push)
      op="push"
      ;;
    --version)
      KUMA_VERSION=$2
      shift
      ;;
    *)
      usage
      break
      ;;
    esac
    shift
  done

  [ -z "$DOCKER_USERNAME" ] && msg_err "\$DOCKER_USERNAME required"
  [ -z "$DOCKER_API_KEY" ] && msg_err "\$DOCKER_API_KEY required"
  [ -z "$KUMA_VERSION" ] && msg_err "Error: --version required"

  case $op in
  build)
    build
    ;;
  push)
    push
    ;;
  esac
}

main "$@"
