---
name: kuma-manual-test
description: >-
  Execute reproducible manual tests on local k3d infrastructure for Kuma service mesh features.
  Use when running manual verification, testing policy changes on real clusters, validating xDS
  config generation, or doing k3d manual test runs for any Kuma feature area.
argument-hint: "[suite-path] [--profile single-zone|multi-zone] [--repo /path/to/kuma] [--run-id ID] [--resume RUN_ID]"
allowed-tools: AskUserQuestion, Bash, Read, Task, Write
user-invocable: true
disable-model-invocation: true
hooks:
  PreToolUse:
    - matcher: "Bash"
      hooks:
        - type: command
          command: "$CLAUDE_PROJECT_DIR/.claude/skills/kuma-manual-test/scripts/hooks/guard-bash.sh"
    - matcher: "Write"
      hooks:
        - type: command
          command: "$CLAUDE_PROJECT_DIR/.claude/skills/kuma-manual-test/scripts/hooks/guard-write.sh"
  PostToolUse:
    - matcher: "Bash"
      hooks:
        - type: command
          command: "$CLAUDE_PROJECT_DIR/.claude/skills/kuma-manual-test/scripts/hooks/verify-bash.sh"
        - type: command
          command: "$CLAUDE_PROJECT_DIR/.claude/skills/kuma-manual-test/scripts/hooks/audit.sh"
    - matcher: "Write"
      hooks:
        - type: command
          command: "$CLAUDE_PROJECT_DIR/.claude/skills/kuma-manual-test/scripts/hooks/verify-write.sh"
        - type: command
          command: "$CLAUDE_PROJECT_DIR/.claude/skills/kuma-manual-test/scripts/hooks/audit.sh"
  PostToolUseFailure:
    - matcher: "Bash"
      hooks:
        - type: command
          command: "$CLAUDE_PROJECT_DIR/.claude/skills/kuma-manual-test/scripts/hooks/enrich-failure.sh"
        - type: command
          command: "$CLAUDE_PROJECT_DIR/.claude/skills/kuma-manual-test/scripts/hooks/audit.sh"
  SubagentStart:
    - matcher: "general-purpose"
      hooks:
        - type: command
          command: "$CLAUDE_PROJECT_DIR/.claude/skills/kuma-manual-test/scripts/hooks/context-preflight.sh"
  SubagentStop:
    - matcher: "general-purpose"
      hooks:
        - type: command
          command: "$CLAUDE_PROJECT_DIR/.claude/skills/kuma-manual-test/scripts/hooks/validate-preflight.sh"
  Stop:
    - hooks:
        - type: command
          command: "$CLAUDE_PROJECT_DIR/.claude/skills/kuma-manual-test/scripts/hooks/guard-incomplete-stop.sh"
---

# Kuma manual test

Execute reproducible manual tests on local k3d clusters for any Kuma feature. Track every manifest, command, and artifact for full run reproducibility.

## Arguments

Parse from `$ARGUMENTS`:

| Argument     | Default         | Purpose                                                                                                                           |
| :----------- | :-------------- | :-------------------------------------------------------------------------------------------------------------------------------- |
| (positional) | -               | Suite path or bare name. Bare names (no `/`) are looked up in `${DATA_DIR}/suites/` first. Prompt with AskUserQuestion if omitted |
| `--profile`  | `single-zone`   | Cluster profile: `single-zone`, `multi-zone`                                                                                      |
| `--repo`     | auto-detect cwd | Path to Kuma repo checkout                                                                                                        |
| `--run-id`   | timestamp-based | Override run identifier                                                                                                           |
| `--resume`   | -               | Resume a partial run by its run ID                                                                                                |

## Preprocessed context

