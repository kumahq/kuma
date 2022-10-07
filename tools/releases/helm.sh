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

# dev_version updates Chart.yaml with:
#  appVersion equal to the kuma version
#  version:
#    with the git commit suffix, if the kuma version has a git commit suffix
#    otherwise it leaves version alone, assuming it's been chosen intentionally
#  dependencies:
#    for any dependency $chart where charts/$chart exists, it deletes $chart from
#    .dependencies so that the embedded $chart is used and not the one fetched
#    from the repository. `cr` fetches explicit dependencies and they take
#    precedence over embedded files.
function dev_version {
  for dir in "${CHARTS_DIR}"/*; do
    if [ ! -d "${dir}" ]; then
      continue
    fi

    # Fail if there are uncommitted changes
    git diff --exit-code HEAD -- "${dir}"

    kuma_version=$("${SCRIPT_DIR}/version.sh")
    yq -i ".appVersion = \"${kuma_version}\"" "${dir}/Chart.yaml"
    yq -i ".version = \"${kuma_version}\"" "${dir}/Chart.yaml"

    for chart in $(yq e '.dependencies[].name' "${dir}/Chart.yaml"); do
        if [ ! -d "${dir}/charts/${chart}" ]; then
            continue
        fi
        yq -i e "del(.dependencies[] | select(.name == \"${chart}\"))" "${dir}/Chart.yaml"
    done

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
  [ -z "$GH_REPO_URL" ] && GH_REPO_URL="https://${GH_TOKEN}@github.com/${GH_OWNER}/${GH_REPO}.git"

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
      --dev-version)
        op="dev-version"
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
    dev-version)
      if ! command -v yq >/dev/null || ! [[ $(command yq --version) =~ "mikefarah" ]]; then
        msg_err "mikefarah/yq not installed"
      fi

      dev_version
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
