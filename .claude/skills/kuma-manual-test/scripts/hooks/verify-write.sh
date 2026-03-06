#!/usr/bin/env bash
# PostToolUse/Write hook for kuma-manual-test
# Combines: M6 (run-status.yaml fields), M7 (manifest outside manifests/), M15 (report compactness)
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/_lib.sh"

input="$(cat)"
file_path="$(printf '%s' "${input}" | jq -r '.tool_input.file_path // ""')"

if [[ -z "${file_path}" ]]; then
  exit 0
fi

warnings=()

# M6: run-status.yaml missing required fields
if [[ "${file_path}" == *run-status.yaml ]]; then
  if [[ -f "${file_path}" ]]; then
    missing=()
    for field in last_completed_group next_planned_group counts last_updated_utc; do
      if ! grep -qF "${field}" "${file_path}" 2>/dev/null; then
        missing+=("${field}")
      fi
    done
    if [[ ${#missing[@]} -gt 0 ]]; then
      joined="$(IFS=', '; echo "${missing[*]}")"
      warnings+=("[KMT006] run-status.yaml missing fields: ${joined}. Hard gate requires all four fields before proceeding to next group.")
    fi
  fi
fi

# M7: manifest outside manifests/, state/, baseline/
if [[ "${file_path}" =~ \.(yaml|yml)$ ]] && [[ "${file_path}" == */runs/* ]]; then
  if [[ "${file_path}" != */manifests/* ]] && \
     [[ "${file_path}" != */state/* ]] && \
     [[ "${file_path}" != */baseline/* ]] && \
     [[ "${file_path}" != *run-status.yaml ]] && \
     [[ "${file_path}" != *run-metadata.yaml ]]; then
    warnings+=("[KMT007] YAML file written outside manifests/, state/, or baseline/ directory. Check if this manifest belongs in manifests/.")
  fi
fi

# M15: report approaching compactness limits
if [[ "${file_path}" == *manual-test-report.md ]]; then
  if [[ -f "${file_path}" ]]; then
    line_count="$(wc -l < "${file_path}" | tr -d ' ')"
    code_blocks="$(grep -c '```' "${file_path}" 2>/dev/null || echo 0)"
    # Each pair of ``` is one block, divide by 2
    block_count=$(( code_blocks / 2 ))

    if [[ "${line_count}" -gt 176 ]]; then
      warnings+=("[KMT015] Report has ${line_count} lines (80% threshold: 176 of 220). Move raw output to artifacts/ and reference file paths.")
    fi
    if [[ "${block_count}" -gt 3 ]]; then
      warnings+=("[KMT015] Report has ${block_count} code blocks (80% threshold: 3 of 4). Consolidate or move to artifacts/.")
    fi
  fi
fi

if [[ ${#warnings[@]} -gt 0 ]]; then
  joined="$(IFS=' '; echo "${warnings[*]}")"
  jq -nc --arg msg "kuma-manual-test check: ${joined}" '{
    systemMessage: $msg
  }'
fi

exit 0
