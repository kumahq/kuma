#!/usr/bin/env bash
# List MADRs with their status. Optionally filter by status.
#
# Usage:
#   ./list.sh              # list all MADRs

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DECISIONS_DIR="${SCRIPT_DIR}/decisions"

for file in "${DECISIONS_DIR}"/*.md; do
  name=$(basename "$file")
  if [[ "$name" == "000-template.md" ]]; then
    continue
  fi

  # Extract status value, strip markdown bold and HTML comments
  status=$(awk 'tolower($0) ~ /^[*-] status:/ {sub(/.*: */, ""); gsub(/\*|<.*/, ""); sub(/ *$/, ""); print tolower($0); exit}' "$file")
  status="${status:-unknown}"

  printf "%-80s [%s]\n" "$name" "$status"
done
