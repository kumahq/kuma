#!/usr/bin/env bash
# SubagentStop/Explore hook for kuma-suite-author
# S7: warn if code-reading agent returned raw code instead of structured summary
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

# Check for code blocks exceeding 20 lines
has_oversized=false
in_block=false
block_lines=0

while IFS= read -r line; do
  if [[ "${line}" == '```'* ]]; then
    if [[ "${in_block}" == "true" ]]; then
      # Closing fence
      if [[ "${block_lines}" -gt 20 ]]; then
        has_oversized=true
        break
      fi
      in_block=false
      block_lines=0
    else
      # Opening fence
      in_block=true
      block_lines=0
    fi
  elif [[ "${in_block}" == "true" ]]; then
    block_lines=$((block_lines + 1))
  fi
done <<< "${message}"

if [[ "${has_oversized}" == "true" ]]; then
  jq -nc '{
    systemMessage: "[KSA007] Code-reading agent may have returned raw code instead of structured summary. Found code blocks exceeding 20 lines. Review the output - it should contain only G1-G7 material entries and S1-S7 signal entries."
  }'
fi

exit 0
