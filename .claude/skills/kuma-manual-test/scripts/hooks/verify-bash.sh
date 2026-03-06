#!/usr/bin/env bash
# PostToolUse/Bash hook for kuma-manual-test
# Combines: M9 (state capture gaps), M10 (missing post-apply artifacts)
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/_lib.sh"

input="$(cat)"
command="$(printf '%s' "${input}" | jq -r '.tool_input.command // ""')"
exit_code="$(printf '%s' "${input}" | jq -r '.tool_response.exit_code // ""')"

if [[ -z "${command}" ]]; then
  exit 0
fi

# Only check on success
if [[ "${exit_code}" != "0" ]]; then
  exit 0
fi

warnings=()

# M9: capture-state.sh produced no files
if printf '%s' "${command}" | grep -qF 'capture-state.sh'; then
  # Extract --run-dir and --label from command
  run_dir="$(printf '%s' "${command}" | grep -oE -- '--run-dir[[:space:]]+[^[:space:]]+' | awk '{print $2}' || true)"
  label="$(printf '%s' "${command}" | grep -oE -- '--label[[:space:]]+[^[:space:]]+' | awk '{print $2}' || true)"

  if [[ -n "${run_dir}" ]] && [[ -n "${label}" ]]; then
    state_dir="${run_dir}/state"
    if [[ -d "${state_dir}" ]]; then
      match_count="$(find "${state_dir}" -name "*${label}*" -type f 2>/dev/null | wc -l | tr -d ' ')"
      if [[ "${match_count}" -eq 0 ]]; then
        warnings+=("[KMT009] capture-state.sh produced no files for label '${label}'. State capture is a hard gate - investigate before proceeding.")
      fi
    fi
  fi
fi

# M10: apply-tracked-manifest.sh succeeded but artifacts missing
if printf '%s' "${command}" | grep -qF 'apply-tracked-manifest.sh'; then
  run_dir="$(printf '%s' "${command}" | grep -oE -- '--run-dir[[:space:]]+[^[:space:]]+' | awk '{print $2}' || true)"
  step="$(printf '%s' "${command}" | grep -oE -- '--step[[:space:]]+[^[:space:]]+' | awk '{print $2}' || true)"

  if [[ -n "${run_dir}" ]] && [[ -n "${step}" ]]; then
    missing=()
    # Check for tracked manifest copy
    if [[ -d "${run_dir}/manifests" ]]; then
      copy_count="$(find "${run_dir}/manifests" -name "*${step}*" -type f 2>/dev/null | wc -l | tr -d ' ')"
      if [[ "${copy_count}" -eq 0 ]]; then
        missing+=("tracked manifest copy")
      fi
    fi
    # Check for apply log
    if [[ -d "${run_dir}/commands" ]]; then
      log_count="$(find "${run_dir}/commands" -name "*apply*" -newer "${run_dir}/run-metadata.yaml" -type f 2>/dev/null | wc -l | tr -d ' ')"
      if [[ "${log_count}" -eq 0 ]]; then
        missing+=("apply log")
      fi
    fi
    if [[ ${#missing[@]} -gt 0 ]]; then
      joined="$(IFS=', '; echo "${missing[*]}")"
      warnings+=("[KMT010] apply-tracked-manifest.sh completed but missing: ${joined}. Check the run directory.")
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
