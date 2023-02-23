#!/bin/sh

set -e

: "${KUBERNETES_RELEASE:?must be set!}"

build_docker_image(){
    docker build --build-arg ARCH="${1?required}" \
    --build-arg BASE_IMAGE="${2?required}" \
    --build-arg KUBERNETES_RELEASE="${3?required}" \
    --platform linux/"${1}" \
    --tag kumahq/kubectl:"${3}-${1}" .
    docker image push kumahq/kubectl:"${3}-${1}"
}

echo "Building docker image for amd64 and kubernetes ${KUBERNETES_RELEASE}"
build_docker_image amd64 amd64/alpine:latest ${KUBERNETES_RELEASE} .
echo "Building docker image for arm64 and kubernetes ${KUBERNETES_RELEASE}"
build_docker_image arm64 arm64v8/alpine:latest ${KUBERNETES_RELEASE} .
echo "Building docker image for arm and kubernetes ${KUBERNETES_RELEASE}"
build_docker_image arm arm32v7/alpine:latest ${KUBERNETES_RELEASE} .

docker manifest create kumahq/kubectl:$KUBERNETES_RELEASE \
  --amend kumahq/kubectl:${KUBERNETES_RELEASE}-amd64 \
  --amend kumahq/kubectl:${KUBERNETES_RELEASE}-arm64 \
  --amend kumahq/kubectl:${KUBERNETES_RELEASE}-arm 

docker manifest annotate kumahq/kubectl:$KUBERNETES_RELEASE kumahq/kubectl:$KUBERNETES_RELEASE-arm64 --os linux --arch arm64
docker manifest annotate kumahq/kubectl:$KUBERNETES_RELEASE kumahq/kubectl:$KUBERNETES_RELEASE-amd64 --os linux --arch amd64
docker manifest annotate kumahq/kubectl:$KUBERNETES_RELEASE kumahq/kubectl:$KUBERNETES_RELEASE-arm --os linux --arch arm

echo "Publishing manifest"
docker manifest push kumahq/kubectl:$KUBERNETES_RELEASE
