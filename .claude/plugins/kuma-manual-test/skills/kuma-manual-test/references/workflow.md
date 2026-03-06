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
DATA_DIR="$(echo "${XDG_DATA_HOME:-$HOME/.local/share}/sai/kuma-manual-test")"
mkdir -p "${DATA_DIR}/suites" "${DATA_DIR}/runs"
```

Resolve `REPO_ROOT`: `--repo` flag > check if cwd has `go.mod` with `kumahq/kuma` > fail with message.

Build and verify kumactl:

```bash
make --directory "${REPO_ROOT}" build/kumactl
KUMACTL="$("$SKILL_DIR/scripts/find-local-kumactl.sh" --repo-root "${REPO_ROOT}")"
"${KUMACTL}" version
```

**Gate**: kumactl version output matches the repo HEAD.

## Phase 1 - initialize run

```bash
RUNS_DIR="${DATA_DIR}/runs"
RUN_ID="$(date +%Y%m%d-%H%M%S)-manual"

"$SKILL_DIR/scripts/init-run.sh" --runs-dir "${RUNS_DIR}" "${RUN_ID}"
RUN_DIR="${RUNS_DIR}/${RUN_ID}"
```

Suite resolution uses the three-step order (directory suite, legacy `.md` file, literal path). Set `SUITE_DIR` and `SUITE_FILE` accordingly.

Fill `run-metadata.yaml` before touching the cluster.

**Gate**: `run-metadata.yaml` exists and has profile, feature scope, and environment filled in.

## Phase 2 - bootstrap cluster

Read `references/cluster-setup.md` before starting this phase.

Pick a profile and start the cluster:

```bash
"$SKILL_DIR/scripts/cluster-lifecycle.sh" --repo-root "${REPO_ROOT}" single-up kuma-1
# or
"$SKILL_DIR/scripts/cluster-lifecycle.sh" --repo-root "${REPO_ROOT}" global-two-zones-up kuma-1 kuma-2 kuma-3 zone-1 zone-2
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
"$SKILL_DIR/scripts/preflight.sh" \
  --kubeconfig "${HOME}/.kube/kind-kuma-1-config" \
  --run-dir "${RUN_DIR}" \
  --repo-root "${REPO_ROOT}"

"$SKILL_DIR/scripts/capture-state.sh" \
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

1. Read `${SUITE_DIR}/suite.md` for overview, group table, and execution contract.
2. Before G1, apply each baseline manifest listed in the baseline table from `${SUITE_DIR}/baseline/`.
3. Before each group, read the group file from `${SUITE_DIR}/groups/` using the path from the group table.
4. After completing a group, the group file content can be dropped from context.

For legacy single-file suites: read the entire suite file as before.

For each test step:

1. Create or copy manifest to a working file.
2. Apply through the tracked script only.
3. Collect runtime artifacts (log ad-hoc commands with `"$SKILL_DIR/scripts/record-command.sh"`).
4. Write result into the report.

```bash
"$SKILL_DIR/scripts/apply-tracked-manifest.sh" \
  --run-dir "${RUN_DIR}" \
  --kubeconfig "${HOME}/.kube/kind-kuma-1-config" \
  --manifest "<manifest-path>" \
  --step "<step-name>"
```

After each test group, update `run-status.yaml` with `last_completed_group`, `next_planned_group`, and pass/fail counts.

On first unexpected failure, go to Phase 5.

**Gate**: all planned tests have pass/fail entries in the report.

## Phase 5 - failure handling

Read `references/troubleshooting.md` for known failure modes.

1. Stop progression.
2. Capture immediate state snapshot.
3. Document expected vs observed.
4. Classify the issue (manifest, environment, product bug).
5. Continue only when classification is explicit.

```bash
"$SKILL_DIR/scripts/capture-state.sh" \
  --kubeconfig "${HOME}/.kube/kind-kuma-1-config" \
  --run-dir "${RUN_DIR}" \
  --label "failure-<test-id>"
```

## Phase 6 - closeout

```bash
"$SKILL_DIR/scripts/capture-state.sh" \
  --kubeconfig "${HOME}/.kube/kind-kuma-1-config" \
  --run-dir "${RUN_DIR}" \
  --label "postrun"

"$SKILL_DIR/scripts/report-compactness-check.sh" \
  --report "${RUN_DIR}/reports/manual-test-report.md"
```

**Gate**: all of these are true before marking the run complete:

- Command log is complete
- Manifest index includes every apply
- Report has pass or fail for all planned tests
- Failures include triage details and artifact paths
- Report compactness check passes

Cluster teardown is optional. Leave clusters running if another run follows immediately. Otherwise:

```bash
KIND_CLUSTER_NAME=kuma-1 make k3d/stop
# For multi-zone, also stop kuma-2, kuma-3
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
  "$SKILL_DIR/scripts/cluster-lifecycle.sh" --repo-root "${REPO_ROOT}" single-up kuma-1
```
