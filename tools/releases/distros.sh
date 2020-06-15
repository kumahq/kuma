#!/usr/bin/env bash

set -e

GOARCH=( amd64 )

# first component is the distribution name, second is the system - must map to
# valid $GOOS values
DISTRIBUTIONS=(debian:linux ubuntu:linux rhel:linux centos:linux darwin:darwin)


BINTRAY_ENDPOINT="https://api.bintray.com/"
BINTRAY_SUBJECT="kong"
[ -z "$BINTRAY_REPOSITORY" ] && BINTRAY_REPOSITORY="kuma"
ENVOY_VERSION=1.14.2

function msg_green {
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

  local status=$(curl -L -o build/envoy-$distro -u $BINTRAY_USERNAME:$BINTRAY_API_KEY \
    --write-out %{http_code} --silent --output /dev/null \
    "https://kong.bintray.com/envoy/envoy-$ENVOY_VERSION-$distro")

  [ "$status" -ne "200" ] && msg_err "Error: failed downloading Envoy" || true
}


function create_tarball {
  local system=$1
  local arch=$2
  local distro=$3

  local dest_dir=build/kuma-$distro-$arch
  local kuma_dir=$dest_dir/kuma-$KUMA_VERSION

  rm -rf $dest_dir
  mkdir $dest_dir
  mkdir $kuma_dir
  mkdir $kuma_dir/bin
  mkdir $kuma_dir/conf

  get_envoy $distro
  chmod 755 build/envoy-$distro

  cp -p build/envoy-$distro $kuma_dir/bin/envoy
  cp -p build/artifacts-$system-$arch/kuma-cp/kuma-cp $kuma_dir/bin
  cp -p build/artifacts-$system-$arch/kuma-dp/kuma-dp $kuma_dir/bin
  cp -p build/artifacts-$system-$arch/kumactl/kumactl $kuma_dir/bin
  cp -p build/artifacts-$system-$arch/kuma-prometheus-sd/kuma-prometheus-sd $kuma_dir/bin
  cp -p pkg/config/app/kuma-cp/kuma-cp.defaults.yaml $kuma_dir/conf/kuma-cp.conf.yml

  cp tools/releases/templates/* $kuma_dir

  tar -czf build/artifacts-$system-$arch/kuma-$distro-$arch.tar.gz -C $dest_dir .
}


function package {
  for os in "${DISTRIBUTIONS[@]}"; do
    local distro=$(echo "$os" | awk '{split($0,parts,":"); print parts[1]}')
    local system=$(echo "$os" | awk '{split($0,parts,":"); print parts[2]}')

    for arch in "${GOARCH[@]}"; do

      msg ">>> Packaging Kuma for $distro ($system-$arch)..."
      msg

      make GOOS=$system GOARCH=$arch BUILD_INFO_GIT_TAG=$KUMA_VERSION BUILD_INFO_GIT_COMMIT=$KUMA_COMMIT build
      create_tarball $system $arch $distro

      msg
      msg_green "... success!"
      msg
    done
  done
}


function create_bintray_package {
  local package_name=$1
  local creation_status="$(curl --write-out %{http_code} --silent --output /dev/null \
    -XPOST -H 'Content-Type: application/json' -u $BINTRAY_USERNAME:$BINTRAY_API_KEY\
    -d '{"name":"'"$package_name"'"}' \
    $BINTRAY_ENDPOINT/packages/$BINTRAY_SUBJECT/$BINTRAY_REPOSITORY)"
  [ "$creation_status" -eq "409" ] && return
  [ "$creation_status" -ne "201" ] && msg_err "Error: could not create package $package_name"
}


function release {
  for os in "${DISTRIBUTIONS[@]}"; do
    local distro=$(echo "$os" | awk '{split($0,parts,":"); print parts[1]}')
    local system=$(echo "$os" | awk '{split($0,parts,":"); print parts[2]}')

    for arch in "${GOARCH[@]}"; do
      local artifact="build/artifacts-$system-$arch/kuma-$distro-$arch.tar.gz"
      [ ! -f "$artifact" ] && msg_yellow "Package '$artifact' not found, skipping..." && continue

      msg_green "Releasing Kuma for '$os', '$arch'..."

      local package_status="$(curl --write-out %{http_code} --silent --output /dev/null \
                                   -u $BINTRAY_USERNAME:$BINTRAY_API_KEY \
                                   $BINTRAY_ENDPOINT/content/$BINTRAY_SUBJECT/$BINTRAY_REPOSITORY/$distro)"
      [[ "$package_status" -eq "404" ]] && create_bintray_package "$distro"

      local upload_status=$(curl -T $artifact \
           --write-out %{http_code} --silent --output /dev/null \
           -u $BINTRAY_USERNAME:$BINTRAY_API_KEY \
           "$BINTRAY_ENDPOINT/content/$BINTRAY_SUBJECT/$BINTRAY_REPOSITORY/$distro/$KUMA_VERSION-$arch/kuma-$KUMA_VERSION-$distro-$arch.tar.gz?publish=1")

      [ "$upload_status" -eq "409" ] && msg_red "Error: package for '$os', '$arch' already exists" && continue
      [ "$upload_status" -ne "201" ] && msg_red "Error: could not upload package for '$os', '$arch' :(" && continue
      [ "$upload_status" -eq "201" ] && msg_green "Success! :)" && continue
    done
  done
}


function usage {
  echo "Usage: $0 [--package|--release]"
  exit 0
}


function main {
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

  [ -z "$BINTRAY_USERNAME" ] && msg_err "BINTRAY_USERNAME required"
  [ -z "$BINTRAY_API_KEY" ] && msg_err "BINTRAY_API_KEY required"
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