- Data directory: !`echo "${XDG_DATA_HOME:-$HOME/.local/share}/kuma/kuma-manual-test"`
- Home: !`echo "$HOME"`
- Timestamp: !`date +%Y%m%d-%H%M%S`
- Session ID: ${CLAUDE_SESSION_ID}
- Docker: !`docker info >/dev/null 2>&1 && echo "running" || echo "not running"`
- k3d: !`command -v k3d >/dev/null 2>&1 && echo "installed" || echo "MISSING"`
- kubectl: !`command -v kubectl >/dev/null 2>&1 && echo "installed" || echo "MISSING"`
- helm: !`command -v helm >/dev/null 2>&1 && echo "installed" || echo "MISSING"`

Use these pre-resolved values throughout the run. `DATA_DIR` is the data directory above. `HOME` is the home path above. The timestamp above becomes the default `RUN_ID` suffix. The session ID tracks which Claude Code session produced this run. If the session ID is empty or contains literal `${`, use `standalone` instead. If Docker shows "not running" or any tool shows "MISSING", stop immediately and report the problem.

## Non-negotiable rules

1. Use locally built `kumactl` from `build/` only.
2. Apply every manifest through `"${CLAUDE_SKILL_DIR}/scripts/apply-tracked-manifest.sh"`.
3. Record **every** command executed against the cluster or kumactl via `"${CLAUDE_SKILL_DIR}/scripts/record-command.sh"`. This includes `kumactl inspect`, `curl`, `kubectl delete`, `kubectl exec`, `kubectl get`, `kubectl logs` - not just applies. If it touches the cluster or produces test evidence, record it.
4. Stop and triage on first unexpected failure.
5. Never use `--validate=false` on any kubectl command. Validation errors mean the manifest or CRD is wrong - fix the root cause.
6. Never create manifests in `/tmp`. Write all manifests to `${RUN_DIR}/manifests/` before applying.
7. When a suite group provides inline manifest YAML, use it verbatim. Do not improvise names, namespaces, or fields unless the user explicitly requests it.
8. After completing each test group, capture state with `capture-state.sh`, update `run-status.yaml`, and verify these files before starting the next group. This is a hard gate - do not proceed without it.
9. Every artifact path written in the report must resolve to an existing file. Verify before closeout.
10. **No autonomous deviations.** If a test step needs to diverge from the suite definition (different manifest values, skipped steps, changed order, extra steps, modified expected outcomes), use AskUserQuestion to get explicit approval BEFORE making the change. The only exception is when the suite definition itself explicitly marks something as agent-discretionary. Record every deviation decision in the report with: what changed, why, and whether it was user-approved or suite-allowed.

Read [references/agent-contract.md](references/agent-contract.md) for full agent behavior rules and artifact requirements.

## Workflow

Read [references/workflow.md](references/workflow.md) for supplementary phase details and verification gates.

### Phase 0: Environment check

1. Set `DATA_DIR` to the pre-resolved data directory from "Preprocessed context". Create the subdirectories:

```bash
DATA_DIR="<data directory from Preprocessed context>"
mkdir -p "${DATA_DIR}/suites" "${DATA_DIR}/runs"
```

2. Resolve `REPO_ROOT`: use `--repo` flag if provided, otherwise check if cwd has `go.mod` containing `kumahq/kuma`. Fail with a message if neither works.
3. Docker status and tool availability are pre-resolved in "Preprocessed context". If Docker is "not running" or any tool is "MISSING", stop and report the problem. No need to re-check.
4. Build local kumactl:

```bash
make --directory "${REPO_ROOT}" build/kumactl
KUMACTL="$("${CLAUDE_SKILL_DIR}/scripts/find-local-kumactl.sh" --repo-root "${REPO_ROOT}")"
"${KUMACTL}" version
```

**Gate**: kumactl version output matches the repo HEAD.

### Phase 1: Initialize run

