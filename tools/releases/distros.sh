#!/usr/bin/env bash

set -e

SCRIPT_DIR="$(dirname -- "${BASH_SOURCE[0]}")"
source "${SCRIPT_DIR}/../common.sh"

# first component is the system - must map to valid $GOOS values
# if present, second is the distribution and third is the envoy distribution
# without a distribution we package only kumactl as a static binary
DISTRIBUTIONS=(
  linux:debian:linux:amd64
  linux:ubuntu:linux:amd64
  linux:rhel:linux:amd64
  linux:centos:linux:amd64
  darwin:darwin:darwin:amd64
  linux:::amd64
  linux:debian:linux:arm64
  linux:ubuntu:linux:arm64
  darwin:darwin:darwin:arm64
  linux:::arm64
)

PULP_HOST="https://api.pulp.konnect-prod.konghq.com"
PULP_PACKAGE_TYPE="mesh"
PULP_DIST_NAME="alpine"
[ -z "$RELEASE_NAME" ] && RELEASE_NAME="kuma"
BUILD_INFO=$("${SCRIPT_DIR}/../releases/version.sh")
KUMA_VERSION=$(echo "$BUILD_INFO" | cut -d " " -f 1)
[ -z "$KUMA_CONFIG_PATH" ] && KUMA_CONFIG_PATH=pkg/config/app/kuma-cp/kuma-cp.defaults.yaml
CTL_NAME="kumactl"
[ -z "$EBPF_PROGRAMS_IMAGE" ] && EBPF_PROGRAMS_IMAGE="kumahq/kuma-net-ebpf:0.8.6"

function get_ebpf_programs() {
  local arch=$1
  local system=$2
  local kuma_dir=$3
  local container

  if [[ "$system" != "linux" ]]; then
    return
  fi

  if [[ "$arch" != "amd64" ]] && [[ "$arch" != "arm64" ]]; then
    return
  fi

  container=$(DOCKER_DEFAULT_PLATFORM=$system/$arch docker create "$EBPF_PROGRAMS_IMAGE" copy)
  docker cp "$container:/ebpf" "$kuma_dir"
  docker rm -v "$container"
}

# create_kumactl_tarball packages only kumactl
function create_kumactl_tarball() {
  local arch=$1
  local system=$2

  msg ">>> Packaging ${RELEASE_NAME} static kumactl binary for $system-$arch..."
  msg

  make GOOS="$system" GOARCH="$arch" build/kumactl

  local dest_dir=build/$RELEASE_NAME-$arch
  local kuma_subdir="$RELEASE_NAME-$KUMA_VERSION"
  local kuma_dir=$dest_dir/$kuma_subdir

  rm -rf "$dest_dir"
  mkdir "$dest_dir"
  mkdir "$kuma_dir"
  mkdir "$kuma_dir/bin"

  artifact_dir=$(artifact_dir "$arch" "$system")
  cp -p "$artifact_dir/kumactl/kumactl" "$kuma_dir/bin"

  cp tools/releases/templates/LICENSE "$kuma_dir"
  cp tools/releases/templates/NOTICE-kumactl "$kuma_dir"

  archive_path=$(archive_path "$arch" "$system")

  tar -czf "${archive_path}" -C "$dest_dir" "$kuma_subdir"
}

