#!/usr/bin/env bash

set -Eeuo pipefail

command -v osv-scanner >/dev/null 2>&1 || { echo >&2 "osv-scanner not installed!"; exit 1; }
command -v jq >/dev/null 2>&1 || { echo >&2 "jq not installed!"; exit 1; }

SCRIPT_PATH="${BASH_SOURCE[0]:-$0}";
SCRIPT_DIR="$(dirname -- "$SCRIPT_PATH")"

OSV_FLAGS=(--lockfile=go.mod --json)

# Loop over the array, add only non-empty values to the new array
for i in "${OSV_SCANNER_ADDITIONAL_OPTS[@]}"; do
   # Skip null items
   if [ -z "$i" ]; then
     continue
   fi

   # Add the rest of the elements to an array
   OSV_FLAGS+=("${i}")
done

SKIP_PACKAGES='["github.com/kumahq/kuma", "github.com/open-policy-agent/opa"]'

for dep in $(osv-scanner "${OSV_FLAGS[@]}" | jq -c --argjson skipPackages "$SKIP_PACKAGES" '
    .results[].packages[]
    | .package.name as $vulnerablePackage
    | select($vulnerablePackage | IN($skipPackages[]) | not)
    | {
        name: $vulnerablePackage,
        current: .package.version,
        fixedVersions: [
          .vulnerabilities[].affected[]
          | select(.package.name == $vulnerablePackage)
          | .ranges[].events
          | map(select(.fixed != null) | .fixed)
        ] | map(select(length > 0))
      }
  '); do

  fixVersion=$(go run "$SCRIPT_DIR"/main.go <<< "$dep")

  if [ "$fixVersion" != "null" ]; then
    package=$(jq -r .name <<< "$dep")

    echo "Updating $package to $fixVersion"

    if [[ "$package" == "stdlib" ]]; then
      go mod edit -go="$fixVersion"
    else
      go get "$package"@v"$fixVersion"
    fi
  fi
done

go mod tidy
