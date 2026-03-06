#!/usr/bin/env bash
# PreToolUse/Bash hook for kuma-manual-test
# Combines: M1 (bare kubectl apply), M2 (--validate=false), M4 (system kumactl), M5 (unrecorded cluster cmd)
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/_lib.sh"

input="$(cat)"
command="$(printf '%s' "${input}" | jq -r '.tool_input.command // ""')"

if [[ -z "${command}" ]]; then
  exit 0
fi

# M2: --validate=false is forbidden (first match = block)
if printf '%s' "${command}" | grep -qF -- '--validate=false'; then
  jq -nc '{
    hookSpecificOutput: {
      hookEventName: "PreToolUse",
      permissionDecision: "deny",
      permissionDecisionReason: "[KMT002] --validate=false is forbidden. Validation errors mean the manifest or CRD is wrong. Fix the root cause.",
      additionalContext: "Automated kuma-manual-test hook. Remove --validate=false and retry."
    },
    systemMessage: "\nKMT002: --validate=false blocked\n  Fix: Remove the flag and fix the manifest or CRD\n"
  }'
  exit 0
fi

# M1: bare kubectl apply (allow dry-run, tracked scripts)
if printf '%s' "${command}" | grep -qE 'kubectl[[:space:]].*apply'; then
  if ! printf '%s' "${command}" | grep -qE '(apply-tracked-manifest\.sh|validate-manifest\.sh|--dry-run)'; then
    jq -nc '{
      hookSpecificOutput: {
        hookEventName: "PreToolUse",
        permissionDecision: "deny",
        permissionDecisionReason: "[KMT001] Bare kubectl apply detected. All manifests must go through apply-tracked-manifest.sh for tracking and reproducibility.",
        additionalContext: "Automated kuma-manual-test hook. Use apply-tracked-manifest.sh --run-dir ... --manifest ... --step ..."
      },
      systemMessage: "\nKMT001: Bare kubectl apply blocked\n  Fix: Use apply-tracked-manifest.sh\n"
    }'
    exit 0
  fi
fi

# M4: bare kumactl (not a locally built path)
if printf '%s' "${command}" | grep -qE '(^|[;&|[:space:]])kumactl[[:space:]]'; then
  if ! printf '%s' "${command}" | grep -qE '(build/|find-local-kumactl)'; then
    jq -nc '{
      hookSpecificOutput: {
        hookEventName: "PreToolUse",
        permissionDecision: "allow",
        additionalContext: "kuma-manual-test warning: [KMT004] Bare kumactl command detected. Use the locally built ${KUMACTL} resolved in Phase 0, not the system binary."
      }
    }'
    exit 0
  fi
fi

# M5: cluster command not wrapped in record-command.sh
# Detect kubectl/kumactl/curl commands against the cluster
if printf '%s' "${command}" | grep -qE '(kubectl|kumactl|curl).*(--(kubeconfig|context)|kind-kuma)'; then
  # Skip if already wrapped in record-command.sh, apply-tracked-manifest.sh, capture-state.sh, preflight.sh, or cluster-lifecycle.sh
  if ! printf '%s' "${command}" | grep -qE '(record-command\.sh|apply-tracked-manifest\.sh|capture-state\.sh|preflight\.sh|cluster-lifecycle\.sh|find-local-kumactl\.sh|validate-manifest\.sh)'; then
    jq -nc '{
      hookSpecificOutput: {
        hookEventName: "PreToolUse",
        permissionDecision: "deny",
        permissionDecisionReason: "[KMT005] Bare cluster command blocked. ALL kubectl, kumactl, and curl commands against the cluster MUST be executed through record-command.sh for reproducibility. No exceptions.",
        additionalContext: "Automated kuma-manual-test hook. Wrap the command: record-command.sh --run-dir ... --phase test --label ... -- <command>"
      },
      systemMessage: "\nKMT005: Bare cluster command blocked\n  Fix: Execute through record-command.sh --run-dir ${RUN_DIR} --phase test --label <label> -- <command>\n"
    }'
    exit 0
  fi
fi

# Clean pass - no output
exit 0
