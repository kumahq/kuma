#!/usr/bin/env bash
# Shared library for kuma-manual-test hook scripts
# Source this at the top of every hook: source "$(dirname "${BASH_SOURCE[0]}")/_lib.sh"

# Required tools - hooks cannot function without these
_hooks_check_deps() {
  local missing=()

  if ! command -v jq >/dev/null 2>&1; then
    missing+=("jq")
  fi

  if [[ ${#missing[@]} -gt 0 ]]; then
    local joined
    joined="$(IFS=', '; echo "${missing[*]}")"
    # Emit warning and exit cleanly - don't block operations when deps are missing
    printf '{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow","additionalContext":"kuma-manual-test hook warning: missing required tools: %s. Hook checks are disabled until these are installed."}}\n' "${joined}" >&2
    exit 0
  fi
}

_hooks_check_deps
