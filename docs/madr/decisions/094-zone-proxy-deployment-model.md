# Zone Proxy Deployment Model

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/9030

## Context and Problem Statement

Currently, zone proxies are **global-scoped** resources, meaning a single ZoneIngress or ZoneEgress instance handles traffic for all meshes in a zone.
This global nature creates fundamental limitations:

1. **Cannot issue MeshIdentity for zone egress**: MeshIdentity is mesh-scoped, but a global zone egress serves multiple meshes.
   This creates identity conflicts and prevents proper mTLS certificate issuance.
   See [MADR 090](090-zone-egress-identity.md) for detailed analysis.

2. **Cannot apply policies on zone proxies**: Kuma policies (MeshTrafficPermission, MeshTimeout, etc.) are mesh-scoped.
   A global zone proxy cannot be targeted by mesh-specific policies, limiting observability and traffic control for cross-zone communication.

To resolve these limitations, zone proxies are being changed to **mesh-scoped** resources represented as Dataplane resources with specific tags.
This architectural change requires revisiting the deployment model for zone proxies.

**Scope of this document**: This MADR focuses on **deployment tooling** — how users deploy zone proxies via Helm, Konnect UI, and Terraform.
The resource model (Dataplane representation, labels, tokens, workload identity) is covered in a separate MADR. [^1]

**Single-mesh focus**: This document assumes **single-mesh-per-zone as the default** deployment pattern.
For multi-mesh scenarios, deploy additional zone proxies using separate Helm releases with a dedicated zone-proxy chart. This can be packaged either as an independent `kuma-zone-proxy` chart (separate release cycle, full independence) or as a subchart of the main kuma chart (single repo, tighter coupling). See the multi-mesh deployment guide for details.

This document addresses the following questions:

1. Should we continue supporting `kuma.io/ingress-public-address` annotation?
2. What should be the default Helm installation behavior for zone proxies?

Note: Whether zone ingress and egress share a single deployment is addressed in a separate MADR. [^2]

### Decision Summary

| Tooling Decision | Choice |
|------------------|--------|
| Per-mesh Services | **Yes** - each mesh gets its own Service/LoadBalancer for mTLS isolation |
| Namespace placement | **kuma-system** |
| Deployment mechanism | **Helm-managed** (current pattern extended for mesh-scoped zone proxies) |

| Question | Decision |
|----------|----------|
| 1. Support ingress-public-address? | **Yes** - keep as escape hatch |
| 2. Default Helm behavior? | `zoneProxy.enabled: false` — explicit opt-in, like `ingress.enabled` |

Note: Resource model (Dataplane representation, labels, tokens, workload identity) is in a separate MADR. [^1]
Note: Zone proxy deployment topology (shared vs separate ingress/egress) is addressed in a separate MADR. [^2]

### Document Structure

This document is organized in two parts:

1. **Tooling and User Flows** - Describes how users will deploy zone proxies using different tools (Konnect UI, Helm, Terraform). This covers the UX and configuration experience.

2. **Questions 1-2** - Answers deployment-related design questions. Each question analyzes options and recommends a decision.

### Out of Scope

The following topics are deferred to the resource model MADR:

- Dataplane representation (fields, labels, `kuma.io/proxy-type` tag)
- `kuma.io/workload` annotation and auto-generation
- Workload identity model
- Token model (zone tokens → DP tokens transition)
- Universal deployment specifics (VM/bare metal)
- Sidecar vs standalone deployment question

## Design

### Tooling and User Flows

With zone proxies becoming mesh-scoped, users need to specify which mesh their zone proxy should serve.
For the single-mesh default case, this is straightforward — one mesh per zone.

Deployment contexts:

1. **Konnect (MinK)** - Global CP is managed, UI has full mesh visibility
2. **Self-hosted Global CP** - Zone CP deployed via Helm, limited mesh visibility at install time
3. **Unfederated Zone** - Standalone zone, no global CP
4. **Terraform** - Infrastructure-as-code with dependency management

