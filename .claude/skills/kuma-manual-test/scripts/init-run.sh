#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  init-run.sh [--runs-dir <path>] [--session-id <id>] <run-id>

Options:
  --runs-dir <path>      Base directory for runs (default: tmp/manual-test-runs)
  --session-id <id>      Claude Code session ID for provenance tracking (default: standalone)

Example:
  init-run.sh 20260304-154500-manual
  init-run.sh --runs-dir tmp/my-runs --session-id abc-123 20260304-154500-manual
EOF
}

runs_dir_override=""
session_id="standalone"
run_id=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --help)
      usage
      exit 0
      ;;
    --runs-dir)
      runs_dir_override="$2"
      shift 2
      ;;
    --session-id)
      session_id="$2"
      shift 2
      ;;
    -*)
      echo "Error: unknown option: $1" >&2
      usage
      exit 1
      ;;
    *)
      if [[ -n "${run_id}" ]]; then
        echo "Error: unexpected argument: $1" >&2
        usage
        exit 1
      fi
      run_id="$1"
      shift
      ;;
  esac
done

if [[ -z "${run_id}" ]]; then
  usage
  exit 1
fi

if [[ "${run_id}" == -* ]]; then
  echo "Error: run-id cannot start with '-'" >&2
  exit 1
fi

if [[ ! "${run_id}" =~ ^[a-zA-Z0-9._-]+$ ]]; then
  echo "Error: run-id contains unsupported characters: ${run_id}" >&2
  exit 1
fi

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
skill_dir="$(cd "${script_dir}/.." && pwd)"
runs_dir="${runs_dir_override:-tmp/manual-test-runs}"
run_dir="${runs_dir}/${run_id}"
assets_dir="${skill_dir}/assets"

if [[ -e "${run_dir}" ]]; then
  echo "Error: run directory already exists: ${run_dir}" >&2
  exit 1
fi

if [[ ! -d "${assets_dir}" ]]; then
  echo "Error: assets directory not found: ${assets_dir}" >&2
  exit 1
fi

mkdir -p "${run_dir}/commands"
mkdir -p "${run_dir}/manifests"
mkdir -p "${run_dir}/artifacts"
mkdir -p "${run_dir}/state"
mkdir -p "${run_dir}/reports"

created_at="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

harness_version="$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")"

render_template() {
  local template_file="$1"
  local output_file="$2"

  sed \
    -e "s/__RUN_ID__/${run_id}/g" \
    -e "s/__SESSION_ID__/${session_id}/g" \
    -e "s/__CREATED_AT__/${created_at}/g" \
    -e "s/__HARNESS_VERSION__/${harness_version}/g" \
    "${template_file}" >"${output_file}"
}

render_template \
  "${assets_dir}/run-metadata.template.yaml" \
  "${run_dir}/run-metadata.yaml"

render_template \
  "${assets_dir}/command-log.template.md" \
  "${run_dir}/commands/command-log.md"

render_template \
  "${assets_dir}/manifest-index.template.md" \
  "${run_dir}/manifests/manifest-index.md"

render_template \
  "${assets_dir}/manual-test-report.template.md" \
  "${run_dir}/reports/manual-test-report.md"

render_template \
  "${assets_dir}/run-status.template.yaml" \
  "${run_dir}/run-status.yaml"

# Write .current-run pointer for stop hook (M14)
echo "${run_id}" > "${runs_dir}/.current-run"

printf 'Created run directory:\n  %s\n\nNext steps:\n  1) Start cluster with scripts/cluster-lifecycle.sh\n  2) Run preflight with scripts/preflight.sh\n  3) Apply manifests with scripts/apply-tracked-manifest.sh\n  4) Capture state snapshots with scripts/capture-state.sh\n  5) Run scripts/report-compactness-check.sh before closing report\n' \
  "${run_dir}"
