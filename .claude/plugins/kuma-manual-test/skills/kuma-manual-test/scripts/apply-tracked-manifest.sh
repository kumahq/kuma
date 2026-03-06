#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  apply-tracked-manifest.sh \
    --run-dir <run-dir> \
    --kubeconfig <kubeconfig> \
    --manifest <source-manifest> \
    --step <step-name>
EOF
}

run_dir=""
kubeconfig_path=""
source_manifest=""
step_name=""

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
    --manifest)
      source_manifest="$2"
      shift 2
      ;;
    --step)
      step_name="$2"
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

if [[ -z "${run_dir}" || -z "${kubeconfig_path}" || -z "${source_manifest}" || -z "${step_name}" ]]; then
  usage
  exit 1
fi

if [[ ! -d "${run_dir}" ]]; then
  echo "Error: run directory does not exist: ${run_dir}" >&2
  exit 1
fi

if [[ ! -f "${kubeconfig_path}" ]]; then
  echo "Error: kubeconfig does not exist: ${kubeconfig_path}" >&2
  exit 1
fi

if [[ ! -f "${source_manifest}" ]]; then
  echo "Error: source manifest does not exist: ${source_manifest}" >&2
  exit 1
fi

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
validate_script="${script_dir}/validate-manifest.sh"

manifests_dir="${run_dir}/manifests"
artifacts_dir="${run_dir}/artifacts"
commands_log="${run_dir}/commands/command-log.md"
manifest_index="${manifests_dir}/manifest-index.md"

mkdir -p "${manifests_dir}"
mkdir -p "${artifacts_dir}"
mkdir -p "${run_dir}/commands"

if [[ ! -f "${commands_log}" ]]; then
  cat >"${commands_log}" <<'EOF'
# command log

| timestamp (utc) | phase | command | exit code | output file |
|---|---|---|---:|---|
EOF
fi

if [[ ! -f "${manifest_index}" ]]; then
  cat >"${manifest_index}" <<'EOF'
# manifest index

| id | file | purpose | sha256 | source | validated | applied | timestamp (utc) |
|---:|---|---|---|---|---|---|---|
EOF
fi

next_sequence() {
  local last=0
  local file_name
  local base_name
  local seq

  for file_name in "${manifests_dir}"/*.yaml "${manifests_dir}"/*.yml; do
    [[ -e "${file_name}" ]] || continue
    base_name="$(basename "${file_name}")"
    seq="${base_name%%-*}"

    if [[ "${seq}" =~ ^[0-9]{3}$ ]] && (( 10#${seq} > last )); then
      last=$((10#${seq}))
    fi
  done

  printf "%03d" "$((last + 1))"
}

sanitize_step() {
  printf "%s" "$1" | tr '[:space:]' '-' | tr -cd 'a-zA-Z0-9._-'
}

append_command_log() {
  local timestamp="$1"
  local phase="$2"
  local command_text="$3"
  local exit_code="$4"
  local output_file="$5"

  printf "| %s | %s | \`%s\` | %s | \`%s\` |\\n" \
    "${timestamp}" "${phase}" "${command_text}" "${exit_code}" "${output_file}" \
    >>"${commands_log}"
}

sequence="$(next_sequence)"
step_slug="$(sanitize_step "${step_name}")"

if [[ -z "${step_slug}" ]]; then
  echo "Error: step name became empty after sanitization" >&2
  exit 1
fi

extension="yaml"
if [[ "${source_manifest}" == *.yml ]]; then
  extension="yml"
fi

tracked_manifest="${manifests_dir}/${sequence}-${step_slug}.${extension}"
cp "${source_manifest}" "${tracked_manifest}"

sha256="$(shasum -a 256 "${tracked_manifest}" | awk '{print $1}')"
timestamp="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

validate_log_rel="artifacts/${sequence}-validate-${step_slug}.log"
apply_log_rel="artifacts/${sequence}-apply-${step_slug}.log"
validate_log="${run_dir}/${validate_log_rel}"
apply_log="${run_dir}/${apply_log_rel}"

set +e
"${validate_script}" \
  --kubeconfig "${kubeconfig_path}" \
  --manifest "${tracked_manifest}" \
  --output "${validate_log}"
validate_exit=$?
set -e

validate_cmd="${validate_script} --kubeconfig ${kubeconfig_path} --manifest ${tracked_manifest} --output ${validate_log}"
append_command_log "${timestamp}" "validate" "${validate_cmd}" "${validate_exit}" "${validate_log_rel}"

if [[ ${validate_exit} -ne 0 ]]; then
  printf '| %s | %s | %s | %s | %s | %s | %s | %s |\n' \
    "${sequence}" "$(basename "${tracked_manifest}")" "${step_name}" "${sha256}" \
    "${source_manifest}" "FAIL" "NO" "${timestamp}" \
    >>"${manifest_index}"

  echo "Validation failed. Manifest was not applied." >&2
  exit "${validate_exit}"
fi

apply_cmd=(kubectl --kubeconfig "${kubeconfig_path}" apply --filename "${tracked_manifest}")
apply_cmd_string="$(printf '%q ' "${apply_cmd[@]}")"

set +e
"${apply_cmd[@]}" >"${apply_log}" 2>&1
apply_exit=$?
set -e

append_command_log "${timestamp}" "apply" "${apply_cmd_string}" "${apply_exit}" "${apply_log_rel}"

applied_status="PASS"
if [[ ${apply_exit} -ne 0 ]]; then
  applied_status="FAIL"
fi

printf '| %s | %s | %s | %s | %s | %s | %s | %s |\n' \
  "${sequence}" "$(basename "${tracked_manifest}")" "${step_name}" "${sha256}" \
  "${source_manifest}" "PASS" "${applied_status}" "${timestamp}" \
  >>"${manifest_index}"

if [[ ${apply_exit} -ne 0 ]]; then
  echo "Apply failed. Stop and triage before continuing." >&2
  exit "${apply_exit}"
fi

cat <<EOF
Manifest applied and tracked:
  tracked manifest: ${tracked_manifest}
  validation log:   ${validate_log}
  apply log:        ${apply_log}
EOF
