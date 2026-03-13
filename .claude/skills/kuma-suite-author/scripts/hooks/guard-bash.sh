#!/usr/bin/env bash
# PreToolUse/Bash hook for kuma-suite-author
# S1: block generic suite names
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/_lib.sh"

input="$(cat)"
command="$(printf '%s' "${input}" | jq -r '.tool_input.command // ""')"

if [[ -z "${command}" ]]; then
  exit 0
fi

# S1: detect mkdir with suites/ path and check name against reject list
if printf '%s' "${command}" | grep -qE 'mkdir.*suites/'; then
  # Extract the suite name (last path component under suites/)
  suite_name="$(printf '%s' "${command}" | grep -oE 'suites/[^/[:space:]]+' | head -1 | sed 's|suites/||')"

  if [[ -n "${suite_name}" ]]; then
    # Reject list: generic names
    if printf '%s' "${suite_name}" | grep -qxE '(test-suite-[0-9]+|full|feature-branch|my-test|test|suite)'; then
      jq -nc '{
        hookSpecificOutput: {
          hookEventName: "PreToolUse",
          permissionDecision: "deny",
          permissionDecisionReason: "[KSA001] Generic suite name detected. Use {feature}-{scope} pattern (e.g., meshretry-core, motb-pipe-mode).",
          additionalContext: "Automated kuma-suite-author hook. Choose a descriptive suite name and retry."
        },
        systemMessage: "\nKSA001: Generic suite name blocked\n  Fix: Use {feature}-{scope} pattern\n"
      }'
      exit 0
    fi
  fi
fi

# Clean pass
exit 0
