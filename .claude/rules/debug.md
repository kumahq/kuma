# Debugging and local development

## k3d cluster setup

```bash
make k3d/cluster/start CLUSTER=kuma-1
export KUBECONFIG=~/.kube/k3d-kuma-1.yaml
```

Kubeconfig: `~/.kube/k3d-<name>.yaml`. k3d context: `k3d-<name>`.

### Build, load, and deploy

```bash
# One step: build images, load into k3d, clean previous release, helm install, wait for CP
make k3d/cluster/deploy/helm CLUSTER=kuma-1

# Or step by step:
make k3d/cluster/load CLUSTER=kuma-1
make k3d/cluster/deploy/helm/upgrade k3d/cluster/deploy/wait/cp CLUSTER=kuma-1
```

### Variables

| Variable | Default | Purpose |
|:---------|:--------|:--------|
| `CLUSTER` | `kuma` | Cluster name (accepts `kuma`, digits, or `kuma-<N>`) |
| `KUMA_MODE` | `zone` | `zone` or `global` |
| `K3D_HELM_DEPLOY_NO_CNI` | unset | Set `true` to skip CNI (lighter for local dev) |
| `K3D_HELM_DEPLOY_ADDITIONAL_SETTINGS` | unset | Extra helm values, space-separated |
| `K3D_DEPLOY_HELM_DONT_CLEAN` | unset | Set to skip cleaning previous release |

### Deploy with custom settings

```bash
make k3d/cluster/deploy/helm CLUSTER=kuma-1 \
  K3D_HELM_DEPLOY_NO_CNI=true \
  K3D_HELM_DEPLOY_ADDITIONAL_SETTINGS="dataPlane.features.unifiedResourceNaming=true"
```

Other targets: `k3d/cluster/deploy/kumactl` (deploy via CLI), `k3d/cluster/deploy/demo` (demo app), `k3d/cluster/stop` (delete cluster), `k3d/destroy` (delete all clusters + network).

Old target names still work but print a deprecation warning. Targets `k3d/deploy/kuma` and `k3d/restart` will error with the new name.

### Skaffold dev loop

Hot-reload: watches code changes, rebuilds, redeploys. Config: `skaffold.yaml`.

```bash
make k3d/cluster/start CLUSTER=kuma-1
export KUBECONFIG=~/.kube/k3d-kuma-1.yaml
make dev/fetch-demo
skaffold dev
```

`skaffold debug` exposes a dlv port (logged on startup) for remote debugging from GoLand/VS Code.

## Envoy admin API

Access via the sidecar container on port 9901:

```bash
# Full config dump
kubectl exec deploy/<name> -c kuma-sidecar -- wget -qO- localhost:9901/config_dump

# Filter by section (replace <Section> with Listeners, ClustersConfigDump, Routes, etc.)
kubectl exec deploy/<name> -c kuma-sidecar -- \
  wget -qO- localhost:9901/config_dump | \
  jq '.configs[] | select(."@type" | contains("<Section>"))'
```

Other endpoints: `/stats` (Envoy metrics), `/clusters` (upstream info), `/server_info` (version).

## Control plane

```bash
kubectl logs -n kuma-system deploy/kuma-control-plane -f
kubectl exec -n kuma-system deploy/kuma-control-plane -- \
  wget -qO- localhost:5681/meshes/default/dataplanes
```

The CP REST API runs on port 5681. See `pkg/api-server/` for available endpoints.

## Common tasks

### Enable unified resource naming

Requires both helm value AND mesh patch:

```bash
# Set during deploy via K3D_HELM_DEPLOY_ADDITIONAL_SETTINGS, then:
kubectl patch mesh default --type merge \
  -p '{"spec":{"meshServices":{"mode":"Exclusive"}}}'
```

### Test a policy change locally

1. Build, load, deploy: `make k3d/cluster/deploy/helm CLUSTER=kuma-1`
2. Create a test namespace with sidecar injection: `kubectl label namespace <ns> kuma.io/sidecar-injection=enabled`
3. Deploy a test workload and apply the policy
4. Inspect Envoy config to verify xDS changes (see Envoy admin API above)
5. Clean up: `make k3d/cluster/stop CLUSTER=kuma-1`

## CPU limit workaround

k3d init containers sometimes get throttled on local machines. Set `K3D_HELM_DEPLOY_NO_CNI=true` to reduce resource pressure. If still slow, comment out CPU limits in `pkg/plugins/runtime/k8s/webhooks/injector/injector.go` (search for `NewScaledQuantity(100, kube_api.Milli)`). Revert before committing.

## Test output

Filter noisy macOS linker warnings: `| grep -vE 'LC_DYSYMTAB|#'`
