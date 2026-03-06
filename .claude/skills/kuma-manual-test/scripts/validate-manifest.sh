#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  validate-manifest.sh --kubeconfig <path> --manifest <file> [--output <file>]

Checks:
  1) kinds exist in API server
  2) server-side dry-run apply
  3) kubectl diff sanity check
EOF
}

kubeconfig_path=""
manifest_path=""
output_file=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --kubeconfig)
      kubeconfig_path="$2"
      shift 2
      ;;
    --manifest)
      manifest_path="$2"
      shift 2
      ;;
    --output)
      output_file="$2"
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

if [[ -z "${kubeconfig_path}" || -z "${manifest_path}" ]]; then
  usage
  exit 1
fi

if [[ ! -f "${kubeconfig_path}" ]]; then
  echo "Error: kubeconfig not found: ${kubeconfig_path}" >&2
  exit 1
fi

if [[ ! -f "${manifest_path}" ]]; then
  echo "Error: manifest not found: ${manifest_path}" >&2
  exit 1
fi

if [[ -n "${output_file}" ]]; then
  mkdir -p "$(dirname "${output_file}")"
  : >"${output_file}"
fi

log() {
  local message="$1"

  if [[ -n "${output_file}" ]]; then
    printf '%s\n' "${message}" | tee -a "${output_file}"
  else
    printf '%s\n' "${message}"
  fi
}

run_logged() {
  local -a cmd=("$@")
  local cmd_string
  cmd_string="$(printf '%q ' ${cmd[@]+"${cmd[@]}"})"
  log "\$ ${cmd_string}"

  if [[ -n "${output_file}" ]]; then
    ${cmd[@]+"${cmd[@]}"} 2>&1 | tee -a "${output_file}"
  else
    ${cmd[@]+"${cmd[@]}"}
  fi
}

log "Validating manifest: ${manifest_path}"
log "Using kubeconfig: ${kubeconfig_path}"

mapfile -t kinds < <(
  awk '/^[[:space:]]*kind:[[:space:]]*/ { print $2 }' "${manifest_path}" | sort -u
)

if [[ ${#kinds[@]} -eq 0 ]]; then
  log "Error: no kind fields found in manifest"
  exit 1
fi

for kind in ${kinds[@]+"${kinds[@]}"}; do
  if [[ -z "${kind}" || "${kind}" == "List" ]]; then
    continue
  fi

  log "\$ kubectl --kubeconfig ${kubeconfig_path} explain ${kind}"
  if ! kubectl --kubeconfig "${kubeconfig_path}" explain "${kind}" >/dev/null 2>&1; then
    log "Error: kind not found in API server: ${kind}"
    exit 1
  fi
done

run_logged \
  kubectl --kubeconfig "${kubeconfig_path}" \
  apply --server-side --dry-run=server --filename "${manifest_path}"

set +e
if [[ -n "${output_file}" ]]; then
  kubectl --kubeconfig "${kubeconfig_path}" diff --filename "${manifest_path}" \
    2>&1 | tee -a "${output_file}"
  diff_exit=${PIPESTATUS[0]}
else
  kubectl --kubeconfig "${kubeconfig_path}" diff --filename "${manifest_path}"
  diff_exit=$?
fi
set -e

if [[ ${diff_exit} -gt 1 ]]; then
  log "kubectl diff failed with exit code ${diff_exit}"
  exit "${diff_exit}"
fi

log "Validation completed successfully."
