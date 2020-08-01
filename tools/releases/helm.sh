#!/usr/bin/env bash

set -e

CHARTS_REPO_URL="https://kumahq.github.io/charts"
CHARTS_DIR="./deployments/charts"
CHARTS_PACKAGE_PATH=".cr-release-packages"
CHARTS_INDEX_FILE="index.yaml"
GH_PAGES_BRANCH="gh-pages"
GH_OWNER="kumahq"
GH_REPO="charts"

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

function package {
  # First package all the charts
  for dir in "${CHARTS_DIR}"/*; do
    if [ ! -d "$dir" ]; then
      continue
    fi

    helm package \
      --app-version "${KUMA_VERSION}" \
      --destination "${CHARTS_PACKAGE_PATH}" \
      --dependency-update \
      "$dir"
  done
}

function release {
  # First upload the packaged charts to the release
  cr upload \
    --owner "${GH_OWNER}" \
    --git-repo "${GH_REPO}" \
    --token "${GH_TOKEN}" \
    --package-path "${CHARTS_PACKAGE_PATH}"


  # Then build and upload the index file to github pages
  git clone --single-branch --branch "${GH_PAGES_BRANCH}" git@github.com:kumahq/${GH_REPO}.git

  cr index \
    --owner "${GH_OWNER}" \
    --git-repo "${GH_REPO}" \
    --charts-repo "${CHARTS_REPO_URL}" \
    --package-path "${CHARTS_PACKAGE_PATH}" \
    --index-path "charts/${CHARTS_INDEX_FILE}"

  pushd charts
  git add "${CHARTS_INDEX_FILE}"
  git commit -m "ci(helm) publish charts for version ${KUMA_VERSION}@${KUMA_COMMIT}"
  git push
  popd
  rm -rf charts
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

  [ -z "${GH_TOKEN}" ] && msg_err "GH_TOKEN required"
  [ -z "${KUMA_VERSION}" ] && msg_err "Error: --version required"
  [ -z "${KUMA_COMMIT}" ] && msg_err "Error: --sha required"

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

