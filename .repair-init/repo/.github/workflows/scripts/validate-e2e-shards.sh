#!/usr/bin/env bash

set -euo pipefail

target="${1:-}"
shards="${2:-1}"

if [[ -z "$target" ]]; then
  echo "usage: $0 <target> <shards>" >&2
  exit 1
fi

suite_file="test/e2e_env/$target/${target}_suite_test.go"

if [[ ! -f "$suite_file" ]]; then
  echo "missing sharded e2e suite file: $suite_file" >&2
  exit 1
fi

awk -v shards="$shards" '
/^[[:space:]]*_ = Describe\(/ {
  found = 1

  if (!match($0, /Label\("job-[0-9]+"\)/)) {
    printf "%s:%d: sharded e2e suite registration is missing a job label: %s\n", FILENAME, NR, $0 > "/dev/stderr"
    failed = 1
    next
  }

  job_label = substr($0, RSTART + 11, RLENGTH - 13) + 0
  if (job_label >= shards) {
    printf "%s:%d: sharded e2e suite registration uses job-%d but only %d shard(s) are configured\n", FILENAME, NR, job_label, shards > "/dev/stderr"
    failed = 1
  }
}

END {
  if (!found) {
    printf "%s: no top-level Describe registrations found to shard\n", FILENAME > "/dev/stderr"
    failed = 1
  }

  exit failed
}
' "$suite_file"