```bash
RUNS_DIR="${DATA_DIR}/runs"
RUN_ID="${RUN_ID:-<timestamp from Preprocessed context>-manual}"  # override with --run-id flag
SESSION_ID="<session ID from Preprocessed context, or 'standalone' if empty/unreplaced>"
"${CLAUDE_SKILL_DIR}/scripts/init-run.sh" --runs-dir "${RUNS_DIR}" --session-id "${SESSION_ID}" "${RUN_ID}"
RUN_DIR="${RUNS_DIR}/${RUN_ID}"
```

If `--resume` was passed, read `${RUNS_DIR}/${RESUME_ID}/run-status.yaml` for `last_completed_group` and skip to the next planned group.

Suite resolution for bare names (no `/`):

1. Directory suite: check `${DATA_DIR}/suites/${name}/suite.md`
2. Legacy file: check `${DATA_DIR}/suites/${name}.md`
3. Literal path

For directory suites, set `SUITE_DIR` to the suite directory and `SUITE_FILE` to `suite.md`. For legacy/literal suites, set `SUITE_FILE` to the resolved path and `SUITE_DIR` to empty.

Fill `run-metadata.yaml` with profile, feature scope, and kumactl version before touching the cluster.

**Gate**: `run-metadata.yaml` exists with profile, feature scope, and environment filled in.

### Phase 2: Bootstrap cluster

Read [references/cluster-setup.md](references/cluster-setup.md) before starting this phase.

Select the cluster topology based on the `--profile` flag (default: `single-zone`):

```bash
# Single-zone (--profile single-zone):
"${CLAUDE_SKILL_DIR}/scripts/cluster-lifecycle.sh" --repo-root "${REPO_ROOT}" single-up kuma-1

# Multi-zone (--profile multi-zone):
"${CLAUDE_SKILL_DIR}/scripts/cluster-lifecycle.sh" --repo-root "${REPO_ROOT}" global-two-zones-up kuma-1 kuma-2 kuma-3 zone-1 zone-2
```

If changes modify CRDs, refresh them after deploy:

```bash
kubectl --kubeconfig "${HOME}/.kube/kind-kuma-1-config" \
  apply --server-side --force-conflicts \
  -f "${REPO_ROOT}/deployments/charts/kuma/crds/"
```

**Gate**: `kubectl get pods -n kuma-system` shows all pods Running/Ready.

### Phase 3: Preflight (spawned agent)

Spawn a `general-purpose` preflight agent to run cluster readiness checks and capture initial state. This isolates verbose cluster introspection (kubectl describe, logs, events) from the main context.

Create the agent with a prompt that includes:

- Kubeconfig path: `<Home from Preprocessed context>/.kube/kind-kuma-1-config`
- Run directory: `${RUN_DIR}`
- Repo root: `${REPO_ROOT}`
- Preflight script path: `${CLAUDE_SKILL_DIR}/scripts/preflight.sh`
- Capture script path: `${CLAUDE_SKILL_DIR}/scripts/capture-state.sh`
- Reference files to read: the cluster-setup and validation references from Phase 4 (pass their absolute paths)

The agent must:

1. Run `preflight.sh` with the kubeconfig, run-dir, and repo-root flags.
2. Run `capture-state.sh` with kubeconfig, run-dir, and label `preflight`.
3. Return ONLY: pass/fail result, state capture directory path, and any warnings or blockers.

Poll the agent task until complete. Do not start tests until the agent reports pass.

### Phase 4: Execute tests

Read [references/validation.md](references/validation.md) for the pre-apply checklist and safe apply flow before applying manifests.
Read [references/mesh-policies.md](references/mesh-policies.md) for policy authoring rules when the suite tests any `Mesh*` policy.

Select a suite from the positional argument, or use AskUserQuestion if none was provided. Read [examples/suite-template.md](examples/suite-template.md) as the starting point when creating a new suite for an untested feature. Read [examples/example-motb-core-suite.md](examples/example-motb-core-suite.md) for a worked example of the expected suite format.

For directory suites (`SUITE_DIR` is set):

