#!/usr/bin/env bash

should_log() {
  case "${LOG_QUIET:-}" in
    true|1|on|enabled) return 1 ;;
  esac
  return 0
}

is_verbose() {
  case "${LOG_VERBOSE:-}" in
    true|1|on|enabled) return 0 ;;
  esac
  return 1
}

log::info() {
  should_log || return 0
  echo "[INFO]$(is_verbose && echo '   ') $*" >&2
}

log::verbose() {
  should_log || return 0
  is_verbose || return 0
  echo "[VERBOSE] $*" >&2
}
