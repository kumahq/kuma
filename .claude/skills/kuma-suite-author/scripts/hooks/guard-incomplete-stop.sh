#!/usr/bin/env bash
# Stop hook for kuma-suite-author
# S8: block stop if current suite generation is incomplete
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/_lib.sh"

input="$(cat)"

# Prevent infinite loop
stop_active="$(printf '%s' "${input}" | jq -r '.stop_hook_active // false')"
if [[ "${stop_active}" == "true" ]]; then
  exit 0
fi

data_dir="${XDG_DATA_HOME:-$HOME/.local/share}/kuma/kuma-manual-test"
current_suite_file="${data_dir}/suites/.current-suite"

if [[ ! -f "${current_suite_file}" ]]; then
  exit 0
fi

suite_name="$(cat "${current_suite_file}" | tr -d '[:space:]')"
if [[ -z "${suite_name}" ]]; then
  exit 0
fi

suite_dir="${data_dir}/suites/${suite_name}"
if [[ ! -d "${suite_dir}" ]]; then
  # Suite dir doesn't exist - stale pointer, clean up
  rm -f "${current_suite_file}"
  exit 0
fi

missing=()

if [[ ! -f "${suite_dir}/suite.md" ]]; then
  missing+=("suite.md")
fi

baseline_count=0
if [[ -d "${suite_dir}/baseline" ]]; then
  baseline_count="$(find "${suite_dir}/baseline" -name '*.yaml' -o -name '*.yml' 2>/dev/null | wc -l | tr -d ' ')"
fi
if [[ "${baseline_count}" -eq 0 ]]; then
  missing+=("baseline manifests")
fi

group_count=0
if [[ -d "${suite_dir}/groups" ]]; then
  group_count="$(find "${suite_dir}/groups" -name 'g*.md' 2>/dev/null | wc -l | tr -d ' ')"
fi
if [[ "${group_count}" -eq 0 ]]; then
  missing+=("group files")
fi

if [[ ${#missing[@]} -gt 0 ]]; then
  joined="$(IFS=', '; echo "${missing[*]}")"
  echo "[KSA008] Suite '${suite_name}' is incomplete. Missing: ${joined}. Complete Step 8 (save suite) before stopping." >&2
  exit 2
fi

# Suite is complete - clean up pointer
rm -f "${current_suite_file}"
exit 0
