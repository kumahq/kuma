#!/usr/bin/env bash
# SubagentStart/general-purpose hook for kuma-manual-test
# M12: inject preflight requirements into spawned agent
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/_lib.sh"

jq -nc '{
  hookSpecificOutput: {
    hookEventName: "SubagentStart",
    additionalContext: "[KMT012] PREFLIGHT REQUIREMENTS: (1) Run preflight.sh with --kubeconfig, --run-dir, --repo-root flags. (2) Run capture-state.sh with --label preflight. (3) Return ONLY: pass/fail verdict, state capture directory path, and any warnings or blockers. Do not return raw kubectl output."
  }
}'
exit 0