1. Read `${SUITE_DIR}/suite.md` for the group table, baseline table, and execution contract.
2. Before G1, apply all baseline manifests from `${SUITE_DIR}/baseline/` using the baseline table.
3. Before each group, read the group file from `${SUITE_DIR}/groups/` using the file path in the group table. The group file is the **authoritative source** for that group's manifests, commands, and expected outcomes. Use its inline YAML manifests verbatim - copy them to `${RUN_DIR}/manifests/` and apply from there. Do not improvise or rewrite manifests.
4. After completing a group, the group file content can be dropped from context.

For legacy single-file suites: read the entire file as before.

For each test step:

1. Write manifest to `${RUN_DIR}/manifests/` (copy from suite or create there - never `/tmp`).
2. Validate the manifest.
3. Apply through the tracked script.
4. Run verification commands (`kumactl inspect`, `curl`, `kubectl get`, etc.) and record each one via `record-command.sh`, saving output to `artifacts/`.
5. Run any cleanup commands (`kubectl delete`) and record each one via `record-command.sh`.
6. Write result into the report. Every artifact path referenced must point to a real file.

```bash
"${CLAUDE_SKILL_DIR}/scripts/apply-tracked-manifest.sh" \
  --run-dir "${RUN_DIR}" \
  --kubeconfig "${HOME}/.kube/kind-kuma-1-config" \
  --manifest "${RUN_DIR}/manifests/<manifest-file>" \
  --step "<step-name>"
```

After completing each group (hard gate - do not skip):

```bash
# 1. Capture state
"${CLAUDE_SKILL_DIR}/scripts/capture-state.sh" \
  --kubeconfig "${HOME}/.kube/kind-kuma-1-config" \
  --run-dir "${RUN_DIR}" \
  --label "after-<group-id>"

# 2. Update run-status.yaml with last_completed_group, next_planned_group, counts
# 3. Verify run-status.yaml was written correctly before proceeding
```

On first unexpected failure, go to Phase 5.

**Gate**: all planned tests have pass/fail entries in the report. Every artifact path in the report resolves to an existing file. `run-status.yaml` reflects final counts.

### Phase 5: Failure handling

Read [references/troubleshooting.md](references/troubleshooting.md) for known failure modes.

1. Stop progression.
2. Re-run the failing step once to check determinism.
3. Capture state snapshot with `"${CLAUDE_SKILL_DIR}/scripts/capture-state.sh"`.
4. Classify: manifest issue, environment issue, or product bug.
5. Continue only when classification is explicit in the report.

### Phase 6: Closeout

```bash
"${CLAUDE_SKILL_DIR}/scripts/capture-state.sh" \
  --kubeconfig "${HOME}/.kube/kind-kuma-1-config" \
  --run-dir "${RUN_DIR}" \
  --label "postrun"

"${CLAUDE_SKILL_DIR}/scripts/report-compactness-check.sh" \
  --report "${RUN_DIR}/reports/manual-test-report.md"
```

**Gate**: command log complete (every command has an entry), manifest index complete, all tests have pass/fail, every artifact path in the report resolves to an existing file, `run-status.yaml` has correct final counts, state captures exist for preflight + each completed group + postrun, compactness check passes.

After all gates pass, tear down the clusters. This is the default - always clean up unless the user explicitly asks to keep clusters running or the suite metadata includes `keep_clusters: true`.

```bash
# Single-zone
"${CLAUDE_SKILL_DIR}/scripts/cluster-lifecycle.sh" --repo-root "${REPO_ROOT}" single-down kuma-1

# Multi-zone (global + 2 zones)
"${CLAUDE_SKILL_DIR}/scripts/cluster-lifecycle.sh" --repo-root "${REPO_ROOT}" global-two-zones-down kuma-1 kuma-2 kuma-3
```

## Performance toggles

