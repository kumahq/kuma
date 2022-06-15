#!/usr/bin/env bash

# Kuma version is built as follows:
# 1) If a git tag is present on the current commit, then the version is a git tag (without a starting v)
# 2) If the branch is "release-X.Y" look at the existing tags and use either X.Y.0-<shortHash> if there's none or
#     the latest tag with a patch version increased by one (.e.g if latest tag is X.Y.1 the version will be `X.Y.2-<shortHash>`)
# 3) In non release branch use `0.0.0-$shortHash`

# Note: this format must be changed carefully, other scripts depend on it
exactTag=$(git describe --exact-match --tags 2> /dev/null)
# We only support tags of the format: "v?X.Y.Z(-<alphaNumericName>)?" all other tags will just be ignored and use the regular versioning scheme
if [[ ${exactTag} =~ ^v?[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+)?$ ]]; then
  echo "${exactTag/^v//}"
  exit 0
fi

shortHash=$(git rev-parse --short HEAD 2> /dev/null)
currentBranch=$(git rev-parse --abbrev-ref HEAD 2> /dev/null)
if [[ ${currentBranch} == release-* ]]; then
    releasePrefix=${currentBranch//release-/}
    lastGitTag=$(git tag -l | grep -E "^v?${releasePrefix}\.[0-9]+$" | sed 's/^v//'| sort -V | tail -1)
    if [[ ${lastGitTag} ]]; then
      IFS=. read -r major minor patch <<< "${lastGitTag}"
      echo "${major}.${minor}.$((++patch))-preview.${shortHash}"
    else
      echo "${releasePrefix}.0-preview.${shortHash}"
    fi
else
  echo "0.0.0-preview.${shortHash}"
fi
