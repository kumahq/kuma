#!/usr/bin/env bash
# PostToolUse/Write hook for kuma-suite-author
# Combines: S3 (YAML syntax), S4 (incomplete suite.md), S5 (incomplete group),
#           S9 (oversized group), S10 (missing session provenance)
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/_lib.sh"

input="$(cat)"
file_path="$(printf '%s' "${input}" | jq -r '.tool_input.file_path // ""')"

if [[ -z "${file_path}" ]]; then
  exit 0
fi

if [[ ! -f "${file_path}" ]]; then
  exit 0
fi

warnings=()

# S3: baseline YAML syntax validation
if [[ "${file_path}" == */baseline/*.yaml ]] || [[ "${file_path}" == */baseline/*.yml ]]; then
  yaml_error="$(python3 -c "import sys; import yaml; yaml.safe_load(open(sys.argv[1]))" "${file_path}" 2>&1)" || true
  if [[ -n "${yaml_error}" ]] && printf '%s' "${yaml_error}" | grep -qiE 'error|exception|invalid'; then
    warnings+=("[KSA003] Generated YAML has syntax errors. Fix before saving the suite.")
  fi
fi

# S4: suite.md missing required sections
if [[ "${file_path}" == */suites/*/suite.md ]]; then
  missing_sections=()
  for section in "Baseline" "Groups" "Execution contract" "Metadata"; do
    if ! grep -qiF "${section}" "${file_path}" 2>/dev/null; then
      missing_sections+=("${section}")
    fi
  done
  if [[ ${#missing_sections[@]} -gt 0 ]]; then
    joined="$(IFS=', '; echo "${missing_sections[*]}")"
    warnings+=("[KSA004] suite.md missing sections: ${joined}. All four sections required.")
  fi

  # S10: missing session provenance
  if ! grep -q 'session_id' "${file_path}" 2>/dev/null; then
    warnings+=("[KSA010] suite.md missing session_id in metadata. Add for provenance tracking.")
  fi
fi

# S5: incomplete group file structure
if [[ "${file_path}" == */groups/g*.md ]]; then
  missing_structure=()
  if ! grep -qE '^#' "${file_path}" 2>/dev/null; then
    missing_structure+=("heading")
  fi
  if ! grep -qiE '(steps|procedure|instructions)' "${file_path}" 2>/dev/null; then
    missing_structure+=("steps section")
  fi
  if ! grep -qiE '(artifacts|captures|outputs)' "${file_path}" 2>/dev/null; then
    missing_structure+=("artifacts section")
  fi
  if [[ ${#missing_structure[@]} -gt 0 ]]; then
    joined="$(IFS=', '; echo "${missing_structure[*]}")"
    warnings+=("[KSA005] Group file missing structure: ${joined}.")
  fi

  # S9: oversized group file
  line_count="$(wc -l < "${file_path}" | tr -d ' ')"
  if [[ "${line_count}" -gt 100 ]]; then
    warnings+=("[KSA009] Group file exceeds 100 lines (${line_count} lines). Consider splitting.")
  fi
fi

if [[ ${#warnings[@]} -gt 0 ]]; then
  joined="$(IFS=' '; echo "${warnings[*]}")"
  jq -nc --arg msg "kuma-suite-author checks: ${joined}" '{
    systemMessage: $msg
  }'
fi

exit 0
