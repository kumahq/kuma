#!/usr/bin/env bash

set -e

SCRIPT_DIR="$(dirname -- "${BASH_SOURCE[0]}")"
source "${SCRIPT_DIR}/../common.sh"

KUMA_DOCKER_REPO="${KUMA_DOCKER_REPO:-docker.io}"
KUMA_DOCKER_REPO_ORG="${KUMA_DOCKER_REPO_ORG:-${KUMA_DOCKER_REPO}/kumahq}"
KUMA_COMPONENTS="${KUMA_COMPONENTS:-kuma-cp kuma-dp kumactl kuma-init kuma-cni}"
BUILD_INFO=$("${SCRIPT_DIR}/../releases/version.sh")
ENVOY_VERSION=$(echo "$BUILD_INFO" | cut -d " " -f 5)
KUMA_VERSION=$(echo "$BUILD_INFO" | cut -d " " -f 1)
BUILD_ARCH="${BUILD_ARCH:-amd64 arm64}"

function build() {
  for component in ${KUMA_COMPONENTS}; do
    for arch in ${BUILD_ARCH}; do
      msg "Building $component..."
      build_args=(
        --build-arg ARCH="${arch}"
        --build-arg ENVOY_VERSION="${ENVOY_VERSION}"
        --build-arg BASE_IMAGE_ARCH="${arch}"
      )
      additional_args=()
      if [ "$arch" == "arm64" ]; then
        read -ra additional_args <<< "${ARM64_BUILD_ARGS[@]}"
      fi
      docker build --pull "${build_args[@]}" "${additional_args[@]}" -t "${KUMA_DOCKER_REPO_ORG}/${component}:${KUMA_VERSION}-${arch}" \
        -f tools/releases/dockerfiles/Dockerfile."${component}" .
      docker tag "${KUMA_DOCKER_REPO_ORG}/${component}:${KUMA_VERSION}-${arch}" "${KUMA_DOCKER_REPO_ORG}/${component}:latest-${arch}"
      msg_green "... done!"
    done
  done
}

function docker_login() {
  [ -z "$DOCKER_USERNAME" ] && msg_err "\$DOCKER_USERNAME required"
  [ -z "$DOCKER_API_KEY" ] && msg_err "\$DOCKER_API_KEY required"
  docker login -u "$DOCKER_USERNAME" -p "$DOCKER_API_KEY" "$KUMA_DOCKER_REPO"
}

function docker_logout() {
  docker logout "$KUMA_DOCKER_REPO"
}

function push() {
  docker_login

  for component in ${KUMA_COMPONENTS}; do
    for arch in ${BUILD_ARCH}; do
      msg "Pushing $component:$KUMA_VERSION-$arch ..."
      docker push "${KUMA_DOCKER_REPO_ORG}/${component}:${KUMA_VERSION}-${arch}"
      msg_green "... done!"
    done
  done

  docker_logout
}

# allows to push many arch types as one tag
function manifest() {
  docker_login

  for component in ${KUMA_COMPONENTS}; do
    images=()
    for arch in ${BUILD_ARCH}; do
      images+=("--amend ${KUMA_DOCKER_REPO_ORG}/${component}:${KUMA_VERSION}-${arch}")
    done
    command="docker manifest create ${KUMA_DOCKER_REPO_ORG}/${component}:${KUMA_VERSION} ${images[*]}"
    msg "Creating manifest for ${KUMA_DOCKER_REPO_ORG}/${component}:${KUMA_VERSION}..."
    eval "$command"
    msg "Pushing manifest ${KUMA_DOCKER_REPO_ORG}/${component}:${KUMA_VERSION} ..."
    docker manifest push "${KUMA_DOCKER_REPO_ORG}/${component}:${KUMA_VERSION}"
    msg ".. done!"
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
    --manifest)
      op="manifest"
      ;;
    *)
      usage
      break
      ;;
    esac
    shift
  done

  case $op in
  build)
    build
    ;;
  push)
    push
    ;;
  manifest)
    manifest
    ;;
  esac
}

main "$@"
