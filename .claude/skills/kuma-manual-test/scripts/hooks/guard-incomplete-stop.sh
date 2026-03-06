#!/usr/bin/env bash
# Stop hook for kuma-manual-test
# M14: block stop if active run has missing closeout gates
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/_lib.sh"

input="$(cat)"

# Prevent infinite loop
stop_active="$(printf '%s' "${input}" | jq -r '.stop_hook_active // false')"
if [[ "${stop_active}" == "true" ]]; then
  exit 0
fi

data_dir="${XDG_DATA_HOME:-$HOME/.local/share}/sai/kuma-manual-test"
runs_dir="${data_dir}/runs"
current_run_file="${runs_dir}/.current-run"

if [[ ! -f "${current_run_file}" ]]; then
  exit 0
fi

run_id="$(cat "${current_run_file}" | tr -d '[:space:]')"
if [[ -z "${run_id}" ]]; then
  exit 0
fi

run_dir="${runs_dir}/${run_id}"
if [[ ! -d "${run_dir}" ]]; then
  # Run dir doesn't exist - stale pointer, clean up
  rm -f "${current_run_file}"
  exit 0
fi

status_file="${run_dir}/run-status.yaml"
state_dir="${run_dir}/state"

# Check if run has started (has completed groups)
has_completed=false
if [[ -f "${status_file}" ]] && grep -q 'last_completed_group' "${status_file}" 2>/dev/null; then
  last="$(grep 'last_completed_group' "${status_file}" | head -1 | awk '{print $2}')"
  if [[ -n "${last}" ]] && [[ "${last}" != "~" ]] && [[ "${last}" != "null" ]]; then
    has_completed=true
  fi
fi

if [[ "${has_completed}" == "false" ]]; then
  # Run exists but no groups completed - allow stop
  exit 0
fi

# Check for postrun state capture
has_postrun=false
if [[ -d "${state_dir}" ]]; then
  postrun_count="$(find "${state_dir}" -name "*postrun*" -type f 2>/dev/null | wc -l | tr -d ' ')"
  if [[ "${postrun_count}" -gt 0 ]]; then
    has_postrun=true
  fi
fi

if [[ "${has_postrun}" == "false" ]]; then
  echo "[KMT014] Run ${run_id} is incomplete. Missing: postrun state capture. Complete Phase 6 closeout before stopping." >&2
  exit 2
fi

# Run properly closed out - clean up pointer
rm -f "${current_run_file}"
exit 0