#### Flow 1: Konnect UI (Mesh in Konnect)

##### Current State

```
┌─────────────────────────────────────────────────────────┐
│ Connect zone                                            │
├─────────────────────────────────────────────────────────┤
│                                                         │
│ ☑ Deploy Ingress                                        │
│ ☑ Deploy Egress                                         │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

Generates `values.yaml` with `ingress.enabled: true` / `egress.enabled: true` (explicit opt-in)

##### Proposed Flow (Single-Mesh Default)

**Step 1: UI Enhancement**

For single-mesh deployments (the common case):

```
┌─────────────────────────────────────────────────────────┐
│ Connect zone                                            │
├─────────────────────────────────────────────────────────┤
│                                                         │
│ Mesh: [default ▼]                                       │
│ (if multiple meshes detected there will be info here on │
│ how to handle that)                                     │
└─────────────────────────────────────────────────────────┘
```

**Step 2: Generated values.yaml**

For the default mesh, the generated values.yaml enables the zone proxy explicitly:

```yaml
kuma:
  controlPlane:
    mode: zone
    zone: zone-1
    kdsGlobalAddress: grpcs://us.mesh.sync.konghq.tech:443

  zoneProxy:
    enabled: true
    # mesh defaults to "default"
```

If the user selected a non-default mesh name:

```yaml
kuma:
  controlPlane:
    mode: zone
    zone: zone-1
    kdsGlobalAddress: grpcs://us.mesh.sync.konghq.tech:443

  zoneProxy:
    enabled: true
    mesh: payments-mesh
```

Note: Zone proxy can be deployed before the mesh exists. It will wait and retry until the mesh is created on the Global CP.

**Why this works for Konnect:**

- Konnect UI has API access to global CP
- Can fetch mesh list via `GET /meshes`
- Validation happens UI-side before generating values.yaml
- If multiple meshes exist in the zone, the UI should inform the user that additional zone proxies must be configured manually (link to multi-mesh deployment docs)

#### Flow 2: Self-Hosted Global CP (Helm)

##### Single-Mesh (Default)

```yaml
kuma:
  controlPlane:
    mode: zone
    zone: zone-1
    kdsGlobalAddress: grpcs://global-cp:5685

  zoneProxy:
    enabled: true
    # mesh defaults to "default", explicit only if different
