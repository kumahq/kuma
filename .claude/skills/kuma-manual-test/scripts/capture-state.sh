#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  capture-state.sh \
    --run-dir <run-dir> \
    --kubeconfig <kubeconfig> \
    [--label <label>] \
    [--resource <resource>]...

Examples:
  capture-state.sh --run-dir runs/20260304 --kubeconfig ~/.kube/kind-kuma-1-config
  capture-state.sh --run-dir runs/20260304 --kubeconfig ~/.kube/kind-kuma-1-config --label failure-g3
  capture-state.sh --run-dir runs/20260304 --kubeconfig ~/.kube/kind-kuma-1-config --resource meshopentelemetrybackends
EOF
}

run_dir=""
kubeconfig_path=""
label="snapshot"
extra_resources=()

while [[ $# -gt 0 ]]; do
  case "$1" in
    --run-dir)
      run_dir="$2"
      shift 2
      ;;
    --kubeconfig)
      kubeconfig_path="$2"
      shift 2
      ;;
    --label)
      label="$2"
      shift 2
      ;;
    --resource)
      extra_resources+=("$2")
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

if [[ -z "${run_dir}" || -z "${kubeconfig_path}" ]]; then
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

state_dir="${run_dir}/state"
commands_log="${run_dir}/commands/command-log.md"

mkdir -p "${state_dir}"
mkdir -p "${run_dir}/commands"

if [[ ! -f "${commands_log}" ]]; then
  cat >"${commands_log}" <<'EOF'
# command log

| timestamp (utc) | phase | command | exit code | output file |
|---|---|---|---:|---|
EOF
fi

timestamp="$(date -u +"%Y%m%dT%H%M%SZ")"

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

capture() {
  local name="$1"
  shift
  local -a cmd=("$@")
  local output_rel="state/${timestamp}-${label}-${name}.log"
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

  append_command_log "${now}" "state-capture" "${cmd_string}" "${exit_code}" "${output_rel}"

  if [[ ${exit_code} -ne 0 ]]; then
    printf 'Warning: capture failed (%s), see %s\n' "${name}" "${output_file}" >&2
  fi
}

capture "cluster-info" kubectl --kubeconfig "${kubeconfig_path}" cluster-info
capture "nodes" kubectl --kubeconfig "${kubeconfig_path}" get nodes --output=wide
capture "pods-all" kubectl --kubeconfig "${kubeconfig_path}" get pods --all-namespaces --output=wide
capture "services-all" kubectl --kubeconfig "${kubeconfig_path}" get services --all-namespaces
capture "events-all" kubectl --kubeconfig "${kubeconfig_path}" get events --all-namespaces --sort-by=.lastTimestamp
capture "kuma-system-pods" kubectl --kubeconfig "${kubeconfig_path}" get pods --namespace kuma-system --output=wide
capture "kuma-control-plane-logs" kubectl --kubeconfig "${kubeconfig_path}" logs --namespace kuma-system deploy/kuma-control-plane --tail=400
capture "meshes" kubectl --kubeconfig "${kubeconfig_path}" get meshes --output=yaml
capture "dataplanes" kubectl --kubeconfig "${kubeconfig_path}" get dataplanes --all-namespaces --output=yaml
capture "dataplaneinsights" kubectl --kubeconfig "${kubeconfig_path}" get dataplaneinsights --all-namespaces --output=yaml
capture "zones" kubectl --kubeconfig "${kubeconfig_path}" get zones --all-namespaces --output=yaml
capture "zoneinsights" kubectl --kubeconfig "${kubeconfig_path}" get zoneinsights --all-namespaces --output=yaml
capture "zoneingresses" kubectl --kubeconfig "${kubeconfig_path}" get zoneingresses --all-namespaces --output=yaml
capture "zoneegresses" kubectl --kubeconfig "${kubeconfig_path}" get zoneegresses --all-namespaces --output=yaml

for resource in "${extra_resources[@]}"; do
  safe_name="$(printf "%s" "${resource}" | tr '/.' '--')"
  capture "resource-${safe_name}" kubectl --kubeconfig "${kubeconfig_path}" get "${resource}" --all-namespaces --output=yaml
done

echo "State snapshot complete: label=${label}, timestamp=${timestamp}"
