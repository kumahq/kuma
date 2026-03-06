# Contents

1. [Pods do not schedule](#1-pods-in-non-system-namespaces-do-not-schedule)
2. [Schema rejected](#2-resource-schema-rejected-before-business-validation)
3. [Wrong kumactl](#3-wrong-kumactl-behavior)
4. [Workload crashes](#4-dependency-workload-crashes-after-start)
5. [Missing xDS config](#5-expected-xds-config-not-visible)
6. [Stale after redeploy](#6-changes-not-reflected-after-redeploy)
7. [Multi-zone sync](#7-multi-zone-sync-unclear)
8. [Leftovers](#8-unexpected-leftovers-after-redeploy)
9. [KDS connection](#9-zone-cannot-connect-to-global-kds)
10. [MetalLB file missing](#10-k3dstart-fails-for-kuma-3-with-missing-metallb-file)
11. [Failure triage checklist](#failure-triage-checklist)

---

# Troubleshooting

Known local failure modes and fixes.

For policy matching and inspect workflow, use `references/mesh-policies.md`.

## 1) Pods in non-system namespaces do not schedule

Symptoms: workloads stay Pending, CNI-related scheduling issues.

Fix:

```bash
K3D_HELM_DEPLOY_NO_CNI=true KIND_CLUSTER_NAME=kuma-1 make k3d/deploy/helm
```

## 2) Resource schema rejected before business validation

Symptoms: `kubectl apply` fails with CRD required-field or enum errors.

Fix:

1. Run server-side dry-run with `"$SKILL_DIR/scripts/validate-manifest.sh"`.
2. Confirm CRDs are current:

```bash
kubectl --kubeconfig "${KUBECONFIG}" apply --server-side --force-conflicts \
  -f deployments/charts/kuma/crds/
```

3. Retry validation.

## 3) Wrong kumactl behavior

Symptoms: unknown resource types, missing command support for newer resources.

Fix:

```bash
make build/kumactl
KUMACTL="$("$SKILL_DIR/scripts/find-local-kumactl.sh")"
"${KUMACTL}" version
```

Use `"${KUMACTL}"` explicitly in all test commands.

## 4) Dependency workload crashes after start

Symptoms: supporting workload pod CrashLoopBackOff, startup logs show invalid configuration.

Fix: fix config according to pod logs, then re-apply tracked manifest.

## 5) Expected xDS config not visible

Symptoms: expected filters/clusters/listeners are missing.

Checks:

- Verify policy target selectors match the intended dataplanes
- Verify service protocol annotations when HTTP-specific behavior is expected
- Verify listeners are generated in expected mode (HCM vs TCP proxy)

## 6) Changes not reflected after redeploy

Symptoms: xDS/config does not match latest code.

Fix:

1. Redeploy with local build target.
2. Restart test workloads to refresh sidecars and certs.
3. Capture state snapshot and compare timestamps.

## 7) Multi-zone sync unclear

Symptoms: resource present on global but absent on zone.

Checks:

- Zone connection status is Online
- Correct KDS global address configured
- `kuma.io/origin` and `kuma.io/display-name` labels preserved

## 8) Unexpected leftovers after redeploy

Symptoms: behavior looks like old resources are still active.

Root cause: default lifecycle mode keeps helm release and namespace for speed.

Fix:

```bash
HARNESS_HELM_CLEAN=1 \
  "$SKILL_DIR/scripts/cluster-lifecycle.sh" single-up kuma-1
```

Then rerun preflight and continue tests.

## 9) Zone cannot connect to global KDS

Symptoms: zone CP logs show `dial tcp <cluster-ip>:5685: i/o timeout`.

Root cause: using global cluster service ClusterIP in `controlPlane.kdsGlobalAddress`.

Fix: expose global sync service as NodePort and use `grpcs://<global-node-ip>:<node-port>` instead of the service ClusterIP.

## 10) k3d/start fails for kuma-3 with missing MetalLB file

Symptoms: error references missing `mk/metallb-k3d-kuma-3.yaml`.

Fix: use `"$SKILL_DIR/scripts/cluster-lifecycle.sh"`, which auto-generates a temporary manifest for numeric names `kuma-<n>`.

## Failure triage checklist

When a test fails:

1. Run `"$SKILL_DIR/scripts/capture-state.sh"` immediately with label `failure-<test-id>`.
2. Record exact failing command in command log.
3. Record expected vs observed behavior.
4. Classify root cause: manifest issue, cluster/infrastructure issue, or product bug.
5. Do not continue until classification is explicit.
