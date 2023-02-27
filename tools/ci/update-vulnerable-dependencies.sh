#!/usr/bin/env bash

command -v osv-scanner >/dev/null 2>&1 || { echo >&2 "osv-scanner not installed!"; exit 1; }
command -v jq >/dev/null 2>&1 || { echo >&2 "jq not installed!"; exit 1; }

for dep in $(osv-scanner --lockfile=go.mod --json | jq -c '.results[].packages[] | .package.name as $vulnerablePackage | {
  name: $vulnerablePackage,
  current: .package.version,
  fixedVersions: [.vulnerabilities[].affected[] | select(.package.name == $vulnerablePackage) | .ranges[].events[] | select(.fixed != null) | .fixed] | unique
} | select(.fixedVersions | length > 0)'); do
  IFS=. read -r currentMajor currentMinor currentPatch <<< "$(jq -r .current <<< "$dep")"
  # Update to the first version that's greater than our current version
  for version in $(jq -cr .fixedVersions[] <<< "$dep" | sort -V); do # sort supports semver sort
    IFS=. read -r fixMajor fixMinor fixPatch <<< "$version"
    # Do not downgrade in order to fix the vulnerability
    if ((currentMajor <= fixMajor)) && ((currentMinor <= fixMinor)) && ((currentPatch <= fixPatch)); then
      package=$(jq -r .name <<< "$dep")
      go get -u "$package"@v"$version"
      break
    fi
  done
done
go mod tidy

