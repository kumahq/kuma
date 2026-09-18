#!/usr/bin/env bash
# List MADRs with their status, summary and tags.
#
# Output format:
#   <file> [<status>] <summary> #tag #tag
#
# Usage:
#   ./list.sh                              # list all MADRs
#   ./list.sh | grep '\[accepted\]'        # example: filter by status using grep
#   ./list.sh | grep '#zone-egress'        # example: filter by tag using grep

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DECISIONS_DIR="${SCRIPT_DIR}/decisions"

for file in "${DECISIONS_DIR}"/*.md; do
  name=$(basename "$file")
  if [[ "$name" == "000-template.md" ]]; then
    continue
  fi

  awk -v name="$name" '
    # front matter: everything between the first pair of --- lines
    NR == 1 && $0 == "---" { fm = 1; next }
    fm && $0 == "---" { fm = 0; next }
    fm && /^status:/ { status = tolower(value()) }
    fm && /^summary:/ { summary = value() }
    fm && /^tags:/ { tags = value(); gsub(/[\[\]]/, "", tags) }

    # fall back to the "* Status: ..." bullet for MADRs without front matter
    !fm && !status && tolower($0) ~ /^[*-] status:/ {
      status = tolower(value()); gsub(/\*|<.*/, "", status); sub(/ *$/, "", status)
    }

    function value(  v) { v = $0; sub(/^[^:]*: */, "", v); return v }

    END {
      line = sprintf("%-72s [%s]", name, status ? status : "unknown")
      if (summary) line = line " " summary
      n = split(tags, t, / *, */)
      for (i = 1; i <= n; i++) if (t[i]) line = line " #" t[i]
      print line
    }
  ' "$file"
done
