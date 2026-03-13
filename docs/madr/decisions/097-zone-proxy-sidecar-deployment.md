# Zone Proxy Deployment as a Regular Sidecar

* Status: accepted

Technical Story: https://github.com/Kong/kong-mesh/issues/9182

## Context and Problem Statement

Zone proxies (zone ingress and zone egress) are currently deployed as standalone Kubernetes Deployments
where `kuma-dp` is the **main container** in the pod. This differs from how Kuma deploys proxies for
regular workloads, where `kuma-dp` is injected as a sidecar alongside the application container.

This standalone model creates two problems:

1. **Helm upgrades restart zone proxies**: Any change to the Deployment spec triggers a
   Kubernetes rollout. Since `kuma-dp` is the main container, this restarts the zone proxy, which is unexpected in this situation. For regular workloads with sidecar injection, Helm upgrades do not affect running pods because the Deployment spec contains only the application container (`kuma-dp` is injected at admission time and is not part of the Deployment spec).

2. **Inconsistent deployment model**: Zone proxies are the only Kuma-managed proxies not deployed via
   sidecar injection. This creates a special code path for zone proxies in Helm templates and the
   control plane, and makes zone proxies behave differently from regular Dataplanes.

## Design

### Option 1: Keep standalone Deployment (status quo)

The zone proxy Deployment contains only `kuma-dp` as the main container.
The Helm chart manages the full pod spec, including the `kuma-dp` image and arguments.

* Good, because no changes required
* Good, because simple - no injection machinery involved
* Good, because it's easier to adjust resources for a ZoneProxy
* Bad, because any Helm upgrade that touches the Deployment spec triggers a zone proxy restart
* Bad, because inconsistent with the regular Dataplane deployment model

### Option 2: Sidecar injection with a dummy main container

The zone proxy Deployment contains a **dummy main container** (a no-op process) as the main container.
`kuma-dp` is injected as a sidecar by the existing Kuma sidecar injection webhook, the same way it
is injected into regular workload pods.

```yaml
containers:
  - name: pause          # dummy main container; image rarely changes across Kuma upgrades
    image: registry.k8s.io/pause:3.10  # override via Helm if image is not available
  # kuma-dp injected here by webhook:
  - name: kuma-sidecar
    image: kumahq/kuma-dp:v2.13.2
    args: ["run", "--cp-address=..."]
```

The Deployment spec is owned by Helm and contains only the dummy container.
When Kuma is upgraded (new `kuma-dp` image), the Deployment spec is unchanged → **no rollout**.

The pod template is labeled to opt into sidecar injection:
```yaml
spec:
  template:
    metadata:
      labels:
        kuma.io/sidecar-injection: enabled
```

For the `combinedProxies` Deployment (which runs both ingress and egress roles), the same
`kuma.io/sidecar-injection: enabled` label applies. The zone proxy type (ingress, egress, or
combined) is determined by the `k8s.kuma.io/zone-proxy-type` label on the **Service**, not the
Deployment pod template (per MADR 095).

**Webhook namespace constraint**: The default Helm installation configures the sidecar injector
webhook with a `namespaceSelector` that excludes the control-plane release namespace
(`kubernetes.io/metadata.name NotIn [{{ .Release.Namespace }}]`, see
`deployments/charts/kuma/templates/cp-webhooks-and-secrets.yaml`). Because zone proxies currently
run in `kuma-system`, sidecar injection will not trigger for them unless the webhook is adjusted
to include that namespace. The implementation must update the webhook `namespaceSelector` to
allow injection in `kuma-system`.

The webhook detects the zone proxy type from the `k8s.kuma.io/zone-proxy-type` label on the
associated Service and configures `kuma-dp` with the appropriate zone proxy arguments instead of
the regular inbound/outbound configuration.

* Good, because Helm upgrades no longer restart zone proxies
* Good, because consistent deployment model with regular Dataplanes
* Good, because reuses existing sidecar injection and bootstrap machinery
* Good, because aligns with the mesh-scoped zone proxy model (Dataplanes with `listeners`)
* Good, because resource requests/limits use pod-level resources (`spec.resources`, alpha in
  K8s 1.32, beta and enabled by default from K8s 1.34); `ContainerPatch` is available as
  fallback on K8s 1.32–1.33 or clusters with the feature gate disabled
* Bad, because requires a dummy container, which is non-obvious to users inspecting the pod
* Bad, because zone proxy update cadence decouples from Kuma upgrade cadence
  (operators must trigger restarts explicitly to pick up new `kuma-dp`);
  this is the same behavior as regular Dataplanes — operators already restart workload pods
  after upgrades to pick up new injected `kuma-dp` versions
* Bad, because the default webhook excludes `kuma-system`; the webhook `namespaceSelector` must
  be adjusted to enable injection in the control-plane namespace

#### Waiting container image

**Decision: `registry.k8s.io/pause:3.10`**

