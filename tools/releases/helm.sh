#!/usr/bin/env bash

set -e

SCRIPT_DIR="$(dirname -- "${BASH_SOURCE[0]}")"
source "${SCRIPT_DIR}/../common.sh"

[ -z "$GH_OWNER" ] && GH_OWNER="kumahq"
[ -z "$GH_REPO" ] && GH_REPO="charts"
CHARTS_REPO_URL="https://$GH_OWNER.github.io/$GH_REPO"
CHARTS_DIR="./deployments/charts"
CHARTS_PACKAGE_PATH=".cr-release-packages"
CHARTS_INDEX_FILE="index.yaml"
GH_PAGES_BRANCH="gh-pages"

# This updates Chart.yaml with:
#  appVersion equal to the kuma version
#  version:
#    using version.sh logic
#  dependencies (with non-empty first argument):
#    for any dependency $chart where charts/$chart exists, it deletes $chart from
#    .dependencies so that the embedded $chart is used and not the one fetched
#    from the repository. `cr` fetches explicit dependencies and they take
#    precedence over embedded files.
function update_version {
  dev=${1}
  for dir in "${CHARTS_DIR}"/*; do
    if [ ! -d "${dir}" ]; then
      continue
    fi

    # Fail if there are uncommitted changes
    git diff --exit-code HEAD -- "${dir}"

    kuma_version=$("${SCRIPT_DIR}/version.sh")
    yq -i ".appVersion = \"${kuma_version}\"" "${dir}/Chart.yaml"
    yq -i ".version = \"${kuma_version}\"" "${dir}/Chart.yaml"

    if [ -n "${dev}" ]; then
      for chart in $(yq e '.dependencies[].name' "${dir}/Chart.yaml"); do
          if [ ! -d "${dir}/charts/${chart}" ]; then
              continue
          fi
          yq -i e "del(.dependencies[] | select(.name == \"${chart}\"))" "${dir}/Chart.yaml"
      done
    fi
  done
}

function package {
  # First package all the charts
  for dir in "${CHARTS_DIR}"/*; do
    if [ ! -d "${dir}" ]; then
      continue
    fi

    # Fail if there are uncommitted changes
    git diff --exit-code HEAD -- "${dir}"

    # TODO remove this when Gateway API is no longer experimental
    # Gateway CRDs are installed conditionally via install missing CRDs job
    if [[ $(basename "${dir}") == "kuma" ]]; then
      find "${dir}/crds" -name "*meshgatewayconfigs.yaml" -delete
    fi

    cr package \
      --package-path "${CHARTS_PACKAGE_PATH}" \
      "${dir}"

    # Restore files removed above
    git checkout -- "${dir}"
  done
}

function release {
  if [ -n "${GITHUB_APP}" ]; then
    [ -z "$GH_REPO_URL" ] && GH_REPO_URL="https://x-access-token:${GH_TOKEN}@github.com/${GH_OWNER}/${GH_REPO}.git"
  else
    [ -z "$GH_REPO_URL" ] && GH_REPO_URL="https://${GH_TOKEN}@github.com/${GH_OWNER}/${GH_REPO}.git"
  fi

  # First upload the packaged charts to the release
  cr upload \
    --owner "${GH_OWNER}" \
    --git-repo "${GH_REPO}" \
    --token "${GH_TOKEN}" \
    --package-path "${CHARTS_PACKAGE_PATH}"

  # Then build and upload the index file to github pages
  git clone --single-branch --branch "${GH_PAGES_BRANCH}" "$GH_REPO_URL"

  cr index \
    --owner "${GH_OWNER}" \
    --git-repo "${GH_REPO}" \
    --charts-repo "${CHARTS_REPO_URL}" \
    --package-path "${CHARTS_PACKAGE_PATH}" \
    --index-path "${GH_REPO}/${CHARTS_INDEX_FILE}"

  pushd ${GH_REPO}

  git config user.name "${GH_USER}"
  git config user.email "${GH_EMAIL}"

  git add "${CHARTS_INDEX_FILE}"
  git commit -m "ci(helm): publish charts"
  git push

  popd
  rm -rf ${GH_REPO}
}


function usage {
  echo "Usage: $0 [--package|--release|--update-version [--dev]]"
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
      --update-version)
        op="update-version"
        ;;
      --dev)
        dev="true"
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
    update-version)
      if ! command -v yq >/dev/null || ! [[ $(command yq --version) =~ "mikefarah" ]]; then
        msg_err "mikefarah/yq not installed"
      fi

      update_version "${dev}"
      ;;
    release)
      if [ -z "${GH_TOKEN}" ] || [ -z "${GH_USER}" ] || [ -z "${GH_EMAIL}" ]; then
        msg_err "GH_TOKEN, GH_USER and GH_EMAIL required"
      fi

      release
      ;;
  esac
}


main "$@"
