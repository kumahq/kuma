#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  report-compactness-check.sh \
    --report <path> \
    [--max-lines <n>] \
    [--max-line-length <n>] \
    [--max-code-blocks <n>] \
    [--max-code-block-lines <n>]

Defaults:
  --max-lines 220
  --max-line-length 220
  --max-code-blocks 4
  --max-code-block-lines 30

Purpose:
  Fails when a report gets bloated.
EOF
}

report_path=""
max_lines=220
max_line_length=220
max_code_blocks=4
max_code_block_lines=30

while [[ $# -gt 0 ]]; do
  case "$1" in
    --report)
      report_path="$2"
      shift 2
      ;;
    --max-lines)
      max_lines="$2"
      shift 2
      ;;
    --max-line-length)
      max_line_length="$2"
      shift 2
      ;;
    --max-code-blocks)
      max_code_blocks="$2"
      shift 2
      ;;
    --max-code-block-lines)
      max_code_block_lines="$2"
      shift 2
      ;;
    --help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage
      exit 1
      ;;
  esac
done

if [[ -z "${report_path}" ]]; then
  usage
  exit 1
fi

if [[ ! -f "${report_path}" ]]; then
  echo "Error: report file not found: ${report_path}" >&2
  exit 1
fi

validate_positive_int() {
  local name="$1"
  local value="$2"

  if [[ ! "${value}" =~ ^[0-9]+$ ]] || (( value < 1 )); then
    echo "Error: ${name} must be a positive integer, got: ${value}" >&2
    exit 1
  fi
}

validate_positive_int "max-lines" "${max_lines}"
validate_positive_int "max-line-length" "${max_line_length}"
validate_positive_int "max-code-blocks" "${max_code_blocks}"
validate_positive_int "max-code-block-lines" "${max_code_block_lines}"

line_count="$(wc -l < "${report_path}" | tr -d ' ')"
if (( line_count > max_lines )); then
  echo "FAIL: report has ${line_count} lines (max ${max_lines})" >&2
  exit 1
fi

long_line_count="$(awk -v limit="${max_line_length}" 'length($0) > limit {count++} END {print count+0}' "${report_path}")"
if (( long_line_count > 0 )); then
  echo "FAIL: report has ${long_line_count} line(s) longer than ${max_line_length} chars" >&2
  exit 1
fi

code_analysis="$(awk -v max_lines="${max_code_block_lines}" '
  BEGIN {
    in_block = 0
    block_count = 0
    current_lines = 0
    oversized_blocks = 0
  }
  /^```/ {
    if (in_block == 0) {
      in_block = 1
      block_count++
      current_lines = 0
    } else {
      if (current_lines > max_lines) {
        oversized_blocks++
      }
      in_block = 0
      current_lines = 0
    }
    next
  }
  {
    if (in_block == 1) {
      current_lines++
    }
  }
  END {
    if (in_block == 1 && current_lines > max_lines) {
      oversized_blocks++
    }
    printf "%d %d\n", block_count, oversized_blocks
  }
' "${report_path}")"

code_block_count="${code_analysis%% *}"
oversized_code_blocks="${code_analysis##* }"

if (( code_block_count > max_code_blocks )); then
  echo "FAIL: report has ${code_block_count} code blocks (max ${max_code_blocks})" >&2
  exit 1
fi

if (( oversized_code_blocks > 0 )); then
  echo "FAIL: report has ${oversized_code_blocks} oversized code block(s)" >&2
  exit 1
fi

echo "PASS: report compactness check passed"
