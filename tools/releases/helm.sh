#!/usr/bin/env bash

set -e

SCRIPT_DIR="$(dirname -- "${BASH_SOURCE[0]}")"
source "${SCRIPT_DIR}/../common.sh"

[ -z "$GH_OWNER" ] && GH_OWNER="kumahq"
[ -z "$GH_REPO" ] && GH_REPO="charts"
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
  local dev=$1
  local kuma_version
  local chart

  for dir in "${CHARTS_DIR}"/*; do
    if [ ! -d "${dir}" ]; then
      continue
    fi

    # Fail if there are uncommitted changes
    git diff --exit-code HEAD -- "${dir}"

    kuma_version=$("${SCRIPT_DIR}/version.sh" | cut -d " " -f 1)
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

    make helm-docs
  done
}

function package {
  local f

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

    # repackage archive to remove potential duplicates
    for f in "${CHARTS_PACKAGE_PATH}"/*.tgz; do
      local tmpdir
      tmpdir=$(mktemp --directory)

      tar -xzf "${f}" --directory "${tmpdir}"

      tar -czf "${f}" \
        --directory "${tmpdir}" \
        --owner 0 \
        --group 0 \
        "$(basename "${dir}")"
    done

    # Restore files removed above
    git checkout -- "${dir}"
  done
}

function release {
  local CHART_TAR
  local CHART_FILE
  local CHART_VERSION

  if [ -z "${GH_TOKEN}" ] || [ -z "${GH_USER}" ] || [ -z "${GH_EMAIL}" ]; then
    msg_err "GH_TOKEN, GH_USER and GH_EMAIL are required"
  fi

  # This line assigns a default value to GH_REPO_URL only if it is unset or empty.
  # Syntax: ${VAR:=default} sets VAR to 'default' if VAR is unset or null.
  #
  # It constructs the GitHub HTTPS URL using the provided token for authentication.
  # The expression ${GITHUB_APP:+x-access-token:} conditionally inserts the prefix
  # "x-access-token:" only when GITHUB_APP is set and non-empty.
  # This is required for GitHub App tokens, while personal access tokens omit the prefix.
  #
  # The leading colon ':' is a no-op command used here to enable safe default assignment.
  : "${GH_REPO_URL:=https://${GITHUB_APP:+x-access-token:}${GH_TOKEN}@github.com/${GH_OWNER}/${GH_REPO}.git}"

  git clone --single-branch --branch "${GH_PAGES_BRANCH}" "${GH_REPO_URL}"

  CHART_TAR=$(find "${CHARTS_PACKAGE_PATH}" -name "*.tgz" -type f | head -n 1)
  CHART_FILE=$(tar -tf "${CHART_TAR}" | grep 'Chart.yaml')
  CHART_VERSION=$(tar -zxOf "${CHART_TAR}" "${CHART_FILE}" | yq .version)

  pushd "${GH_REPO}"

  # First upload the packaged charts to the release
  cr upload \
    --owner "${GH_OWNER}" \
    --git-repo "${GH_REPO}" \
    --token "${GH_TOKEN}" \
    --package-path "../${CHARTS_PACKAGE_PATH}"

  # Then build and upload the index file to github pages
  cr index \
    --owner "${GH_OWNER}" \
    --git-repo "${GH_REPO}" \
    --package-path "../${CHARTS_PACKAGE_PATH}" \
    --index-path "${CHARTS_INDEX_FILE}"

  git config user.name "${GH_USER}"
  git config user.email "${GH_EMAIL}"

  git add "${CHARTS_INDEX_FILE}"
  git commit -m "ci(helm): publish chart version ${CHART_VERSION}"
  git push

  popd
  rm -rf "${GH_REPO}"
}


function usage {
  echo "Usage: $0 [--package|--release|--update-version [--dev]]"
  exit 0
}


function main {
  local flag=$1
  local op
  local dev

  while [[ $# -gt 0 ]]; do
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
        ;;
    esac
    shift
  done

  case $op in
    package)
      package
      ;;
    update-version)
      update_version "${dev}"
      ;;
    release)
      release
      ;;
  esac
}


main "$@"
