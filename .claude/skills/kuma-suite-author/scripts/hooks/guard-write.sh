#!/usr/bin/env bash
# PreToolUse/Write hook for kuma-suite-author
# S2: warn on YAML writes outside expected suite directories
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/_lib.sh"

input="$(cat)"
file_path="$(printf '%s' "${input}" | jq -r '.tool_input.file_path // ""')"

if [[ -z "${file_path}" ]]; then
  exit 0
fi

# Only check YAML files
if [[ ! "${file_path}" =~ \.(yaml|yml)$ ]]; then
  exit 0
fi

# Allow writes to recognized paths
if [[ "${file_path}" == */suites/*/baseline/* ]] || \
   [[ "${file_path}" == */suites/*/groups/* ]] || \
   [[ "${file_path}" == */runs/* ]] || \
   [[ "${file_path}" == */state/* ]]; then
  exit 0
fi

# Warn on unexpected YAML write location
jq -nc '{
  hookSpecificOutput: {
    hookEventName: "PreToolUse",
    permissionDecision: "allow",
    additionalContext: "kuma-suite-author warning: [KSA002] YAML write outside expected suite directory. Expected path: suites/<name>/baseline/*.yaml"
  }
}'
exit 0
