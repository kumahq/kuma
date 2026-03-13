#!/usr/bin/env bash
# Install Gateway API CRDs into a cluster.
# Version is extracted from the repo's go.mod to stay in sync.
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  install-gateway-api-crds.sh --kubeconfig <path> [--repo-root <path>] [--check-only]

Options:
  --kubeconfig <path>   Path to kubeconfig file (required)
  --repo-root <path>    Kuma repo root for version detection (default: auto-detect)
  --check-only          Only check if Gateway API CRDs are installed, exit 0/1

Installs the standard Gateway API CRDs (GatewayClass, Gateway, HTTPRoute, etc.)
from the upstream release matching the version in go.mod.
EOF
}

kubeconfig=""
repo_root=""
check_only=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --kubeconfig)
      kubeconfig="$2"
      shift 2
      ;;
    --repo-root)
      repo_root="$2"
      shift 2
      ;;
    --check-only)
      check_only=true
      shift
      ;;
    --help)
      usage
      exit 0
      ;;
    *)
      printf 'Error: unknown flag %s\n' "$1" >&2
      usage
      exit 1
      ;;
  esac
done

if [[ -z "${kubeconfig}" ]]; then
  printf 'Error: --kubeconfig is required\n' >&2
  exit 1
fi

# Resolve repo root for version detection
if [[ -z "${repo_root}" ]]; then
  repo_root="$(git rev-parse --show-toplevel 2>/dev/null)" || {
    printf 'Error: --repo-root not specified and git rev-parse failed\n' >&2
    exit 1
  }
fi

# Check if Gateway API CRDs are already installed
if kubectl --kubeconfig "${kubeconfig}" get crd gatewayclasses.gateway.networking.k8s.io >/dev/null 2>&1; then
  if [[ "${check_only}" == "true" ]]; then
    printf 'Gateway API CRDs are installed.\n'
    exit 0
  fi
  printf 'Gateway API CRDs already installed, skipping.\n'
  exit 0
fi

if [[ "${check_only}" == "true" ]]; then
  printf 'Gateway API CRDs are NOT installed.\n'
  exit 1
fi

# Extract version from go.mod
go_mod="${repo_root}/go.mod"
if [[ ! -f "${go_mod}" ]]; then
  printf 'Error: go.mod not found at %s\n' "${go_mod}" >&2
  exit 1
fi

gw_version="$(grep 'sigs.k8s.io/gateway-api' "${go_mod}" | awk '{print $2}' | head -1)"
if [[ -z "${gw_version}" ]]; then
  printf 'Error: gateway-api version not found in go.mod\n' >&2
  exit 1
fi

install_url="https://github.com/kubernetes-sigs/gateway-api/releases/download/${gw_version}/standard-install.yaml"

printf 'Installing Gateway API CRDs %s...\n' "${gw_version}"
kubectl --kubeconfig "${kubeconfig}" apply -f "${install_url}"
printf 'Gateway API CRDs %s installed.\n' "${gw_version}"
