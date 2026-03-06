#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  preflight.sh --kubeconfig <kubeconfig> --run-dir <run-dir> --repo-root <path>

Options:
  --repo-root <path>  Path to Kuma repository root (required)

Performs:
  - tool availability checks
  - cluster connectivity checks
  - control plane readiness checks
  - local build/kumactl checks
EOF
}

kubeconfig_path=""
run_dir=""
repo_root=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --kubeconfig)
      kubeconfig_path="$2"
      shift 2
      ;;
    --run-dir)
      run_dir="$2"
      shift 2
      ;;
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

if [[ -z "${kubeconfig_path}" || -z "${run_dir}" ]]; then
  usage
  exit 1
fi

if [[ ! -d "${run_dir}" ]]; then
  echo "Error: run directory not found: ${run_dir}" >&2
  exit 1
fi

if [[ ! -f "${kubeconfig_path}" ]]; then
  echo "Error: kubeconfig not found: ${kubeconfig_path}" >&2
  exit 1
fi

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [[ -z "${repo_root}" ]]; then
  repo_root="$(cd "${script_dir}/../../../.." && pwd)"
  printf 'Warning: --repo-root not specified, falling back to %s\n' "${repo_root}" >&2
fi

find_local_kumactl="${script_dir}/find-local-kumactl.sh"
local_kumactl=""

artifacts_dir="${run_dir}/artifacts"
commands_log="${run_dir}/commands/command-log.md"

mkdir -p "${artifacts_dir}"
mkdir -p "${run_dir}/commands"

if [[ ! -f "${commands_log}" ]]; then
  cat >"${commands_log}" <<'EOF'
# command log

| timestamp (utc) | phase | command | exit code | output file |
|---|---|---|---:|---|
EOF
fi

append_command_log() {
  local now="$1"
  local phase="$2"
  local command_text="$3"
  local exit_code="$4"
  local output_file="$5"

  printf "| %s | %s | \`%s\` | %s | \`%s\` |\\n" \
    "${now}" "${phase}" "${command_text}" "${exit_code}" "${output_file}" \
    >>"${commands_log}"
}

failure_count=0

run_check() {
  local name="$1"
  shift
  local -a cmd=("$@")
  local output_rel="artifacts/preflight-${name}.log"
  local output_file="${run_dir}/${output_rel}"
  local cmd_string
  local now
  local exit_code

  cmd_string="$(printf '%q ' "${cmd[@]}")"
  now="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

  set +e
  "${cmd[@]}" >"${output_file}" 2>&1
  exit_code=$?
  set -e

  append_command_log "${now}" "preflight" "${cmd_string}" "${exit_code}" "${output_rel}"

  if [[ ${exit_code} -ne 0 ]]; then
    failure_count=$((failure_count + 1))
    printf 'Preflight check failed: %s (see %s)\n' "${name}" "${output_file}" >&2
  fi
}

for tool in docker k3d kubectl helm make; do
  tool_log="${run_dir}/artifacts/preflight-tool-${tool}.log"
  now="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

  if command -v "${tool}" >"${tool_log}" 2>&1; then
    append_command_log "${now}" "preflight" "command -v ${tool}" "0" "artifacts/preflight-tool-${tool}.log"
  else
    append_command_log "${now}" "preflight" "command -v ${tool}" "1" "artifacts/preflight-tool-${tool}.log"
    failure_count=$((failure_count + 1))
    printf 'Preflight tool check failed: %s\n' "${tool}" >&2
  fi
done

run_check "docker-info" docker info
run_check "cluster-info" kubectl --kubeconfig "${kubeconfig_path}" cluster-info
run_check "nodes" kubectl --kubeconfig "${kubeconfig_path}" get nodes --output=wide
run_check "kuma-system-pods" kubectl --kubeconfig "${kubeconfig_path}" get pods --namespace kuma-system --output=wide
run_check "kuma-control-plane-deployment" kubectl --kubeconfig "${kubeconfig_path}" get deployment --namespace kuma-system kuma-control-plane
run_check "kuma-control-plane-ready" kubectl --kubeconfig "${kubeconfig_path}" wait --for=condition=available deployment/kuma-control-plane --namespace kuma-system --timeout=180s

if local_kumactl="$(${find_local_kumactl} --repo-root "${repo_root}" 2>/dev/null)"; then
  run_check "local-kumactl-version" "${local_kumactl}" version
else
  kumactl_log="${run_dir}/artifacts/preflight-local-kumactl.log"
  "${find_local_kumactl}" --repo-root "${repo_root}" >"${kumactl_log}" 2>&1 || true
  append_command_log \
    "$(date -u +"%Y-%m-%dT%H:%M:%SZ")" \
    "preflight" \
    "${find_local_kumactl} --repo-root ${repo_root}" \
    "1" \
    "artifacts/preflight-local-kumactl.log"
  failure_count=$((failure_count + 1))
  printf 'Preflight failed: build local kumactl first with make build/kumactl\n' >&2
fi

if command -v kumactl >/dev/null 2>&1; then
  system_kumactl_path="$(command -v kumactl)"
  warn_log="${run_dir}/artifacts/preflight-system-kumactl-warning.log"
  printf 'System kumactl path: %s\nLocal kumactl path: %s\n' \
    "${system_kumactl_path}" "${local_kumactl}" >"${warn_log}"

  append_command_log \
    "$(date -u +"%Y-%m-%dT%H:%M:%SZ")" \
    "preflight" \
    "command -v kumactl" \
    "0" \
    "artifacts/preflight-system-kumactl-warning.log"
fi

if [[ ${failure_count} -gt 0 ]]; then
  printf 'Preflight failed with %s issue(s). Fix and rerun.\n' "${failure_count}" >&2
  exit 1
fi

echo "Preflight passed."
