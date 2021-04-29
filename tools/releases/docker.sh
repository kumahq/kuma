#!/usr/bin/env bash

set -e

[ -z "$KUMA_DOCKER_REPO" ] && KUMA_DOCKER_REPO="docker.io"
KUMA_DOCKER_REPO_ORG=${KUMA_DOCKER_REPO}/kumahq
KUMA_COMPONENTS=("kuma-cp" "kuma-dp" "kumactl" "kuma-init" "kuma-prometheus-sd")

function msg_green() {
  builtin echo -en "\033[1;32m"
  echo "$@"
  builtin echo -en "\033[0m"
}

function msg_red() {
  builtin echo -en "\033[1;31m" >&2
  echo "$@" >&2
  builtin echo -en "\033[0m" >&2
}

function msg_yellow() {
  builtin echo -en "\033[1;33m"
  echo "$@"
  builtin echo -en "\033[0m"
}

function msg() {
  builtin echo -en "\033[1m"
  echo "$@"
  builtin echo -en "\033[0m"
}

function msg_err() {
  msg_red "$@"
  exit 1
}

function build() {
  for component in "${KUMA_COMPONENTS[@]}"; do
    msg "Building $component..."
    docker build --build-arg KUMA_ROOT=$(pwd) -t $KUMA_DOCKER_REPO_ORG/$component:$KUMA_VERSION \
      -f tools/releases/dockerfiles/Dockerfile.$component .
    docker tag $KUMA_DOCKER_REPO_ORG/$component:$KUMA_VERSION $KUMA_DOCKER_REPO_ORG/$component:latest
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

  for component in "${KUMA_COMPONENTS[@]}"; do
    msg "Pushing $component:$KUMA_VERSION ..."
    docker push $KUMA_DOCKER_REPO_ORG/$component:$KUMA_VERSION
    re='^[0-9]+([.][0-9]+)+([.][0-9]+)$'
    if [[ $yournumber =~ $re ]]; then
      msg "Pushing $component:latest ..."
      docker push $KUMA_DOCKER_REPO_ORG/$component:latest
    fi
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

  [ -z "$DOCKER_USERNAME" ] && msg_err "$DOCKER_USERNAME required"
  [ -z "$DOCKER_API_KEY" ] && msg_err "$DOCKER_API_KEY required"
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

main $@
