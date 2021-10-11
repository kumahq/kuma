#!/usr/bin/env bash

set -e

GOARCH=(amd64)

# first component is the distribution name, second is the system - must map to
# valid $GOOS values
DISTRIBUTIONS=(debian:linux:alpine ubuntu:linux:alpine rhel:linux:centos centos:linux:centos darwin:darwin:darwin)

PULP_HOST="https://api.pulp.konnect-prod.konghq.com"
PULP_PACKAGE_TYPE="mesh"
PULP_DIST_NAME="alpine"
[ -z "$RELEASE_NAME" ] && RELEASE_NAME="kuma"
ENVOY_VERSION=1.20.0
[ -z "$KUMA_CONFIG_PATH" ] && KUMA_CONFIG_PATH=pkg/config/app/kuma-cp/kuma-cp.defaults.yaml

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
  msg_red $@
  exit 1
}

function get_envoy() {
  local distro=$1
  local envoy_distro=$2

  local status=$(curl -L -o build/envoy-$distro \
    --write-out %{http_code} --silent --output /dev/null \
    "https://download.konghq.com/mesh-alpine/envoy-$ENVOY_VERSION-$envoy_distro")

  [ "$status" -ne "200" ] && msg_err "Error: failed downloading Envoy" || true
}

function create_tarball() {
  local system=$1
  local arch=$2
  local distro=$3
  local envoy_distro=$4

  local dest_dir=build/$RELEASE_NAME-$distro-$arch
  local kuma_dir=$dest_dir/$RELEASE_NAME-$KUMA_VERSION

  rm -rf $dest_dir
  mkdir $dest_dir
  mkdir $kuma_dir
  mkdir $kuma_dir/bin
  mkdir $kuma_dir/conf

  get_envoy $distro $envoy_distro
  chmod 755 build/envoy-$distro

  cp -p build/envoy-$distro $kuma_dir/bin/envoy
  cp -p build/artifacts-$system-$arch/kuma-cp/kuma-cp $kuma_dir/bin
  cp -p build/artifacts-$system-$arch/kuma-dp/kuma-dp $kuma_dir/bin
  cp -p build/artifacts-$system-$arch/kumactl/kumactl $kuma_dir/bin
  cp -p build/artifacts-$system-$arch/coredns/coredns $kuma_dir/bin
  cp -p build/artifacts-$system-$arch/kuma-prometheus-sd/kuma-prometheus-sd $kuma_dir/bin
  cp -p $KUMA_CONFIG_PATH $kuma_dir/conf/kuma-cp.conf.yml

  cp tools/releases/templates/* $kuma_dir

  tar -czf build/artifacts-$system-$arch/$RELEASE_NAME-$KUMA_VERSION-$distro-$arch.tar.gz -C $dest_dir .
}

function package() {
  for os in "${DISTRIBUTIONS[@]}"; do
    local distro=$(echo "$os" | awk '{split($0,parts,":"); print parts[1]}')
    local system=$(echo "$os" | awk '{split($0,parts,":"); print parts[2]}')
    local envoy_distro=$(echo "$os" | awk '{split($0,parts,":"); print parts[3]}')

    for arch in "${GOARCH[@]}"; do

      msg ">>> Packaging Kuma for $distro ($system-$arch)..."
      msg

      make GOOS=$system GOARCH=$arch BUILD_INFO_GIT_TAG=$KUMA_VERSION BUILD_INFO_GIT_COMMIT=$KUMA_COMMIT build
      create_tarball $system $arch $distro $envoy_distro

      msg
      msg_green "... success!"
      msg
    done
  done
}

function release() {
  for os in "${DISTRIBUTIONS[@]}"; do
    local distro=$(echo "$os" | awk '{split($0,parts,":"); print parts[1]}')
    local system=$(echo "$os" | awk '{split($0,parts,":"); print parts[2]}')

    for arch in "${GOARCH[@]}"; do
      local artifact="build/artifacts-$system-$arch/$RELEASE_NAME-$KUMA_VERSION-$distro-$arch.tar.gz"
      [ ! -f "$artifact" ] && msg_yellow "Package '$artifact' not found, skipping..." && continue

      msg_green "Releasing Kuma for '$os', '$arch'..."

      docker run --rm \
        -e PULP_USERNAME=${PULP_USERNAME} -e PULP_PASSWORD=${PULP_PASSWORD} \
        -e PULP_HOST=${PULP_HOST} \
        -v "${PWD}":/files:ro -it kong/release-script \
        --file /files/"$artifact" \
        --package-type ${PULP_PACKAGE_TYPE} --dist-name ${PULP_DIST_NAME} --publish
    done
  done
}

function usage() {
  echo "Usage: $0 [--package|--release]"
  exit 0
}

function main() {
  while [[ $# -gt 0 ]]; do
    flag=$1
    case $flag in
    --help)
      usage
      ;;
    --package)
      op="package"
      ;;
    --release)
      op="release"
      ;;
    --version)
      KUMA_VERSION=$2
      shift
      ;;
    --sha)
      KUMA_COMMIT=$2
      shift
      ;;
    *)
      usage
      break
      ;;
    esac
    shift
  done

  [ -z "$PULP_USERNAME" ] && msg_err "PULP_USERNAME required"
  [ -z "$PULP_PASSWORD" ] && msg_err "PULP_PASSWORD required"
  [ -z "$KUMA_VERSION" ] && msg_err "Error: --version required"

  case $op in
  package)
    package
    ;;
  release)
    release
    ;;
  esac
}

main $@
