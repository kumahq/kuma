#!/usr/bin/env bash

# `release-tool release helm-chart` only checks the charts repo's GitHub release.
# `helm install` resolves charts from the index.yaml served by GitHub Pages, which is
# regenerated separately and can lag behind (or miss) that release, so the chart can be
# released and still not be installable. Verify it from a consumer's point of view:
# add the repo, refresh the index, and pull the released version.

set -euo pipefail

repo_name="${HELM_REPO_NAME:?}"
repo_url="${HELM_REPO_URL:?}"
charts="${HELM_CHARTS:?}"
release="${RELEASE:?}"

workdir="$(mktemp -d)"
trap 'rm -rf "$workdir"' EXIT

helm repo add "$repo_name" "$repo_url" --force-update
helm repo update "$repo_name"

failed=0
IFS=',' read -ra chart_names <<<"$charts"
for chart in "${chart_names[@]}"; do
  ref="${repo_name}/${chart}"

  # `helm search repo` matches substrings and keywords, so filter on the exact chart ref.
  if ! helm search repo "$ref" --version "$release" --devel --output json |
    jq -e --arg ref "$ref" --arg version "$release" \
      'map(select(.name == $ref and .version == $version)) | length > 0' >/dev/null; then
    echo "::error::${ref} ${release} is missing from ${repo_url} index" >&2
    failed=1
    continue
  fi

  if ! helm pull "$ref" --version "$release" --devel --destination "$workdir"; then
    echo "::error::${ref} ${release} is indexed by ${repo_url} but cannot be pulled" >&2
    failed=1
    continue
  fi

  echo "${ref} ${release} is installable from ${repo_url}"
done

exit "$failed"
