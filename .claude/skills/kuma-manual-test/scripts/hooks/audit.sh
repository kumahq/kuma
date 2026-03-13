#!/usr/bin/env bash
# PostToolUse + PostToolUseFailure audit hook for kuma-manual-test
# Logs every Bash command and Write operation to .audit.jsonl
# MA1: post_bash, MA2: post_write, MA3: bash_failure
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/_lib.sh"

input="$(cat)"
event="$(printf '%s' "${input}" | jq -r '.hook_event_name // ""')"
tool="$(printf '%s' "${input}" | jq -r '.tool_name // ""')"
session="$(printf '%s' "${input}" | jq -r '.session_id // "unknown"')"

audit_dir="${XDG_DATA_HOME:-$HOME/.local/share}/kuma/kuma-manual-test"
mkdir -p "${audit_dir}"
audit_file="${audit_dir}/.audit.jsonl"

ts="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

case "${event}/${tool}" in
  PostToolUse/Bash)
    cmd="$(printf '%s' "${input}" | jq -r '.tool_input.command // ""')"
    ec="$(printf '%s' "${input}" | jq -r '.tool_response.exit_code // 0')"
    jq -nc \
      --arg ts "${ts}" \
      --arg cmd "${cmd}" \
      --arg ec "${ec}" \
      --arg session "${session}" \
      '{ts: $ts, event: "post_bash", cmd: $cmd, exit_code: ($ec | tonumber), session: $session}' \
      >> "${audit_file}"
    ;;
  PostToolUse/Write)
    path="$(printf '%s' "${input}" | jq -r '.tool_input.file_path // ""')"
    size=0
    if [[ -f "${path}" ]]; then
      size="$(wc -c < "${path}" | tr -d ' ')"
    fi
    jq -nc \
      --arg ts "${ts}" \
      --arg path "${path}" \
      --arg size "${size}" \
      --arg session "${session}" \
      '{ts: $ts, event: "post_write", path: $path, size: ($size | tonumber), session: $session}' \
      >> "${audit_file}"
    ;;
  PostToolUseFailure/Bash)
    cmd="$(printf '%s' "${input}" | jq -r '.tool_input.command // ""')"
    error="$(printf '%s' "${input}" | jq -r '.error // ""')"
    interrupt="$(printf '%s' "${input}" | jq -r '.is_interrupt // false')"
    jq -nc \
      --arg ts "${ts}" \
      --arg cmd "${cmd}" \
      --arg error "${error}" \
      --argjson interrupt "${interrupt}" \
      --arg session "${session}" \
      '{ts: $ts, event: "bash_failure", cmd: $cmd, error: $error, interrupt: $interrupt, session: $session}' \
      >> "${audit_file}"
    ;;
esac

# Suppress output - pure logging
printf '{"suppressOutput":true}\n'
exit 0
