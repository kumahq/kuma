# Contents

1. [Prerequisites](#prerequisites)
2. [Kubeconfig mapping](#kubeconfig-mapping)
3. [Profiles](#profiles)
4. [Build and deploy local code changes](#build-and-deploy-local-code-changes)
5. [CRD updates](#crd-updates)
6. [Baseline readiness validation](#baseline-readiness-validation)

---

# Cluster setup

Cluster lifecycle commands for local manual testing with k3d.

## Prerequisites

- Docker daemon running
- `k3d`, `kubectl`, `helm`, `make` installed
- `REPO_ROOT` resolved (via `--repo` flag or auto-detected from cwd)

## Kubeconfig mapping

| Cluster  | Role                  | Kubeconfig file                    |
| -------- | --------------------- | ---------------------------------- |
| `kuma-1` | single-zone or global | `${HOME}/.kube/kind-kuma-1-config` |
| `kuma-2` | zone-1                | `${HOME}/.kube/kind-kuma-2-config` |
| `kuma-3` | zone-2                | `${HOME}/.kube/kind-kuma-3-config` |

Use explicit absolute paths. Do not rely on implicit context switching.

## Profiles

### Single-zone

```bash
"${CLAUDE_SKILL_DIR}/scripts/cluster-lifecycle.sh" --repo-root "${REPO_ROOT}" single-up kuma-1
```

Manual equivalent:

```bash
KIND_CLUSTER_NAME=kuma-1 make k3d/start
K3D_HELM_DEPLOY_NO_CNI=true KIND_CLUSTER_NAME=kuma-1 make k3d/deploy/helm
```

Stop:

```bash
KIND_CLUSTER_NAME=kuma-1 make k3d/stop
```

### Global + one zone

```bash
"${CLAUDE_SKILL_DIR}/scripts/cluster-lifecycle.sh" --repo-root "${REPO_ROOT}" global-zone-up kuma-1 kuma-2 zone-1
```

Manual equivalent for global:

```bash
KIND_CLUSTER_NAME=kuma-1 make k3d/start
KUBECONFIG="${HOME}/.kube/kind-kuma-1-config" \
  K3D_HELM_DEPLOY_NO_CNI=true \
  KIND_CLUSTER_NAME=kuma-1 \
  KUMA_MODE=global \
  K3D_HELM_DEPLOY_ADDITIONAL_SETTINGS="controlPlane.mode=global controlPlane.globalZoneSyncService.type=NodePort" \
  make k3d/deploy/helm
```

Manual equivalent for zone:

```bash
GLOBAL_NODE_IP=$(docker inspect k3d-kuma-1-server-0 \
  -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}')

GLOBAL_KDS_PORT=$(KUBECONFIG="${HOME}/.kube/kind-kuma-1-config" kubectl get svc \
  -n kuma-system kuma-global-zone-sync \
  -o jsonpath='{.spec.ports[?(@.name=="global-zone-sync")].nodePort}')

GLOBAL_KDS="grpcs://${GLOBAL_NODE_IP}:${GLOBAL_KDS_PORT}"

KIND_CLUSTER_NAME=kuma-2 make k3d/start
KUBECONFIG="${HOME}/.kube/kind-kuma-2-config" \
  K3D_HELM_DEPLOY_NO_CNI=true \
  KIND_CLUSTER_NAME=kuma-2 \
  KUMA_MODE=zone \
  K3D_HELM_DEPLOY_ADDITIONAL_SETTINGS="controlPlane.mode=zone controlPlane.zone=zone-1 controlPlane.kdsGlobalAddress=${GLOBAL_KDS} controlPlane.tls.kdsZoneClient.skipVerify=true" \
  make k3d/deploy/helm
```

### Global + two zones

```bash
"${CLAUDE_SKILL_DIR}/scripts/cluster-lifecycle.sh" --repo-root "${REPO_ROOT}" global-two-zones-up kuma-1 kuma-2 kuma-3 zone-1 zone-2
```

Stop all:

```bash
KIND_CLUSTER_NAME=kuma-1 make k3d/stop
KIND_CLUSTER_NAME=kuma-2 make k3d/stop
KIND_CLUSTER_NAME=kuma-3 make k3d/stop
```

## Build and deploy local code changes

The lifecycle script handles build, image load, and helm deploy in one step.

For manual control:

```bash
make build
K3D_HELM_DEPLOY_NO_CNI=true KIND_CLUSTER_NAME=kuma-1 make k3d/deploy/helm
```

## CRD updates

After CRD/schema changes, force-update CRDs:

```bash
kubectl --kubeconfig "${HOME}/.kube/kind-kuma-1-config" \
  apply --server-side --force-conflicts \
  -f "${REPO_ROOT}/deployments/charts/kuma/crds/"
```

## Gateway API CRDs

Suites that test builtin gateways (MeshGateway, GatewayClass, HTTPRoute) or compare builtin vs delegated gateways need the Kubernetes Gateway API CRDs installed. These are NOT included in Kuma's CRD bundle - they come from the upstream `gateway-api` project.

Check and install:

```bash
# Check if installed
"${CLAUDE_SKILL_DIR}/scripts/install-gateway-api-crds.sh" \
  --kubeconfig "${HOME}/.kube/kind-kuma-1-config" \
  --repo-root "${REPO_ROOT}" \
  --check-only

# Install (idempotent - skips if already present)
"${CLAUDE_SKILL_DIR}/scripts/install-gateway-api-crds.sh" \
  --kubeconfig "${HOME}/.kube/kind-kuma-1-config" \
  --repo-root "${REPO_ROOT}"
```

The script extracts the version from `go.mod` (`sigs.k8s.io/gateway-api`) to stay in sync with the Kuma codebase. It installs the standard CRDs: GatewayClass, Gateway, HTTPRoute, ReferenceGrant.

Install on every cluster that needs it (in multi-zone setups, zones running builtin gateways need the CRDs).

## Baseline readiness validation

Before test execution:

```bash
kubectl --kubeconfig "${KUBECONFIG}" get pods -n kuma-system
kubectl --kubeconfig "${KUBECONFIG}" wait --for=condition=Ready pod -n kuma-system --all --timeout=180s
kubectl --kubeconfig "${KUBECONFIG}" get mesh default -o yaml
```

For workload namespaces, verify sidecars are present (`2/2` for app + sidecar).

## Notes

- `cluster-lifecycle.sh` auto-generates a temporary `mk/metallb-k3d-kuma-<n>.yaml` when missing for numeric cluster names like `kuma-3`.
- Performance toggles are documented in `references/workflow.md` (performance toggles section).
