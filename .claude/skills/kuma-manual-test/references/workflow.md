# Contents

1. [Resuming a partial run](#resuming-a-partial-run)
2. [Phase 0 - environment check](#phase-0---environment-check)
3. [Phase 1 - initialize run](#phase-1---initialize-run)
4. [Phase 2 - bootstrap cluster](#phase-2---bootstrap-cluster)
5. [Phase 3 - preflight](#phase-3---preflight)
6. [Phase 4 - execute tests](#phase-4---execute-tests)
7. [Phase 5 - failure handling](#phase-5---failure-handling)
8. [Phase 6 - closeout](#phase-6---closeout)
9. [Performance toggles](#performance-toggles)

---

# Workflow

Supplementary detail for the seven-phase execution flow in SKILL.md. Each phase includes its code blocks so this file is self-contained when loaded independently.

## Resuming a partial run

If a previous run was interrupted, check `runs/<run-id>/run-status.yaml` for `last_completed_group` and `next_planned_group`. Skip to the next planned group and continue from there. Do not re-run already-passed groups unless investigating a failure.

## Phase 0 - environment check

Resolve persistent storage and repo root first:

```bash
DATA_DIR="$(echo "${XDG_DATA_HOME:-$HOME/.local/share}/kuma/kuma-manual-test")"
mkdir -p "${DATA_DIR}/suites" "${DATA_DIR}/runs"
```

Resolve `REPO_ROOT`: `--repo` flag > check if cwd has `go.mod` with `kumahq/kuma` > fail with message.

Build and verify kumactl:

```bash
make --directory "${REPO_ROOT}" build/kumactl
KUMACTL="$("${CLAUDE_SKILL_DIR}/scripts/find-local-kumactl.sh" --repo-root "${REPO_ROOT}")"
"${KUMACTL}" version
```

**Gate**: kumactl version output matches the repo HEAD.

## Phase 1 - initialize run

```bash
RUNS_DIR="${DATA_DIR}/runs"
RUN_ID="$(date +%Y%m%d-%H%M%S)-manual"
"${CLAUDE_SKILL_DIR}/scripts/init-run.sh" --runs-dir "${RUNS_DIR}" --session-id "${SESSION_ID}" "${RUN_ID}"
RUN_DIR="${RUNS_DIR}/${RUN_ID}"
```

Suite resolution uses the three-step order (directory suite, legacy `.md` file, literal path). Set `SUITE_DIR` and `SUITE_FILE` accordingly.

Fill `run-metadata.yaml` before touching the cluster.

**Gate**: `run-metadata.yaml` exists and has profile, feature scope, and environment filled in.

## Phase 2 - bootstrap cluster

Read `references/cluster-setup.md` before starting this phase.

Pick a profile and start the cluster:

```bash
# Single-zone:
"${CLAUDE_SKILL_DIR}/scripts/cluster-lifecycle.sh" --repo-root "${REPO_ROOT}" single-up kuma-1

# Or multi-zone:
"${CLAUDE_SKILL_DIR}/scripts/cluster-lifecycle.sh" --repo-root "${REPO_ROOT}" global-two-zones-up kuma-1 kuma-2 kuma-3 zone-1 zone-2
```

If changes modify CRDs, refresh them after deploy:

```bash
kubectl --kubeconfig "${HOME}/.kube/kind-kuma-1-config" \
  apply --server-side --force-conflicts \
  -f "${REPO_ROOT}/deployments/charts/kuma/crds/"
```

**Gate**: `kubectl get pods -n kuma-system` shows all pods Running/Ready.

## Phase 3 - preflight

```bash
"${CLAUDE_SKILL_DIR}/scripts/preflight.sh" \
  --kubeconfig "${HOME}/.kube/kind-kuma-1-config" \
  --run-dir "${RUN_DIR}" \
  --repo-root "${REPO_ROOT}"

"${CLAUDE_SKILL_DIR}/scripts/capture-state.sh" \
  --kubeconfig "${HOME}/.kube/kind-kuma-1-config" \
  --run-dir "${RUN_DIR}" \
  --label "preflight"
```

Do not start tests until preflight is green.

**Gate**: preflight script exits 0 and state snapshot is saved.

## Phase 4 - execute tests

Read `references/validation.md` before applying manifests.
Read `references/mesh-policies.md` for Mesh\* policy targeting and debug flow.

Select a suite that matches the feature area, or copy `examples/suite-template.md` if none exists.

For directory suites (`SUITE_DIR` is set):

1. Read `${SUITE_DIR}/suite.md` for overview, group table, baseline table, and execution contract.
2. **The test groups table is authoritative.** Every group listed in the table MUST be executed. Do not skip groups because they need a different cluster profile. If a group requires multi-zone but the current profile is single-zone, tear down the current cluster and bring up a multi-zone cluster before that group. If profile switching is impractical, use AskUserQuestion - never silently skip.
3. Before G1, apply each baseline manifest from `${SUITE_DIR}/baseline/` using the baseline table. Copy each to `${RUN_DIR}/manifests/` and apply from there.
4. Before each group, read the group file from `${SUITE_DIR}/groups/` using the file path in the group table. The group file is the **authoritative source** for that group's manifests, validation commands, and expected outcomes. Extract inline YAML manifests from fenced code blocks, write them to `${RUN_DIR}/manifests/`, and apply from there. Do not improvise, modify, or rewrite manifests from group files.
5. Follow the group file's validation commands and expected outcomes exactly. If something doesn't match, report it as a finding - do not silently adjust expectations.
6. After completing a group, the group file content can be dropped from context.

For legacy single-file suites: read the entire suite file as before.

**Deviation rule**: if any step requires diverging from the suite definition (different values, skipped step, reordered steps, extra steps, changed expected outcome), use AskUserQuestion for approval BEFORE making the change. Record every deviation in the report with what changed, why, and whether user-approved or suite-allowed.

For each test step:

1. Write manifest to `${RUN_DIR}/manifests/` (copy from suite baseline/groups, or write inline YAML there - never to `/tmp`). When the suite group provides inline YAML, use it verbatim. If the manifest needs changes, ask the user first.
2. Validate and apply through the tracked scripts.
3. Run verification commands (`kumactl inspect`, `curl`, `kubectl get`, etc.) THROUGH `"${CLAUDE_SKILL_DIR}/scripts/record-command.sh"`. Never run these bare - the script is the execution method, not a post-hoc logger. Save output to `artifacts/`.
4. Run cleanup commands (`kubectl delete`) THROUGH `"${CLAUDE_SKILL_DIR}/scripts/record-command.sh"`. Every command that touches the cluster goes through this script.
5. Write result into the report. Every artifact path referenced must point to an existing file.

```bash
# Write manifest to run dir first
"${CLAUDE_SKILL_DIR}/scripts/apply-tracked-manifest.sh" \
  --run-dir "${RUN_DIR}" \
  --kubeconfig "${HOME}/.kube/kind-kuma-1-config" \
  --manifest "${RUN_DIR}/manifests/<file>" \
  --step "<step-name>"

# Record every verification/cleanup command
"${CLAUDE_SKILL_DIR}/scripts/record-command.sh" \
  --run-dir "${RUN_DIR}" \
  --phase "test" \
  --label "<step-label>" \
  -- <command>
```

After completing each group (hard gate - do not skip any of these):

1. Capture state:

```bash
"${CLAUDE_SKILL_DIR}/scripts/capture-state.sh" \
  --kubeconfig "${HOME}/.kube/kind-kuma-1-config" \
  --run-dir "${RUN_DIR}" \
  --label "after-<group-id>"
```

2. Update `run-status.yaml` with `last_completed_group`, `next_planned_group`, pass/fail/skipped counts, and `last_updated_utc`.
3. Verify `run-status.yaml` was written correctly before starting the next group.

On first unexpected failure, go to Phase 5.

**Gate**: all planned tests have pass/fail entries in the report. Every artifact path in the report resolves to an existing file. `run-status.yaml` reflects final counts.

## Phase 5 - failure handling

Read `references/troubleshooting.md` for known failure modes.

1. Stop progression.
2. Capture immediate state snapshot.
3. Document expected vs observed.
4. Classify the issue (manifest, environment, product bug).
5. Continue only when classification is explicit.

```bash
"${CLAUDE_SKILL_DIR}/scripts/capture-state.sh" \
  --kubeconfig "${HOME}/.kube/kind-kuma-1-config" \
  --run-dir "${RUN_DIR}" \
  --label "failure-<test-id>"
```

## Phase 6 - closeout

```bash
"${CLAUDE_SKILL_DIR}/scripts/capture-state.sh" \
  --kubeconfig "${HOME}/.kube/kind-kuma-1-config" \
  --run-dir "${RUN_DIR}" \
  --label "postrun"

"${CLAUDE_SKILL_DIR}/scripts/report-compactness-check.sh" \
  --report "${RUN_DIR}/reports/manual-test-report.md"
```

**Gate**: all of these are true before marking the run complete:

- Command log is complete (every command executed has an entry)
- Manifest index includes every apply
- Report has pass or fail for all planned tests
- Failures include triage details and artifact paths
- Every artifact path referenced in the report resolves to an existing file in the run directory
- `run-status.yaml` has correct final counts matching the report
- State captures exist for preflight, each completed group, and postrun
- Report compactness check passes

After all gates pass, tear down the clusters. This is the default - always clean up unless the user explicitly asks to keep clusters running or the suite metadata includes `keep_clusters: true`.

```bash
# Single-zone
"${CLAUDE_SKILL_DIR}/scripts/cluster-lifecycle.sh" --repo-root "${REPO_ROOT}" single-down kuma-1

# Multi-zone (global + 2 zones)
"${CLAUDE_SKILL_DIR}/scripts/cluster-lifecycle.sh" --repo-root "${REPO_ROOT}" global-two-zones-down kuma-1 kuma-2 kuma-3
```

## Performance toggles

| Profile                      | `HARNESS_BUILD_IMAGES` | `HARNESS_LOAD_IMAGES` | `HARNESS_HELM_CLEAN` | Use when                             |
| ---------------------------- | ---------------------- | --------------------- | -------------------- | ------------------------------------ |
| default (fastest functional) | 1                      | 1                     | 0                    | Normal test runs                     |
| strict clean-state           | 1                      | 1                     | 1                    | Need full isolation between deploys  |
| image-stable fast            | 0                      | 0                     | 0                    | Images already match code under test |

Example:

```bash
HARNESS_BUILD_IMAGES=0 HARNESS_LOAD_IMAGES=0 \
  "${CLAUDE_SKILL_DIR}/scripts/cluster-lifecycle.sh" --repo-root "${REPO_ROOT}" single-up kuma-1
```
