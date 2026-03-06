#!/usr/bin/env bash
# PostToolUseFailure/Bash hook for kuma-manual-test
# M11: add troubleshooting context on apply/validation failures
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/_lib.sh"

input="$(cat)"
error="$(printf '%s' "${input}" | jq -r '.error // ""')"
interrupt="$(printf '%s' "${input}" | jq -r '.is_interrupt // false')"

# Skip user cancellations
if [[ "${interrupt}" == "true" ]]; then
  exit 0
fi

if [[ -z "${error}" ]]; then
  exit 0
fi

hint=""

if printf '%s' "${error}" | grep -qiE 'no matches for kind|unknown field'; then
  hint="[KMT011a] Apply/validation failed. CRDs may be outdated. Try: kubectl apply --server-side --force-conflicts -f \${REPO_ROOT}/deployments/charts/kuma/crds/"
elif printf '%s' "${error}" | grep -qi 'admission webhook'; then
  hint="[KMT011b] Business validation rejection from admission webhook. Check validator.go for the rejection path and fix the manifest."
elif printf '%s' "${error}" | grep -qiE 'connection refused|timeout|connect:'; then
  hint="[KMT011c] Cluster may be unreachable. Check: kubectl cluster-info, docker ps. If k3d cluster was deleted, re-run Phase 2."
elif printf '%s' "${error}" | grep -qi 'already exists'; then
  hint="[KMT011d] Resource already exists from a previous group. Run cleanup for the conflicting resource before retrying."
fi

if [[ -n "${hint}" ]]; then
  jq -nc --arg hint "${hint}" '{
    systemMessage: $hint
  }'
fi

exit 0
