# Contents

1. [Non-negotiable rules](#non-negotiable-rules)
2. [Operating mode](#operating-mode)
3. [Persistent data directory](#persistent-data-directory)
4. [Local kumactl setup](#local-kumactl-setup)
5. [Run status tracking](#run-status-tracking)
6. [Artifacts to collect per test case](#artifacts-to-collect-per-test-case)
7. [Failure policy](#failure-policy)
8. [Bug triage protocol](#bug-triage-protocol)
9. [Reproducibility minimum bar](#reproducibility-minimum-bar)

---

# Agent contract

Rules for any AI agent executing manual tests with this harness.

## Non-negotiable rules

1. Use locally built `kumactl` from `build/` only. Never the system binary.
2. Apply every manifest through `"${CLAUDE_SKILL_DIR}/scripts/apply-tracked-manifest.sh"`.
3. Record **every** command executed against the cluster or kumactl via `"${CLAUDE_SKILL_DIR}/scripts/record-command.sh"`. This means: `kumactl inspect`, `curl`, `kubectl exec`, `kubectl get`, `kubectl logs`, `kubectl delete`, and any other command that touches the cluster or produces test evidence. If you ran it, record it. Unrecorded commands make the run non-reproducible.
4. Run server-side dry-run validation before every apply.
5. Capture cluster state snapshots before and after each test group. This is a hard gate - do not start the next group until the state capture for the completed group is saved.
6. Stop and triage on first unexpected failure.
7. Keep reports concise - store raw output in `artifacts/`, reference file paths.
8. Never create manifests in `/tmp` or any location outside `${RUN_DIR}/manifests/`. All manifests must be written to the run directory before apply.
9. When a suite group provides inline manifest YAML, use it verbatim. Do not change names, namespaces, labels, or any fields. If the suite manifest is wrong, note it in the report and SKIP the step - do not silently modify it.
10. Every artifact path written in the report must resolve to an existing file in the run directory. Before closeout, verify every path. A report that references missing files is a broken run.
11. Update `run-status.yaml` after every completed group with `last_completed_group`, `next_planned_group`, counts, and timestamp. This is a hard gate - do not start the next group without updating status first.
12. **No autonomous deviations.** Any divergence from the suite definition - different manifest values, skipped steps, reordered steps, extra steps, modified expected outcomes - requires explicit user approval via AskUserQuestion BEFORE the change is made. The only exception is when the suite definition itself explicitly marks a decision as agent-discretionary (e.g., "agent may choose" or "optional"). Even suite-allowed deviations must be recorded. Record every deviation in the report as a "Deviation" entry with: (a) what was changed, (b) why, (c) whether it was user-approved or suite-allowed, and (d) the exact user response if user-approved.

## Operating mode

- Be systematic, not fast.
- Favor reproducibility over speed.
- Keep one source of truth per run in `${DATA_DIR}/runs/<run-id>/`.
- Treat missing artifacts as "test not executed".

## Persistent data directory

All run artifacts and authored suites are stored in a persistent data directory outside the repo:

```
${XDG_DATA_HOME:-$HOME/.local/share}/kuma/kuma-manual-test/
├── suites/                  # Authored test suites (by kuma-suite-author)
│   ├── motb-core/           # Directory-format suite (v2)
│   │   ├── suite.md
│   │   ├── baseline/
│   │   │   └── *.yaml
│   │   └── groups/
│   │       └── g{NN}-*.md
│   └── meshretry-basic.md   # Legacy single-file suite (v1)
├── runs/                    # Test run artifacts (by kuma-manual-test)
└── .last-run                # Last run ID (optional, for quick resume)
```

## Local kumactl setup

```bash
make --directory "${REPO_ROOT}" build/kumactl
KUMACTL="$("${CLAUDE_SKILL_DIR}/scripts/find-local-kumactl.sh" --repo-root "${REPO_ROOT}")"
"${KUMACTL}" version
```

Record `kumactl version` output in the run report.

For Mesh\* policy authoring, matching, and debug commands, follow
`references/mesh-policies.md`.

## Run status tracking

After each test group, update `run-status.yaml` with:

- `last_completed_group`: the group just finished (e.g. "G3")
- `next_planned_group`: the next group to execute (e.g. "G4")
- `counts`: running pass/fail/skipped totals
- `last_updated_utc`: timestamp

This enables resuming partial runs. See `references/workflow.md` (resuming a partial run).

## Artifacts to collect per test case

- Manifest file copy in `runs/<run-id>/manifests/*.yaml`
- Command output logs in `runs/<run-id>/artifacts/*.log` for every command (apply, inspect, curl, delete, etc.)
- Command log entry in `runs/<run-id>/commands/command-log.md` for every command executed
- State capture in `runs/<run-id>/state/` after each group completion
- Result and interpretation in `runs/<run-id>/reports/manual-test-report.md`

Every verification command (`kumactl inspect`, `curl`, `kubectl get`) must produce an artifact file. Every cleanup command (`kubectl delete`) must have a command log entry. If a command was worth running, it was worth recording.

## Failure policy

On unexpected behavior:

1. Stop advancing the suite.
2. Re-run the failing step once to check determinism.
3. Snapshot cluster state immediately.
4. Classify the failure: manifest issue, environment issue, or product bug.
5. Continue only when classification is explicit in the report.

Never mark a test as pass with unresolved ambiguity.

## Bug triage protocol

When a bug is suspected, report:

- Exact manifest used and its SHA256
- Exact command sequence
- Observed vs expected output
- Scope assessment (single policy, all policies, one mode, all modes)
- Suggested next isolation step

## Reproducibility minimum bar

A run is complete only when all of these are present:

- Cluster profile used
- Kubeconfig paths used
- Local build/deploy commands
- Full manifest set
- Full command log
- Test report with pass/fail and artifact pointers