```

- Helm deploys zone proxy Deployment
- Zone proxy connects to CP, requests config for the configured mesh
- **If mesh doesn't exist**: CP returns error, zone proxy logs warning, retries
- **User experience**: Check zone proxy logs, see "mesh 'default' not found"

**Bootstrap validation**: Bootstrap already validates mesh existence for Dataplanes (returns HTTP 422 with `mesh: mesh "<name>" does not exist`).
With mesh-scoped zone proxy as Dataplane, this validation applies automatically.

Code references:
- Server validates mesh: [`pkg/xds/bootstrap/generator.go#L356-L366`](https://github.com/kumahq/kuma/blob/master/pkg/xds/bootstrap/generator.go#L356-L366)
- Server returns HTTP 422: [`pkg/xds/bootstrap/handler.go#L89-L96`](https://github.com/kumahq/kuma/blob/master/pkg/xds/bootstrap/handler.go#L89-L96)

This matches the "eventual consistency" model Kuma already uses.

#### Flow 3: Unfederated Zone (Standalone)

##### Single-Mesh (Default)

```yaml
kuma:
  controlPlane:
    mode: zone
    zone: zone-1
    # No kdsGlobalAddress = unfederated

  zoneProxy:
    enabled: true
    mesh: default
    replicas: 1
```

**Mesh creation in unfederated zones**: Unlike federated zones (where the mesh is synced from the Global CP), unfederated zones will create the mesh locally. There are two mechanisms:

1. **CP auto-creation**: When `skipMeshCreation: false` (the default), the CP creates the `default` mesh at startup (`EnsureDefaultMeshExists` in `pkg/defaults/mesh.go`). This only creates a mesh named `default` — non-default mesh names are not auto-created.

2. **Helm template**: For non-default mesh names, or when `skipMeshCreation: true`, Helm can create the Mesh resource:

```yaml
{{- if and .Values.zoneProxy.enabled (not .Values.controlPlane.kdsGlobalAddress) }}
  {{- /* Unfederated zone - create mesh locally */}}
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: {{ .Values.zoneProxy.mesh }}
---
{{- end }}
```

For the common case (`zoneProxy.mesh: default` + `skipMeshCreation: false`), both mechanisms produce the same result — the Helm template is a no-op since the mesh already exists.

#### Flow 4: Terraform

##### Key Advantage

Terraform has `depends_on` for ordering (mesh before zone proxy) and `templatefile()` for parameterizing mesh names. No custom provider resources are needed — zone proxies are deployed via the existing Helm chart using `helm_release`.

##### Single-Mesh (Default)

`main.tf`:

```hcl
resource "konnect_mesh" "payments" {
  name = "payments-mesh"
  # ...
}

resource "helm_release" "zone_proxy" {
  name       = "kuma-zone"
  repository = "https://kumahq.github.io/charts"
  chart      = "kuma"
  namespace  = "kuma-system"

  values = [
    templatefile("${path.module}/values.tftpl", {
      zone               = "zone-1"
      kds_global_address = "grpcs://us.mesh.sync.konghq.tech:443"
      mesh               = konnect_mesh.payments.name
    })
  ]

  depends_on = [konnect_mesh.payments]
}
```

`values.tftpl`:

```yaml
kuma:
  controlPlane:
    mode: zone
    zone: ${zone}
    kdsGlobalAddress: ${kds_global_address}

  zoneProxy:
    enabled: true
    mesh: ${mesh}
    replicas: 1
```

**Validation:** Same runtime model as Helm — zone proxy retries until the mesh exists on the Global CP. Terraform's `depends_on` ensures the mesh is created in Konnect/Global CP before the Helm release is applied, so in practice the mesh will already exist by the time the zone proxy starts.

##### Existing Mesh Variant

When the mesh already exists (e.g., `default` mesh), skip the `konnect_mesh` resource and reference the mesh by variable:

```hcl
variable "mesh_name" {
  default = "default"
}

resource "helm_release" "zone_proxy" {
  name       = "kuma-zone"
  repository = "https://kumahq.github.io/charts"
  chart      = "kuma"
  namespace  = "kuma-system"

  values = [
    templatefile("${path.module}/values.tftpl", {
      zone               = "zone-1"
      kds_global_address = "grpcs://us.mesh.sync.konghq.tech:443"
      mesh               = var.mesh_name
    })
  ]
}
```

#### Summary: Validation Strategies by Tool

| Tool | Can Validate Mesh? | Strategy |
|------|-------------------|----------|
| **Konnect UI** | Yes (API access) | Pre-populate mesh dropdown, validate before generating YAML |
| **Helm (federated)** | No (offline install) | Accept name, fail gracefully at runtime with clear logs |
| **Helm (unfederated)** | Yes (creates mesh) | Cross-reference in templates, fail at install if mismatch |
| **Terraform** | Via `depends_on` | `helm_release` with `templatefile()`; `depends_on` ensures mesh exists before zone proxy |

#### Design Decisions

##### 1. Mesh Deletion Handling

**Existing Protection**: Kuma already prevents mesh deletion when Dataplanes are attached.
The mesh validator returns an error: `"unable to delete mesh, there are still some dataplanes attached"`.
See [`pkg/core/managers/apis/mesh/mesh_validator.go#L70-L88`](https://github.com/kumahq/kuma/blob/master/pkg/core/managers/apis/mesh/mesh_validator.go#L70-L88).

**Mesh-Scoped Zone Proxy Benefit**: With zone proxies represented as Dataplane resources, this protection applies automatically.
No additional implementation is needed for mesh deletion handling - the existing safeguard covers the new deployment model.

**Note**: Current ZoneIngress/ZoneEgress resources are NOT covered by this protection (they're global-scoped).
The move to mesh-scoped Dataplanes resolves this gap.

For single-mesh, cleanup is straightforward: `helm uninstall <release-name>`.

##### 2. Per-Mesh Services (Not Shared)

**Decision**: Each mesh gets its own Service/LoadBalancer.

**Rationale**: With mesh-scoped zone proxies:
- Each mesh has **different mTLS CA certificates**
- Zone egress must verify the correct mesh's CA
- Sharing a LoadBalancer would require SNI-based cert selection (complex)

**Per-mesh services provide**:
- Proper mTLS isolation
- Simpler Envoy configuration
- Independent scaling and failover
- Clear network boundaries

**Service naming**: Follows the existing pattern from [`deployments/charts/kuma/templates/egress-service.yaml`](https://github.com/kumahq/kuma/blob/master/deployments/charts/kuma/templates/egress-service.yaml) but with per-mesh naming: `<release>-<mesh>-zoneproxy` (e.g., `kuma-payments-mesh-zoneproxy`).
This prevents name collisions when multiple meshes are deployed.

**Name length validation**: Helm templates should validate that the combined name fits within the 63-character Kubernetes limit:

```
{{ if gt (add (len .Values.zoneProxy.mesh) (len $prefix)) 63 }} {{ fail "zone proxy service name exceeds 63 characters; use a shorter mesh name" }} {{ end }}
```

This catches the issue at install time rather than silently truncating.

**Cost implication**: More LoadBalancers = higher cloud cost.
Users can use NodePort or Ingress controllers to reduce LB count if needed.

##### 3. Migration Path from Global Zone Proxies

**Phased migration**:

1. **Phase 1**: Deploy mesh-scoped zone proxies alongside global ones
2. **Phase 2**: Update MeshIdentity/policies to use mesh-scoped proxies
3. **Phase 3**: Drain traffic from global zone proxies
4. **Phase 4**: Remove global zone proxy deployment

**Helm migration (single-mesh)**:
```yaml
# Before (global)
ingress:
  enabled: true
egress:
  enabled: true

# After (mesh-scoped)
ingress:
  enabled: false  # Disable global
egress:
  enabled: false

zoneProxy:
  enabled: true
  mesh: default
```

#### Helm Release Structure Options

##### Option 1: Single Release (CP + Zone Proxy) — Recommended for Single-Mesh

```bash
helm install kuma kuma/kuma -f values.yaml
```

```yaml
# values.yaml
controlPlane:
  mode: zone
  zone: zone-1

zoneProxy:
  enabled: true
  # mesh defaults to "default"
```

| Aspect | Analysis |
|--------|----------|
| **Simplicity** | ✅ One command to install everything |
| **Upgrades** | ✅ Single `helm upgrade` updates all components |
| **Lifecycle coupling** | ✅ Single-mesh = single lifecycle, no issue |
| **Failure blast radius** | ⚠️ Bad values.yaml can break entire zone |
| **GitOps** | ✅ Single source of truth for zone configuration |

**Best for**: Single-mesh deployments (the default), teams preferring simplicity, GitOps workflows.

##### Option 2: Two Releases (CP separate from Zone Proxy)

```bash
helm install kuma-cp kuma/kuma -f cp-values.yaml
helm install kuma-zone-proxy kuma/kuma-zone-proxy -f zone-proxy-values.yaml
```

```yaml
# cp-values.yaml
controlPlane:
  mode: zone
  zone: zone-1

# zone-proxy-values.yaml (new chart)
enabled: true
# mesh defaults to "default"
replicas: 1
```

| Aspect | Analysis |
|--------|----------|
| **Simplicity** | ⚠️ Two charts to manage |
| **Upgrades** | ✅ Can upgrade CP independently of zone proxy |
| **Lifecycle coupling** | ✅ Zone proxy changes don't affect CP stability |
| **Failure blast radius** | ✅ Bad zone proxy config doesn't break CP |
| **GitOps** | ✅ Clear separation of concerns |

**Best for**: Production environments wanting CP stability isolation.

##### Helm Chart Structure: Pod Spec Passthrough

**Problem**: Current Helm charts enumerate Kubernetes fields one by one — the ingress chart alone is 292 lines for ~15 fields. Users are blocked when they need an unsupported field (e.g., `shareProcessNamespace`, `initContainers`, `topologySpreadConstraints`), creating a cat-and-mouse problem where the chart never fully covers the Kubernetes API surface.

**Solution**: The zone proxy chart accepts raw `podSpec` and `containers` sections. Helm's `merge` overlays user values onto sensible defaults. Adding any PodSpec or container field requires zero template changes.

```yaml
# values.yaml — open-ended passthrough
zoneProxy:
  enabled: true
  mesh: default
  podSpec: {}        # ANY valid PodSpec field (nodeSelector, tolerations, initContainers, etc.)
  containers: {}     # ANY container field (resources, lifecycle, securityContext, env, etc.)
```

This reduces template code from 292 lines to 71 lines while providing unlimited field coverage. See the [POC repo](https://github.com/slonka/poc-094-helm-zone-proxy) for a working demonstration.

#### Namespace Placement Options

**K8s Naming Constraint**: Kubernetes resource names (Deployments, Services) are limited to 63 characters.
With the naming pattern `zone-proxy-<mesh-name>`, mesh names should be kept under ~50 characters to avoid truncation.
Helm templates truncate at 63 chars (see [`deployments/charts/kuma/templates/_helpers.tpl#L31-L46`](https://github.com/kumahq/kuma/blob/master/deployments/charts/kuma/templates/_helpers.tpl#L31-L46)) which may cause unexpected name collisions with very long mesh names.

##### Option A: kuma-system Namespace (Centralized) — Recommended for Single-Mesh

```yaml
# Zone proxy in kuma-system
apiVersion: apps/v1
kind: Deployment
metadata:
  name: zone-proxy-default
  namespace: kuma-system
```

| Aspect | Analysis |
|--------|----------|
| **Operations** | ✅ All Kuma components in one place |
| **RBAC** | ✅ Simple - one namespace to grant access |
| **Monitoring** | ✅ Single namespace to scrape metrics |
| **Isolation** | ✅ Single-mesh = no isolation concern |
| **Resource quotas** | ✅ Single-mesh = simple quota management |

**Best for**: Single-mesh deployments (the default), simpler environments.

##### Recommendation

Deploy zone proxies in `kuma-system` namespace.

### Zone Proxy Deployment Topology

Whether zone ingress and egress share a single deployment is addressed in a separate MADR. [^2]

### Question 1: Support kuma.io/ingress-public-address

#### Current Implementation

From `pkg/plugins/runtime/k8s/controllers/ingress_converter.go:24-71`, the address resolution follows this priority:

1. Pod annotations: `kuma.io/ingress-public-address` + `kuma.io/ingress-public-port`
2. Service LoadBalancer IP/Hostname
3. NodePort with node address (ExternalIP > InternalIP)

#### Options

| Option | Description |
|--------|-------------|
| **A. Keep annotation** | Continue supporting the annotation as an override mechanism |
| **B. Service-only** | Remove annotation, rely solely on Service configuration |

#### Analysis

Use cases for annotation override:
- NAT gateways where Service IP differs from externally accessible address
- Split DNS environments
- Cloud provider quirks where LoadBalancer metadata is incorrect
- On-premises environments with external load balancers

**Option A: Keep annotation (recommended)**
- Advantages:
  - Supports edge cases where Service address doesn't reflect reality
  - Low maintenance burden
  - Backward compatible
- Disadvantages:
  - Additional configuration option to document

**Option B: Service-only**
- Advantages:
  - Simpler model
  - Encourages proper Service configuration
- Disadvantages:
  - Breaks legitimate use cases
  - Forces workarounds in complex network topologies

#### Recommendation

**Option A: Keep annotation support** but document it as an escape hatch:
- Primary method: Configure Service (LoadBalancer/NodePort) correctly
- Annotation: Use only when Service address is not accessible from other zones
- Consider adding a deprecation warning in logs when annotation is used

### Question 2: Default Helm Installation Behavior

#### Recommendation

**`zoneProxy.enabled: false`** — zone proxy deployment requires explicit opt-in, consistent with the existing `ingress.enabled` / `egress.enabled` pattern:

```yaml
controlPlane:
  mode: zone
  zone: zone-1

zoneProxy:
  enabled: true        # Explicit opt-in (default: false)
  # mesh defaults to "default"
```

This avoids a broken state where a zone proxy Deployment targets a mesh that doesn't exist. The `skipMeshCreation` flag is orthogonal — it controls whether the CP auto-creates the `default` Mesh at startup, and is independent of zone proxy deployment.

## Decision

### Tooling Decisions

1. **Per-mesh Services**: Each mesh gets its own Service/LoadBalancer for proper mTLS isolation.
   Sharing a LoadBalancer would require SNI-based cert selection which adds complexity.

2. **Namespace placement**: Deploy in `kuma-system` namespace.

3. **Deployment mechanism**: **Helm-managed** (current pattern extended for mesh-scoped zone proxies).
   Helm templates directly render Deployment + Service YAML with per-mesh configuration.

### Design Questions

1. **Keep kuma.io/ingress-public-address**: Support the annotation as an escape hatch for complex network topologies, but document Service-based configuration as the primary method.

2. **Helm defaults**: `zoneProxy.enabled: false` — explicit opt-in, consistent with `ingress.enabled` / `egress.enabled`.
   The `skipMeshCreation` flag is orthogonal to zone proxy deployment.

Note: Zone proxy deployment topology (shared vs separate ingress/egress) is addressed in a separate MADR. [^2]

### Out of Scope (Deferred to Resource Model MADR)

The following topics are covered in a separate MADR:

- **Dataplane representation**: Fields, labels, `kuma.io/proxy-type` tag
- **Workload identity**: `kuma.io/workload` annotation and auto-generation pattern (`zone-proxy-<mesh>-<role>`)
- **Token model**: Zone tokens → DP tokens transition for authentication
- **Universal deployment**: VM/bare metal specifics with mesh-scoped Dataplane resources
- **Sidecar vs standalone**: Whether zone proxies should be sidecars to fake containers

## Notes

### Related Issues and MADRs

- [Kuma #15429](https://github.com/kumahq/kuma/issues/15429): Label-based MeshService matching (no inbound tags)
- [Kuma #15431](https://github.com/kumahq/kuma/issues/15431): Protocol stored in Inbound field, not tags
- [KM #9028](https://github.com/Kong/kong-mesh/issues/9028): Dataplane fields for zone proxies
- [KM #9029](https://github.com/Kong/kong-mesh/issues/9029): Policies on zone proxies
- [KM #9032](https://github.com/Kong/kong-mesh/issues/9032): MeshIdentity for zone egress
- MADR 090: Zone Egress Identity

### Key Files Reference

| Component | File Path |
|-----------|-----------|
| ZoneIngress proto | `api/mesh/v1alpha1/zone_ingress.proto` |
| ZoneEgress proto | `api/mesh/v1alpha1/zoneegress.proto` |
| Dataplane proto | `api/mesh/v1alpha1/dataplane.proto` |
| K8s Ingress converter | `pkg/plugins/runtime/k8s/controllers/ingress_converter.go` |
| K8s Egress converter | `pkg/plugins/runtime/k8s/controllers/egress_converter.go` |
| Helm ingress deployment | `deployments/charts/kuma/templates/ingress-deployment.yaml` |
| Helm egress deployment | `deployments/charts/kuma/templates/egress-deployment.yaml` |
| Helm values | `deployments/charts/kuma/values.yaml` |
| Annotations | `pkg/plugins/runtime/k8s/metadata/annotations.go` |

[^1]: Resource model MADR link will be backfilled when created.
[^2]: Zone proxy deployment topology MADR link will be backfilled when created.