The `pause` container is bundled with every Kubernetes cluster
([kubeadm defaults to `registry.k8s.io` as `DefaultImageRepository`](https://github.com/kubernetes/kubernetes/blob/master/cmd/kubeadm/app/apis/kubeadm/v1beta4/defaults.go#L44)):
- Very small (514 kB) and purpose-built for doing nothing
- Version (`3.10`) is pinned and changes independently of Kuma upgrades → the Deployment spec is
  stable between Kuma releases → no unwanted rollouts on `helm upgrade`
- Already present on every node (Kubernetes uses it for pod sandboxes)

Using `kumahq/kuma-dp:<version>` with a `pause` subcommand was considered but **rejected**:
because that image is version-tagged, any Kuma upgrade would change the Deployment spec and
trigger a rollout, negating the main benefit of this design.

A dedicated `kumahq/pause` image was also considered but rejected — it would be functionally
identical to `registry.k8s.io/pause` with extra maintenance overhead (separate pipeline, release
cadence, operators still need to mirror it).

The default registry (`registry.k8s.io`) varies per distribution:

| Distribution | Image reference |
|:-------------|:----------------|
| GKE | `gke.gcr.io/pause:3.8` |
| k3d / Rancher | `docker.io/rancher/mirrored-pause:3.6` |
| upstream kubeadm | `registry.k8s.io/pause:3.10` |
| others | ... |

The Helm chart exposes an image override per mesh proxy for environments where
`registry.k8s.io` is not reachable or images are mirrored under a different path

```yaml
meshes:
  - name: default
    ingress:
      enabled: true
      image: registry.k8s.io/pause:3.10  # override if image is mirrored elsewhere
```

#### Configuring sidecar resources

Because `kuma-dp` is injected by the webhook rather than defined in the Deployment spec,
resource requests and limits cannot be set directly on a named container in the Deployment spec.
**Primary: pod-level resources** — Use the `resources` field inside each proxy section
(`ingress`, `egress`, `combinedProxies`) — already defined in MADR 094. These map to
[pod-level resources](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#example-2)
(`spec.resources`) on the pod template, applying to all containers in the pod.
Pod-level resources were alpha in K8s 1.32 and graduated to beta (enabled by default) in K8s 1.34.
Note: K8s 1.31 is already end-of-life; the oldest actively supported K8s version is 1.32.

```yaml
meshes:
  - name: default
    ingress:
      enabled: true
      resources:
        requests:
          cpu: 100m
          memory: 128Mi
        limits:
          cpu: 500m
          memory: 256Mi
```

**Fallback: `ContainerPatch`** — Operators who require per-container resource control, or who
run on clusters with the `PodLevelResources` feature gate explicitly disabled, can manually
create a `ContainerPatch` targeting the injected sidecar:

```yaml
apiVersion: kuma.io/v1alpha1
kind: ContainerPatch
metadata:
  name: zone-ingress-resources
  namespace: kuma-system
spec:
  sidecarPatch:
    - op: replace
      path: /resources
      value: '{"requests":{"cpu":"100m","memory":"128Mi"},"limits":{"cpu":"500m","memory":"256Mi"}}'
```

Reference it via the `kuma.io/container-patches` annotation on the Deployment pod template:

```yaml
annotations:
  kuma.io/container-patches: zone-ingress-resources
```

## Decision

Chosen option: **Option 2** (sidecar injection with a dummy main container).

This eliminates zone proxy restarts on Helm upgrades and aligns zone proxy deployment with the
standard Kuma Dataplane model (mesh-scoped zone proxies with Dataplane `listeners`).
Resource requests/limits use pod-level resources (`spec.resources`), alpha in K8s 1.32 and
beta (enabled by default) from K8s 1.34. For K8s 1.32–1.33 or clusters with the feature gate
disabled, `ContainerPatch` is the fallback mechanism.

The dummy container runs `registry.k8s.io/pause:3.10` — bundled with every Kubernetes cluster,
with a version that changes independently of Kuma releases. Operators can override the image via
the `image` Helm value per mesh proxy when `registry.k8s.io` is not reachable. No additional images or registry dependencies are introduced beyond what Kubernetes already uses internally.

The sidecar injection webhook is extended to look up the `k8s.kuma.io/zone-proxy-type` label on
the associated Service and generate the appropriate `kuma-dp` arguments for zone proxy mode
instead of regular sidecar mode.

#### reachableBackends

Because the injected `kuma-dp` sidecar is a regular Dataplane, it would by default receive
outbound configuration for all backends in the mesh. Zone proxies must not get populated
outbounds. The Deployment is annotated with an empty `reachableBackends` list so the webhook
instructs `kuma-dp` to request no outbound routes:

```yaml
spec:
  template:
    metadata:
      annotations:
        kuma.io/reachable-backends: '{"refs":[]}'
```

This is the same mechanism available to regular workloads. Using an empty list is
intentional, zone proxies route via the zone ingress/egress path, not direct mesh outbounds.

### Update behavior

After a Kuma upgrade, existing zone proxy pods are **not** automatically restarted.
Operators can trigger a rolling restart explicitly:

```bash
kubectl rollout restart deployment/<zone-proxy-deployment> -n kuma-system
```

This gives operators control over the update window and avoids unplanned cross-zone traffic disruption.

## Security implications and review

No new security surface beyond the existing mutating webhook, which operates within Kubernetes API
server authentication and RBAC; the webhook only mutates objects that have already been admitted
by the API server.
The `pause` container is configured with a minimal-privilege `securityContext` (dropping all Linux
capabilities and exposing no network ports).

## Reliability implications

- Zone proxy restart is decoupled from Helm upgrade, reducing upgrade-time risk
- Operators must explicitly restart zone proxy pods to pick up new `kuma-dp` images;
  running an old `kuma-dp` version after a Kuma upgrade is supported for the duration of
  the same compatibility window as regular Dataplanes

## Notes

- MADR 094: Zone Proxy Deployment Model (Helm schema, per-mesh templates)
- MADR 095: Mesh-scoped zone proxies (Dataplane `listeners` array, `k8s.kuma.io/zone-proxy-type` Service label)
- MADR 096: Syncing zone ingress address across zones
