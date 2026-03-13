#!/usr/bin/env bash
# SubagentStop/general-purpose hook for kuma-manual-test
# M13: warn if preflight agent output lacks pass/fail or state path
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/_lib.sh"

input="$(cat)"

# Prevent infinite loop
stop_active="$(printf '%s' "${input}" | jq -r '.stop_hook_active // false')"
if [[ "${stop_active}" == "true" ]]; then
  exit 0
fi

message="$(printf '%s' "${input}" | jq -r '.last_assistant_message // ""')"

if [[ -z "${message}" ]]; then
  exit 0
fi

missing=()

# Check for pass/fail keyword
if ! printf '%s' "${message}" | grep -qiE '\b(pass|fail)\b'; then
  missing+=("pass/fail verdict")
fi

# Check for state capture path
if ! printf '%s' "${message}" | grep -qF 'state/'; then
  missing+=("state capture path")
fi

if [[ ${#missing[@]} -gt 0 ]]; then
  joined="$(IFS=', '; echo "${missing[*]}")"
  jq -nc --arg missing "${joined}" '{
    systemMessage: ("[KMT013] Preflight agent output may be incomplete. Missing: " + $missing + ". Review the output before proceeding to Phase 4.")
  }'
fi

exit 0
