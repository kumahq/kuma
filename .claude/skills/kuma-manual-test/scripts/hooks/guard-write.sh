#!/usr/bin/env bash
# PreToolUse/Write hook for kuma-manual-test
# M3: block manifest writes to /tmp
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/_lib.sh"

input="$(cat)"
file_path="$(printf '%s' "${input}" | jq -r '.tool_input.file_path // ""')"

if [[ -z "${file_path}" ]]; then
  exit 0
fi

# M3: manifest in /tmp
if [[ "${file_path}" =~ ^/tmp/.*\.(yaml|yml)$ ]]; then
  jq -nc '{
    hookSpecificOutput: {
      hookEventName: "PreToolUse",
      permissionDecision: "deny",
      permissionDecisionReason: "[KMT003] Manifest write to /tmp detected. Write manifests to ${RUN_DIR}/manifests/ for tracking and reproducibility.",
      additionalContext: "Automated kuma-manual-test hook. Write the manifest to the run directory instead."
    },
    systemMessage: "\nKMT003: /tmp manifest write blocked\n  Fix: Write to ${RUN_DIR}/manifests/ instead\n"
  }'
  exit 0
fi

# Clean pass
exit 0
