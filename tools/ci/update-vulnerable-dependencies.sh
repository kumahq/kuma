#!/usr/bin/env bash

set -e

command -v osv-scanner >/dev/null 2>&1 || { echo >&2 "osv-scanner not installed!"; exit 1; }
command -v jq >/dev/null 2>&1 || { echo >&2 "jq not installed!"; exit 1; }

for dep in $(osv-scanner --lockfile=go.mod --json | jq -c '.results[].packages[] | .package.name as $vulnerablePackage | {
  name: $vulnerablePackage,
  current: .package.version,
  fixedVersions: [.vulnerabilities[].affected[] | select(.package.name == $vulnerablePackage) | .ranges[].events[] | select(.fixed != null) | .fixed] | unique
} | select(.fixedVersions | length > 0) | select(.name != "github.com/kumahq/kuma")'); do
  IFS=. read -r currentMajor currentMinor currentPatch <<< "$(jq -r .current <<< "$dep")"
  # Update to the first version that's greater than our current version
  for version in $(jq -cr .fixedVersions[] <<< "$dep" | sort -V); do # sort supports semver sort
    if [[ "$version" =~ ^([[:digit:]]+)\.([[:digit:]]+)\.([[:digit:]]+) ]]; then
        fixMajor="${BASH_REMATCH[1]}"
        fixMinor="${BASH_REMATCH[2]}"
        fixPatch="${BASH_REMATCH[3]}"
    else
        >&2 echo "Couldn't parse fix version ${version} as semver!"
        exit 1
    fi
    # Do not downgrade in order to fix the vulnerability
    if ((currentMajor <= fixMajor)) && ((currentMinor <= fixMinor)) && ((currentPatch <= fixPatch)); then
      package=$(jq -r .name <<< "$dep")
      go get -u "$package"@v"$version"
      break
    fi
  done
done

go mod tidy
