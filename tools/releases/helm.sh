#!/usr/bin/env bash

set -e

source "$(dirname "$(dirname "$0")")/common.sh" # relative path to ../common.sh

[ -z "$GH_OWNER" ] && GH_OWNER="kumahq"
[ -z "$GH_REPO" ] && GH_REPO="charts"
CHARTS_REPO_URL="https://$GH_OWNER.github.io/$GH_REPO"
CHARTS_DIR="./deployments/charts"
CHARTS_PACKAGE_PATH=".cr-release-packages"
CHARTS_INDEX_FILE="index.yaml"
GH_PAGES_BRANCH="gh-pages"
[ -z "$GH_REPO_URL" ] && GH_REPO_URL="git@github.com:${GH_OWNER}/${GH_REPO}.git"

function package {
  # First package all the charts
  for dir in "${CHARTS_DIR}"/*; do
    if [ ! -d "$dir" ]; then
      continue
    fi

    cr package \
      --package-path "${CHARTS_PACKAGE_PATH}" \
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
  git clone --single-branch --branch "${GH_PAGES_BRANCH}" $GH_REPO_URL

  cr index \
    --owner "${GH_OWNER}" \
    --git-repo "${GH_REPO}" \
    --charts-repo "${CHARTS_REPO_URL}" \
    --package-path "${CHARTS_PACKAGE_PATH}" \
    --index-path "${GH_REPO}/${CHARTS_INDEX_FILE}"

  pushd ${GH_REPO}
  # tell git who we are before adding the index file
  git config user.email "helm@kuma.io"
  git config user.name "Helm Releaser"
  git add "${CHARTS_INDEX_FILE}"
  git commit -m "ci(helm) publish charts"
  git push
  popd
  rm -rf ${GH_REPO}
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
      *)
        usage
        break
        ;;
    esac
    shift
  done

  [ -z "${GH_TOKEN}" ] && msg_err "GH_TOKEN required"

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

