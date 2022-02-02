#!/usr/bin/env bash

# Kuma version is built as follows:
# 1) If a git tag is present on the current commit, then the version is a git tag
# 2) If the branch starts with "release" (like "release-1.3"), then the version is "$lastGitTag-$shortHash" (for example: 1.4.1-450174242)
#    At the same time, if $lastGitTag is not final tag (for example 1.4.0-rc1 or 1.4.0-preview1), the version is "dev-$shortHash"
# 3) If the branch does not start with "release", then the version is "dev-$shortHash" (for example: dev-450174242)

lastGitTag=$(git describe --abbrev=0 --tags)
shortHash=$(git rev-parse --short HEAD)
currentBranch=$(git rev-parse --abbrev-ref HEAD)

if git describe --exact-match --tags > /dev/null 2>&1; then # if we are on tag
  echo "$lastGitTag"
else
  if [[ $lastGitTag =~ [0-9]+\.[0-9]+\.[0-9]+ ]] && [[ $currentBranch == release* ]]; then
    # set field separator to dot and read parts of semver
    IFS=. read -r major minor patch <<< "$lastGitTag"
    echo "$major.$minor.$((++patch))-$shortHash"
  else
    echo "dev-$shortHash"
  fi
fi