| Profile | HARNESS_BUILD_IMAGES | HARNESS_LOAD_IMAGES | HARNESS_HELM_CLEAN | Use when                       |
| ------- | -------------------- | ------------------- | ------------------ | ------------------------------ |
| default | 1                    | 1                   | 0                  | Normal test runs               |
| strict  | 1                    | 1                   | 1                  | Full isolation between deploys |
| fast    | 0                    | 0                   | 0                  | Images already match code      |

## Report compactness thresholds

| Limit                | Default   | Flag                     |
| -------------------- | --------- | ------------------------ |
| Total lines          | 220       | `--max-lines`            |
| Line length          | 220 chars | `--max-line-length`      |
| Code blocks          | 4         | `--max-code-blocks`      |
| Lines per code block | 30        | `--max-code-block-lines` |

Store raw output in `artifacts/` and reference file paths from the report.

## Bundled resources

- [references/agent-contract.md](references/agent-contract.md) - agent behavior rules, failure policy, artifact requirements
- [references/workflow.md](references/workflow.md) - detailed phase descriptions with verification gates
- [references/cluster-setup.md](references/cluster-setup.md) - k3d profiles, kubeconfig mapping, deploy commands
- [references/mesh-policies.md](references/mesh-policies.md) - Mesh\* policy authoring, targeting, and debug flow
- [references/validation.md](references/validation.md) - pre-apply checklist and safe apply flow
- [references/troubleshooting.md](references/troubleshooting.md) - 10 known failure modes with fixes
- `${CLAUDE_SKILL_DIR}/scripts/init-run.sh` - create run directory from templates
- `${CLAUDE_SKILL_DIR}/scripts/preflight.sh` - verify cluster readiness
- `${CLAUDE_SKILL_DIR}/scripts/cluster-lifecycle.sh` - start/stop/deploy k3d clusters by profile
- `${CLAUDE_SKILL_DIR}/scripts/validate-manifest.sh` - server-side dry-run and diff before apply
- `${CLAUDE_SKILL_DIR}/scripts/apply-tracked-manifest.sh` - apply with validation, copy, and logging
- `${CLAUDE_SKILL_DIR}/scripts/capture-state.sh` - snapshot cluster state
- `${CLAUDE_SKILL_DIR}/scripts/record-command.sh` - log ad-hoc command to command log
- `${CLAUDE_SKILL_DIR}/scripts/find-local-kumactl.sh` - locate locally built kumactl
- `${CLAUDE_SKILL_DIR}/scripts/report-compactness-check.sh` - verify report size limits
- `assets/run-metadata.template.yaml` - run metadata template
- `assets/run-status.template.yaml` - run progress tracking template
- `assets/command-log.template.md` - command log template
- `assets/manifest-index.template.md` - manifest index template
- `assets/manual-test-report.template.md` - test report template
- [examples/suite-template.md](examples/suite-template.md) - generic test suite template
- [examples/example-motb-core-suite.md](examples/example-motb-core-suite.md) - worked example for MOTB testing (read when authoring new suites to see the expected format)
- `scripts/hooks/` - runtime guardrails (guard-bash.sh, guard-write.sh, verify-bash.sh, verify-write.sh, audit.sh, enrich-failure.sh, context-preflight.sh, validate-preflight.sh, guard-incomplete-stop.sh, _lib.sh)

## Example invocations

```bash
# Run suite from persistent storage (created by kuma-suite-author)
/kuma-manual-test meshretry-basic --repo ~/Projects/kuma

# Run from inside kuma repo (auto-detects repo root)
/kuma-manual-test meshretry-basic

# Explicit suite file path still works
/kuma-manual-test /path/to/my-suite.md --repo ~/Projects/kuma

# Multi-zone profile
/kuma-manual-test my-suite.md --profile multi-zone

# Resume from anywhere
/kuma-manual-test --resume 20260304-180131-manual --repo ~/Projects/kuma

# Custom run ID
/kuma-manual-test my-suite.md --run-id motb-validation-v2
```
