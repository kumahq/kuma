#!/usr/bin/env bash

set -e

command -v osv-scanner >/dev/null 2>&1 || { echo >&2 "osv-scanner not installed!"; exit 1; }
command -v jq >/dev/null 2>&1 || { echo >&2 "jq not installed!"; exit 1; }

compare_versions() {
    local majorV1=$1
    local minorV1=$2
    local patchV1=$3
    local majorV2=$4
    local minorV2=$5
    local patchV2=$6

    if ((majorV1 < majorV2)); then
        echo -1
        return
    elif ((majorV1 > majorV2)); then
        echo 1
        return
    fi

    if ((minorV1 < minorV2)); then
        echo -1
        return
    elif ((minorV1 > minorV2)); then
        echo 1
        return
    fi

    if ((patchV1 < patchV2)); then
        echo -1
        return
    elif ((patchV1 > patchV2)); then
        echo 1
        return
    else
        echo 0
    fi
}

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
    if [[ $(compare_versions currentMajor currentMinor currentPatch fixMajor fixMinor fixPatch) -eq -1]]; then
      package=$(jq -r .name <<< "$dep")
      if [[ "$package" == "stdlib" ]]; then
        yq -i e ".parameters.go_version.default = \"$version\"" .circleci/config.yml
        go mod edit -go="$version"
      else
        go get -u "$package"@v"$version"
      fi
      break
    fi
  done
done

go mod tidy
