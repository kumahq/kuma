#!/usr/bin/env bash

set -e

SCRIPT_DIR="$(dirname -- "${BASH_SOURCE[0]}")"
source "${SCRIPT_DIR}/../common.sh"

GOARCH=(amd64)

# first component is the distribution name, second is the system - must map to
# valid $GOOS values
DISTRIBUTIONS=(debian:linux:alpine ubuntu:linux:alpine rhel:linux:centos centos:linux:centos darwin:darwin:darwin)

PULP_HOST="https://api.pulp.konnect-prod.konghq.com"
PULP_PACKAGE_TYPE="mesh"
PULP_DIST_NAME="alpine"
[ -z "$RELEASE_NAME" ] && RELEASE_NAME="kuma"
ENVOY_VERSION=1.21.1
[ -z "$KUMA_CONFIG_PATH" ] && KUMA_CONFIG_PATH=pkg/config/app/kuma-cp/kuma-cp.defaults.yaml

function get_envoy() {
  local distro=$1
  local envoy_distro=$2

  local status
  status=$(curl -L -o build/envoy-"$distro" \
    --write-out '%{http_code}' --silent --output /dev/null \
    "https://download.konghq.com/mesh-alpine/envoy-$ENVOY_VERSION-$envoy_distro")

  if [ "$status" -ne "200" ]; then msg_err "Error: failed downloading Envoy"; fi
}

function create_tarball() {
  local system=$1
  local arch=$2
  local distro=$3
  local envoy_distro=$4

  local dest_dir=build/$RELEASE_NAME-$distro-$arch
  local kuma_dir=$dest_dir/$RELEASE_NAME-$KUMA_VERSION

  rm -rf "$dest_dir"
  mkdir "$dest_dir"
  mkdir "$kuma_dir"
  mkdir "$kuma_dir/bin"
  mkdir "$kuma_dir/conf"

  get_envoy "$distro" "$envoy_distro"
  chmod 755 build/envoy-"$distro"

  cp -p "build/envoy-$distro" "$kuma_dir"/bin/envoy
  cp -p "build/artifacts-$system-$arch/kuma-cp/kuma-cp" "$kuma_dir/bin"
  cp -p "build/artifacts-$system-$arch/kuma-dp/kuma-dp" "$kuma_dir/bin"
  cp -p "build/artifacts-$system-$arch/kumactl/kumactl" "$kuma_dir/bin"
  cp -p "build/artifacts-$system-$arch/coredns/coredns" "$kuma_dir/bin"
  cp -p "build/artifacts-$system-$arch/kuma-prometheus-sd/kuma-prometheus-sd" "$kuma_dir/bin"
  cp -p "$KUMA_CONFIG_PATH" "$kuma_dir/conf/kuma-cp.conf.yml"

  cp tools/releases/templates/* "$kuma_dir"

  tar -czf "build/artifacts-$system-$arch/$RELEASE_NAME-$KUMA_VERSION-$distro-$arch.tar.gz" -C "$dest_dir" .
}

function package() {
  for os in "${DISTRIBUTIONS[@]}"; do
    local distro
    distro=$(echo "$os" | awk '{split($0,parts,":"); print parts[1]}')
    local system
    system=$(echo "$os" | awk '{split($0,parts,":"); print parts[2]}')
    local envoy_distro
    envoy_distro=$(echo "$os" | awk '{split($0,parts,":"); print parts[3]}')

    for arch in "${GOARCH[@]}"; do

      msg ">>> Packaging Kuma for $distro ($system-$arch)..."
      msg

      make GOOS="$system" GOARCH="$arch" build
      create_tarball "$system" "$arch" "$distro" "$envoy_distro"

      msg
      msg_green "... success!"
      msg
    done
  done
}

function release() {
  for os in "${DISTRIBUTIONS[@]}"; do
    local distro
    distro=$(echo "$os" | awk '{split($0,parts,":"); print parts[1]}')
    local system
    system=$(echo "$os" | awk '{split($0,parts,":"); print parts[2]}')

    for arch in "${GOARCH[@]}"; do
      local artifact
      artifact="build/artifacts-$system-$arch/$RELEASE_NAME-$KUMA_VERSION-$distro-$arch.tar.gz"
      [ ! -f "$artifact" ] && msg_yellow "Package '$artifact' not found, skipping..." && continue

      msg_green "Releasing Kuma for '$os', '$arch', '$KUMA_VERSION'..."

      docker run --rm \
        -e PULP_USERNAME="${PULP_USERNAME}" -e PULP_PASSWORD="${PULP_PASSWORD}" \
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
  KUMA_VERSION=$("${SCRIPT_DIR}/version.sh")

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

  [ -z "$PULP_USERNAME" ] && msg_err "PULP_USERNAME required"
  [ -z "$PULP_PASSWORD" ] && msg_err "PULP_PASSWORD required"

  case $op in
  package)
    package
    ;;
  release)
    release
    ;;
  esac
}

main "$@"
