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
* Good, because it's easier to adjust resources for a ZoneProxy.
* Bad, because any Helm upgrade that touches the Deployment spec triggers a zone proxy restart
* Bad, because inconsistent with the regular Dataplane deployment model

### Option 2: Sidecar injection with a dummy main container

The zone proxy Deployment contains a **dummy main container** (a no-op process) as the main container.
`kuma-dp` is injected as a sidecar by the existing Kuma sidecar injection webhook, the same way it
is injected into regular workload pods.

```yaml
containers:
  - name: pause          # dummy main container, never changes
    image: gcr.io/google_containers/pause:latest
  # kuma-dp injected here by webhook:
  - name: kuma-sidecar
    image: kuma-dp:latest
    args: ["run", "--cp-address=..."]
```

The Deployment spec is owned by Helm and contains only the dummy container.
When Kuma is upgraded (new `kuma-dp` image), the Deployment spec is unchanged → **no rollout**.

The Deployment is labeled to opt into sidecar injection:
```yaml
metadata:
  labels:
    kuma.io/sidecar-injection: enabled
    k8s.kuma.io/zone-proxy-type: ingress  # or: egress
```

The webhook detects the `k8s.kuma.io/zone-proxy-type` label and configures `kuma-dp` with the
appropriate `listeners` (per MADR 095) instead of the regular inbound/outbound configuration.

* Good, because Helm upgrades no longer restart zone proxies
* Good, because consistent deployment model with regular Dataplanes
* Good, because reuses existing sidecar injection and bootstrap machinery
* Good, because aligns with MADR 095 (zone proxies as Dataplanes with `listeners`)
* Bad, because requires a dummy container, which is non-obvious to users inspecting the pod
* Bad, because zone proxy update cadence decouples from Kuma upgrade cadence
  (operators must trigger restarts explicitly to pick up new `kuma-dp`)
* Bad, because configuring sidecar resources requires `ContainerPatch` or CP injector config
  rather than a direct `resources:` field in the pod spec (see "Configuring sidecar resources" below)

#### Waiting container image

The dummy main container must run indefinitely without consuming resources.
The following candidates are evaluated against three criteria: **image size**, **pull overhead**, and **security surface**.

| Image | Compressed size | Pull overhead | Shell | Notes |
|---|---|---|---|---|
| `registry.k8s.io/pause:3.10` | ~300 KB | Mostly none since it's used but sometimes might be required | No | Purpose-built for this use case; used by K8s itself for pod sandboxes |
| `gcr.io/distroless/static-debian12:nonroot` | ~2 MB | Pull required | No | No package manager, no shell; well-maintained by Google |
| `cgr.dev/chainguard/static:latest` | ~1.5 MB | Pull required | No | Minimal, signed images; depends on third-party registry availability |
| `busybox` | ~4 MB | Pull required | Yes | Shell increases attack surface; unnecessary for a no-op container |
| `alpine` | ~8 MB | Pull required | Yes | Package manager + shell; overkill |

**Decision: `registry.k8s.io/pause:3.10`**

- Usually present on every Kubernetes node — the kubelet pre-pulls it to create pod network namespaces. No registry round-trip at zone proxy pod startup.
- Smallest possible binary (~300 KB static ELF that calls `pause(2)` in a loop).
- No shell, no tools, no package manager — minimal CVE surface.
- Maintained by the Kubernetes project with the same release cadence as K8s itself.
- The tag is pinned and updated alongside the supported K8s version matrix:
  `pause:3.10` ships with K8s 1.31+.

The Helm chart exposes the image as a configurable value:

```yaml
meshes:
  - name: default
    ingress:
      ...
      image: registry.k8s.io/pause:3.10 # user can override
```

#### Configuring sidecar resources

Because `kuma-dp` is injected by the webhook rather than defined in the Deployment spec,
resource requests and limits cannot be set directly on the pod template's container list.
Two mechanisms are available, in order of precedence (highest first):

##### 1. Per-zone-proxy: ContainerPatch + annotation

Create a `ContainerPatch` in `kuma-system` with a JSON patch targeting the sidecar container,
then reference it via the `kuma.io/container-patches` annotation on the Deployment pod template:

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

```yaml
# Deployment pod template annotation
annotations:
  kuma.io/container-patches: zone-ingress-resources
```

##### 2. Zone-wide default: CP injector configuration

The CP injector applies a default resource spec to all injected sidecar containers in the zone,
including zone proxy sidecars. Configure via environment variables on the CP:

```shell
KUMA_INJECTOR_SIDECAR_CONTAINER_RESOURCES_REQUESTS_CPU=50m
KUMA_INJECTOR_SIDECAR_CONTAINER_RESOURCES_REQUESTS_MEMORY=64Mi
KUMA_INJECTOR_SIDECAR_CONTAINER_RESOURCES_LIMITS_CPU=1000m
KUMA_INJECTOR_SIDECAR_CONTAINER_RESOURCES_LIMITS_MEMORY=512Mi
```

This acts as the fallback when no `ContainerPatch` is applied to a specific zone proxy Deployment.

## Decision

Chosen option: **Option 2** (sidecar injection with a dummy main container).

This eliminates zone proxy restarts on Helm upgrades without requiring a minimum K8s version.
It also aligns zone proxy deployment with the standard Kuma Dataplane model established by MADR 095.

The dummy container uses the `pause` image, which is usually present on every Kubernetes node
(used by Kubernetes itself for pod sandboxes), adding zero pull overhead.

The sidecar injection webhook is extended to recognize the `k8s.kuma.io/zone-proxy-type` label and
generate the appropriate `kuma-dp` arguments for zone proxy mode instead of regular sidecar mode.

### Update behavior

After a Kuma upgrade, existing zone proxy pods are **not** automatically restarted.
Operators can trigger a rolling restart explicitly:

```bash
kubectl rollout restart deployment/<zone-proxy-deployment> -n kuma-system
```

This gives operators control over the update window and avoids unplanned cross-zone traffic disruption.

## Security implications and review

No new security surface. The sidecar injection webhook already handles authentication and RBAC.
The `pause` container runs with no capabilities and no network ports.

## Reliability implications

- Zone proxy restart is decoupled from Helm upgrade, reducing upgrade-time risk
- Operators must explicitly restart zone proxy pods to pick up new `kuma-dp` images;
  running an old `kuma-dp` version after a Kuma upgrade is supported for the duration of
  the same compatibility window as regular Dataplanes

## Implications for Kong Mesh

None.

## Notes

- MADR 094: Zone Proxy Deployment Model (Helm schema, per-mesh templates)
- MADR 095: Mesh-Scoped Zone Proxies (Dataplane `listeners` array)
- MADR 096: Syncing Zone Ingress Address Across Zones
