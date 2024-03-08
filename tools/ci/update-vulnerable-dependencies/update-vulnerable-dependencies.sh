#!/usr/bin/env bash

set -e

command -v osv-scanner >/dev/null 2>&1 || { echo >&2 "osv-scanner not installed!"; exit 1; }
command -v jq >/dev/null 2>&1 || { echo >&2 "jq not installed!"; exit 1; }


for dep in $(osv-scanner --lockfile=go.mod --json | jq -c '.results[].packages[] | .package.name as $vulnerablePackage | {
  name: $vulnerablePackage,
  current: .package.version,
  fixedVersions: [.vulnerabilities[].affected[] | select(.package.name == $vulnerablePackage) | .ranges[].events |
  map(select(.fixed != null) | .fixed)] | map(select(length > 0)) } | select(.name != "github.com/kumahq/kuma")'); do

  fixVersion=$(go run tools/ci/update-vulnerable-dependencies/main.go <<< "$dep")

  if [ "$fixVersion" != "null" ]; then
    package=$(jq -r .name <<< "$dep")

    echo "Updating $package to $fixVersion"

    if [[ "$package" == "stdlib" ]]; then
      yq -i e ".parameters.go_version.default = \"$fixVersion\"" .circleci/config.yml
      go mod edit -go="$fixVersion"
    else
      go get -u "$package"@v"$fixVersion"
    fi
  fi
done

go mod tidy
