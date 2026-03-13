#!/usr/bin/env bash
# SubagentStart/Explore hook for kuma-suite-author
# S6: inject code reading requirements into spawned agent
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/_lib.sh"

jq -nc '{
  hookSpecificOutput: {
    hookEventName: "SubagentStart",
    additionalContext: "[KSA006] CODE READING REQUIREMENTS: Return ONLY a structured summary with two sections: (1) Group material (G1-G7): one entry per applicable group with one-line description, source path, and yes/no for sufficient material. (2) Variant signals (S1-S7): id, type, source, evidence, strength. Do NOT return raw file contents, full code blocks, or golden file text."
  }
}'
exit 0
