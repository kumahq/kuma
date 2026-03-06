#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  record-command.sh --run-dir <run-dir> --phase <phase> --label <label> -- <command...>

Example:
  record-command.sh \
    --run-dir tmp/investigations/manual-test-harness/runs/20260304-120000 \
    --phase verify \
    --label list-pods \
    -- kubectl --kubeconfig ~/.kube/kind-kuma-1-config get pods -A
EOF
}

run_dir=""
phase="misc"
label="command"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --run-dir)
      run_dir="$2"
      shift 2
      ;;
    --phase)
      phase="$2"
      shift 2
      ;;
    --label)
      label="$2"
      shift 2
      ;;
    --help)
      usage
      exit 0
      ;;
    --)
      shift
      break
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage
      exit 1
      ;;
  esac
done

if [[ -z "${run_dir}" || $# -eq 0 ]]; then
  usage
  exit 1
fi

if [[ ! -d "${run_dir}" ]]; then
  echo "Error: run directory not found: ${run_dir}" >&2
  exit 1
fi

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

safe_label="$(printf "%s" "${label}" | tr '[:space:]' '-' | tr -cd 'a-zA-Z0-9._-')"
if [[ -z "${safe_label}" ]]; then
  safe_label="command"
fi

timestamp="$(date -u +"%Y%m%dT%H%M%SZ")"
output_rel="artifacts/${timestamp}-${safe_label}.log"
output_file="${run_dir}/${output_rel}"

cmd_string="$(printf '%q ' "$@")"

set +e
"$@" >"${output_file}" 2>&1
exit_code=$?
set -e

printf "| %s | %s | \`%s\` | %s | \`%s\` |\\n" \
  "$(date -u +"%Y-%m-%dT%H:%M:%SZ")" \
  "${phase}" \
  "${cmd_string}" \
  "${exit_code}" \
  "${output_rel}" \
  >>"${commands_log}"

cat <<EOF
Command recorded:
  command: ${cmd_string}
  output:  ${output_file}
  exit:    ${exit_code}
EOF

exit "${exit_code}"
