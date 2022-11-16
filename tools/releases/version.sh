#!/usr/bin/env bash

# Kuma version info has the following format:
# version git-tag git-commit build-date envoy-version

# version is built as follows:
# 1) If a git tag is present on the current commit, then the version is a git tag (without a starting v)
# 2) If the branch is "release-X.Y" look at the existing tags and use either X.Y.0-<shortHash> if there's none or
#     the latest tag with a patch version increased by one (.e.g if latest tag is X.Y.1 the version will be `X.Y.2-<shortHash>`)
# 3) In non release branch use `0.0.0-$shortHash`

# git-tag - if the current HEAD has a tag then use it otherwise it's "not-tagged"
# git-commit - HEAD sha
# build-date - local date if built on CI

# Note: this format must be changed carefully, other scripts depend on it
indexStatus=$(git diff --quiet && echo "clean-index" || echo "dirty-index")
shortHash=$(git rev-parse --short HEAD 2> /dev/null || echo "no-commit")
currentBranch=$(git rev-parse --abbrev-ref HEAD 2> /dev/null || echo "no-branch")
envoyVersion=$("${BASH_SOURCE%/*}"/../envoy/version.sh)

if [[ -z "${CI}" ]]; then
versionDate="local-build"
describedTag="local-build"
longHash="local-build"
else
versionDate=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
describedTag=$(shell git describe --tags 2>/dev/null || echo "unknown")
longHash=$(git rev-parse HEAD 2>/dev/null || echo "no-commit")
fi

exactTag=$(git describe --exact-match --tags 2> /dev/null || echo "not-tagged")
# We only support tags of the format: "v?X.Y.Z(-<alphaNumericName>)?" all other tags will just be ignored and use the regular versioning scheme
if [[ ${exactTag} =~ ^v?[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+)?$ ]]; then
  echo "${exactTag/^v//} $describedTag $longHash $versionDate $envoyVersion"
  exit 0
fi

if [[ ${currentBranch} == release-* ]]; then
    releasePrefix=${currentBranch//release-/}
    lastGitTag=$(git tag -l | grep -E "^v?${releasePrefix}\.[0-9]+$" | sed 's/^v//'| sort -V | tail -1)
    if [[ ${lastGitTag} ]]; then
      IFS=. read -r major minor patch <<< "${lastGitTag}"
      version="${major}.${minor}.$((++patch))-preview.v${shortHash}"
    else
      version="${releasePrefix}.0-preview.v${shortHash}"
    fi
else
  version="0.0.0-preview.v${shortHash}"
fi

echo "$version $describedTag $longHash $versionDate $envoyVersion"
