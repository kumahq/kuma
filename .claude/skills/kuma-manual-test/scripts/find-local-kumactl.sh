#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  find-local-kumactl.sh [--repo-root <path>]

Prints path to locally built `kumactl` binary.
Search order:
  1) build/kumactl
  2) build/artifacts-*/kumactl/kumactl
EOF
}

repo_root=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --repo-root)
      repo_root="$2"
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

if [[ -z "${repo_root}" ]]; then
  script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  repo_root="$(cd "${script_dir}/../../../.." && pwd)"
fi

candidate_1="${repo_root}/build/kumactl"
if [[ -x "${candidate_1}" ]]; then
  printf '%s\n' "${candidate_1}"
  exit 0
fi

for candidate in "${repo_root}"/build/artifacts-*/kumactl/kumactl; do
  if [[ -x "${candidate}" ]]; then
    printf '%s\n' "${candidate}"
    exit 0
  fi
done

echo "Error: local kumactl not found in build directory. Run: make build/kumactl" >&2
exit 1