function create_tarball() {
  local arch=$1
  local system=$2
  local distro=$3
  local envoy_distro=$4

  msg ">>> Packaging ${RELEASE_NAME} for $distro ($system-$arch)..."
  msg

  make GOOS="$system" GOARCH="$arch" ENVOY_DISTRO="$envoy_distro" build

  local dest_dir=build/$RELEASE_NAME-$distro-$arch
  local kuma_subdir="$RELEASE_NAME-$KUMA_VERSION"
  local kuma_dir=$dest_dir/$kuma_subdir

  rm -rf "$dest_dir"
  mkdir "$dest_dir"
  mkdir "$kuma_dir"
  mkdir "$kuma_dir/bin"
  mkdir "$kuma_dir/conf"

  get_ebpf_programs "$arch" "$system" "$kuma_dir"

  artifact_dir=$(artifact_dir "$arch" "$system")
  cp -p "$artifact_dir/envoy/envoy" "$kuma_dir/bin"
  cp -p "$artifact_dir/kuma-cp/kuma-cp" "$kuma_dir/bin"
  cp -p "$artifact_dir/kuma-dp/kuma-dp" "$kuma_dir/bin"
  cp -p "$artifact_dir/kumactl/kumactl" "$kuma_dir/bin"
  cp -p "$artifact_dir/coredns/coredns" "$kuma_dir/bin"
  cp -p "$KUMA_CONFIG_PATH" "$kuma_dir/conf/kuma-cp.conf.yml"

  cp tools/releases/templates/* "$kuma_dir"

  archive_path=$(archive_path "$arch" "$system" "$distro")

  tar -czf "${archive_path}" -C "$dest_dir" "$kuma_subdir"
}

function package() {
  for os in "${DISTRIBUTIONS[@]}"; do
    local system
    system=$(echo "$os" | awk '{split($0,parts,":"); print parts[1]}')
    local distro
    distro=$(echo "$os" | awk '{split($0,parts,":"); print parts[2]}')
    local envoy_distro
    envoy_distro=$(echo "$os" | awk '{split($0,parts,":"); print parts[3]}')
    local arch
    arch=$(echo "$os" | awk '{split($0,parts,":"); print parts[4]}')

    if [[ -n $distro ]]; then
      create_tarball "$arch" "$system" "$distro" "$envoy_distro"
    else
      create_kumactl_tarball "$arch" "$system"
    fi

    msg
    msg_green "... success!"
    msg
  done
}

function artifact_dir() {
  local arch=$1
  local system=$2

  echo "build/artifacts-$system-$arch"
}

function archive_path() {
  local arch=$1
  local system=$2
  local distro=$3

  if [[ -n $distro ]]; then
    echo "$(artifact_dir "$arch" "$system")/$RELEASE_NAME-$KUMA_VERSION-$distro-$arch.tar.gz"
  else
    echo "$(artifact_dir "$arch" "$system")/$RELEASE_NAME-$CTL_NAME-$KUMA_VERSION-$system-$arch.tar.gz"
  fi
}

function release() {
  for os in "${DISTRIBUTIONS[@]}"; do
    local system
    system=$(echo "$os" | awk '{split($0,parts,":"); print parts[1]}')
    local distro
    distro=$(echo "$os" | awk '{split($0,parts,":"); print parts[2]}')
    local arch
    arch=$(echo "$os" | awk '{split($0,parts,":"); print parts[4]}')

    local artifact
    artifact="$(archive_path "$arch" "$system" "$distro")"
    [ ! -f "$artifact" ] && msg_yellow "Package '$artifact' not found, skipping..." && continue

    if [[ -n $distro ]]; then
      msg_green ">>> Releasing ${RELEASE_NAME} $KUMA_VERSION for $distro ($system-$arch)..."
    else
      msg_green ">>> Releasing ${RELEASE_NAME} $KUMA_VERSION static kumactl binary for $system-$arch..."
    fi

    docker run --rm \
      -e PULP_USERNAME="${PULP_USERNAME}" -e PULP_PASSWORD="${PULP_PASSWORD}" \
      -e PULP_HOST=${PULP_HOST} \
      -v "${PWD}":/files:ro -it kong/release-script \
      --file /files/"$artifact" \
      --package-type ${PULP_PACKAGE_TYPE} --dist-name ${PULP_DIST_NAME} --publish
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
    *)
      usage
      break
      ;;
    esac
    shift
  done

  case $op in
  package)
    package
    ;;
  release)
    [ -z "$PULP_USERNAME" ] && msg_err "PULP_USERNAME required"
    [ -z "$PULP_PASSWORD" ] && msg_err "PULP_PASSWORD required"

    release
    ;;
  esac
}

main "$@"
