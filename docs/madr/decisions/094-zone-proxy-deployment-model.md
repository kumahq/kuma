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

**Key insight**: At its core, a zone proxy is simply an Envoy instance.
Whether it functions as an ingress, egress, or both is determined by the **listeners the control plane generates** based on labels—not by fundamentally different proxy types.
Ingress and egress are listener configurations pushed via XDS.

This document addresses the following questions:

1. Should zone proxies be deployed as sidecars to "fake" containers on Kubernetes?
2. How should zone proxies be deployed on Universal (VM/bare metal)?
3. Should `kuma.io/workload` annotation be required on zone proxies?
4. Should we continue supporting `kuma.io/ingress-public-address` annotation?
5. What should be the default Helm installation behavior for zone proxies?

Note: Whether zone ingress and egress should be unified into a single zone proxy deployment is addressed in a separate MADR (MADR XXX).
All sections in this document apply to both ingress and egress zone proxies, except Question 4 (ingress only) and Design Decision 2: Per-Mesh Services (mostly ingress).

### Decision Summary

| Tooling Decision | Choice |
|------------------|--------|
| Per-mesh Services | **Yes** - each mesh gets its own Service/LoadBalancer for mTLS isolation |
| Namespace placement | **kuma-system** default, configurable per-mesh |

| Question | Decision                                         |
|----------|--------------------------------------------------|
| 1. Sidecar vs Standalone deployment? | no verdict yet                                   |
| 2. Universal deployment model? | Mesh-scoped Dataplane resources                  |
| 3. Require kuma.io/workload? | **Auto-generated** as `zone-proxy-<mesh>-<role>` |
| 4. Support ingress-public-address? | **Yes** - keep as escape hatch                   |
| 5. Default Helm behavior? | Zone proxy with `role: all`                      |

Note: Unified vs Separate zone proxies is addressed in a separate MADR (MADR XXX).

### Document Structure

This document is organized in two parts:

1. **Tooling and User Flows** - Describes how users will deploy zone proxies using different tools (Konnect UI, Helm, Terraform). This covers the UX and configuration experience.

2. **Questions 1-6** - Answers the design questions from the [technical story](https://github.com/kumahq/kuma/issues/9030). Each question analyzes options and recommends a decision.

## Design

### Tooling and User Flows

With zone proxies becoming mesh-scoped, users need to specify which mesh(es) their zone proxies should serve.
This creates different UX challenges depending on the deployment context:

1. **Konnect (MinK)** - Global CP is managed, UI has full mesh visibility
2. **Self-hosted Global CP** - Zone CP deployed via Helm, limited mesh visibility at install time
3. **Unfederated Zone** - Standalone zone, no global CP
4. **Terraform** - Infrastructure-as-code with dependency management

#### Key Insight

The core challenge is: **How does the deployment tool know which meshes exist?**

- Konnect UI: Has API access to global CP → knows all meshes
- Helm: Runs at install time → no API access to control plane
- Terraform: Can query resources → can enforce dependencies

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

Generates `values.yaml` with `ingress.enabled: true` / `egress.enabled: true`

##### Proposed Flow

**Step 1: UI Enhancement**

Replace simple checkboxes with a mesh-aware configuration:

```
┌─────────────────────────────────────────────────────────┐
│ Connect zone                                            │
├─────────────────────────────────────────────────────────┤
│                                                         │
│ Deployment mode:                                        │
│   ○ Unified (recommended) - single proxy, all roles     │
│   ○ Separate - independent ingress/egress proxies       │
│                                                         │
│ Meshes to serve:                                        │
│   ☑ payments-mesh                                       │
│   ☑ orders-mesh                                         │
│   ☐ staging-mesh                                        │
│   [+ Add all meshes]                                    │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

**Step 2: Generated values.yaml**

```yaml
kuma:
  controlPlane:
    mode: zone
    zone: zone-1
    kdsGlobalAddress: grpcs://us.mesh.sync.konghq.tech:443

  meshes:
    - name: payments-mesh
      role: all  # or: ingress, egress
      replicas: 2
    - name: orders-mesh
      role: all
      replicas: 1
```

Note: `meshes[]` only configures zone proxy deployment per mesh. Mesh creation (mTLS, backends, etc.) is managed separately on the Global CP.

Note: Zone proxy can be deployed before the mesh exists. It will wait and retry until the mesh is created on the Global CP.

**Why this works for Konnect:**

- Konnect UI has API access to global CP
- Can fetch mesh list via `GET /meshes`
- Pre-populates checkbox list with existing meshes
- Validation happens UI-side before generating values.yaml

#### Flow 2: Self-Hosted Global CP (Helm)

##### Challenge

- Helm runs at `helm install` time
- Zone CP hasn't connected to Global CP yet
- No way to query mesh list from Global CP

##### Options Considered

**Option A: Accept mesh names, fail at runtime (Recommended)**

```yaml
meshes:
  - name: payments-mesh
    role: all
```

- Helm deploys zone proxy Deployment
- Zone proxy connects to CP, requests config for `payments-mesh`
- **If mesh doesn't exist**: CP returns error, zone proxy logs warning, retries
- **User experience**: Check zone proxy logs, see "mesh 'payments-mesh' not found"

Pros:
- Simple Helm chart
- Works offline
- Clear error at runtime

Cons:
- Delayed feedback (not at install time)

**Option B: Helm pre-install hook to validate (Complex)**

```yaml
# pre-install-validate-meshes-job.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: validate-meshes
  annotations:
    "helm.sh/hook": pre-install
spec:
  template:
    spec:
      containers:
      - name: validate
        image: curlimages/curl
        command:
        - /bin/sh
        - -c
        - |
          for mesh in {{ .Values.meshes | join " " }}; do
            curl -f https://global-cp/meshes/$mesh || exit 1
          done
```

Pros:
- Fails fast at install time

Cons:
- Requires network access to Global CP from installer
- Auth complexity

##### Recommendation for Self-Hosted

**Option A** with good error messaging:

1. User specifies mesh names in values.yaml
2. Helm deploys zone proxy
3. Zone proxy logs clear message if mesh doesn't exist:
   ```
   WARN: Mesh 'payments-mesh' not found. Zone proxy waiting for mesh creation.
         Create the mesh on the Global CP or check the mesh name.
   ```
4. Once mesh exists, zone proxy auto-registers

**Bootstrap validation**: Bootstrap already validates mesh existence for Dataplanes (returns HTTP 422 with `mesh: mesh "<name>" does not exist`).
With mesh-scoped zone proxy as Dataplane, this validation applies automatically.
Current ZoneIngress/ZoneEgress bootstrap paths don't have this validation.

Code references:
- Server validates mesh: [`pkg/xds/bootstrap/generator.go#L356-L366`](https://github.com/kumahq/kuma/blob/master/pkg/xds/bootstrap/generator.go#L356-L366)
- Server returns HTTP 422: [`pkg/xds/bootstrap/handler.go#L89-L96`](https://github.com/kumahq/kuma/blob/master/pkg/xds/bootstrap/handler.go#L89-L96)
- Client handles 4xx error: [`app/kuma-dp/pkg/dataplane/envoy/remote_bootstrap.go#L258-L259`](https://github.com/kumahq/kuma/blob/master/app/kuma-dp/pkg/dataplane/envoy/remote_bootstrap.go#L258-L259)
- Client prints error: [`app/kuma-dp/cmd/run.go#L263`](https://github.com/kumahq/kuma/blob/master/app/kuma-dp/cmd/run.go#L263)

This matches the "eventual consistency" model Kuma already uses.

#### Flow 3: Unfederated Zone (Standalone)

##### Context

- No `kdsGlobalAddress` configured
- Zone CP manages meshes locally
- Zone proxy is "local" to this zone

##### Proposed Flow

```yaml
kuma:
  controlPlane:
    mode: zone
    zone: zone-1
    # No kdsGlobalAddress = unfederated

  meshes:
    - name: default
      role: all
      replicas: 1
```

For unfederated zones, Helm can also create the Mesh resources:

```yaml
{{- if not .Values.controlPlane.kdsGlobalAddress }}
  {{- /* Unfederated zone - create meshes locally */}}
  {{- range .Values.meshes }}
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: {{ .name }}
---
  {{- end }}
{{- end }}
```

**Validation:**

- If unfederated AND mesh listed in `meshes[]` → Helm creates the Mesh resource
- Zone proxy deployment references the same mesh name

#### Flow 4: Terraform

##### Key Advantage

Terraform has `depends_on` for ordering (mesh before zone proxy) and `templatefile()` for parameterizing mesh names. No custom provider resources are needed — zone proxies are deployed via the existing Helm chart using `helm_release`.

##### Proposed Structure

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
      meshes = [
        { name = konnect_mesh.payments.name, role = "all", replicas = 2 },
      ]
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

  meshes:
%{ for m in meshes ~}
    - name: ${m.name}
      role: ${m.role}
      replicas: ${m.replicas}
%{ endfor ~}
```

**Validation:** Same runtime model as Helm — zone proxy retries until the mesh exists on the Global CP. Terraform's `depends_on` ensures the mesh is created in Konnect/Global CP before the Helm release is applied, so in practice the mesh will already exist by the time the zone proxy starts.

#### Summary: Validation Strategies by Tool

| Tool | Can Validate Mesh? | Strategy |
|------|-------------------|----------|
| **Konnect UI** | Yes (API access) | Pre-populate mesh list, validate before generating YAML |
| **Helm (federated)** | No (offline install) | Accept names, fail gracefully at runtime with clear logs |
| **Helm (unfederated)** | Yes (creates mesh) | Cross-reference in templates, fail at install if mismatch |
| **Terraform** | Via `depends_on` | `helm_release` with `templatefile()`; `depends_on` ensures mesh exists before zone proxy |

#### Design Decisions

##### 1. Mesh Deletion Handling

**Existing Protection**: Kuma already prevents mesh deletion when Dataplanes are attached.
The mesh validator returns an error: `"unable to delete mesh, there are still some dataplanes attached"`.
See [`pkg/core/managers/apis/mesh/mesh_validator.go#L70-L88`](https://github.com/kumahq/kuma/blob/master/pkg/core/managers/apis/mesh/mesh_validator.go#L70-L88).

**Unified Zone Proxy Benefit**: With zone proxies represented as Dataplane resources, this protection applies automatically.
No additional implementation is needed for mesh deletion handling - the existing safeguard covers the new deployment model.

**Note**: Current ZoneIngress/ZoneEgress resources are NOT covered by this protection (they're global-scoped).
The move to mesh-scoped Dataplanes resolves this gap.

**K8s Deployment cleanup**: When a mesh is deleted (after removing its Dataplanes), the K8s Deployment for zone proxies remains.
For Helm, use one release per mesh for clean lifecycle management:
```bash
helm install zone-proxy-payments kuma/kuma-zone-proxy -f payments-values.yaml
helm install zone-proxy-orders kuma/kuma-zone-proxy -f orders-values.yaml
```

Then cleanup is: `helm uninstall zone-proxy-payments`

##### 2. Per-Mesh Services (Not Shared)

**Decision**: Each mesh gets its own Service/LoadBalancer.

**Rationale**: With mesh-scoped zone proxies:
- Each mesh has **different mTLS CA certificates**
- Zone ingress must present the correct mesh's certificate
- Zone egress must verify the correct mesh's CA
- Sharing a LoadBalancer would require SNI-based cert selection (complex)

**Per-mesh services provide**:
- Proper mTLS isolation
- Simpler Envoy configuration
- Independent scaling and failover
- Clear network boundaries

**Service naming**: Follows the existing pattern from [`deployments/charts/kuma/templates/egress-service.yaml`](https://github.com/kumahq/kuma/blob/master/deployments/charts/kuma/templates/egress-service.yaml) but with per-mesh naming: `<release>-<mesh>-zoneproxy` (e.g., `kuma-payments-mesh-zoneproxy`).
This prevents name collisions when multiple meshes are deployed.

**Cost implication**: More LoadBalancers = higher cloud cost.
Users can use NodePort or Ingress controllers to reduce LB count if needed.

##### 3. Migration Path from Global Zone Proxies

**Phased migration**:

1. **Phase 1**: Deploy mesh-scoped zone proxies alongside global ones
2. **Phase 2**: Update MeshIdentity/policies to use mesh-scoped proxies
3. **Phase 3**: Drain traffic from global zone proxies
4. **Phase 4**: Remove global zone proxy deployment

**Helm migration**:
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

meshes:
  - name: payments-mesh
    role: all
```

#### Helm Release Structure Options

##### Option 1: Single Release (CP + All Zone Proxies)

```bash
helm install kuma kuma/kuma -f values.yaml
```

```yaml
# values.yaml
controlPlane:
  mode: zone
  zone: zone-1

meshes:
  - name: payments-mesh
    role: all
  - name: orders-mesh
    role: all
```

| Aspect | Analysis |
|--------|----------|
| **Simplicity** | ✅ One command to install everything |
| **Upgrades** | ✅ Single `helm upgrade` updates all components |
| **Lifecycle coupling** | ⚠️ Adding/removing a mesh requires full release upgrade |
| **Failure blast radius** | ⚠️ Bad values.yaml can break entire zone |
| **Cleanup** | ⚠️ Can't remove one mesh's zone proxy without touching others |
| **GitOps** | ✅ Single source of truth for zone configuration |

**Best for**: Small deployments, teams preferring simplicity, GitOps workflows.

##### Option 2: Two Releases (CP separate from Zone Proxies)

```bash
helm install kuma-cp kuma/kuma -f cp-values.yaml
helm install kuma-zone-proxies kuma/kuma-zone-proxy -f zone-proxy-values.yaml
```

```yaml
# cp-values.yaml
controlPlane:
  mode: zone
  zone: zone-1

# zone-proxy-values.yaml (new chart)
meshes:
  - name: payments-mesh
  - name: orders-mesh
```

| Aspect | Analysis |
|--------|----------|
| **Simplicity** | ⚠️ Two charts to manage |
| **Upgrades** | ✅ Can upgrade CP independently of zone proxies |
| **Lifecycle coupling** | ✅ Zone proxy changes don't affect CP stability |
| **Failure blast radius** | ✅ Bad zone proxy config doesn't break CP |
| **Cleanup** | ⚠️ Still can't remove one mesh without touching others |
| **GitOps** | ✅ Clear separation of concerns |

**Best for**: Production environments, teams wanting CP stability isolation.

##### Option 3: N Releases (CP + One per Mesh)

```bash
helm install kuma-cp kuma/kuma -f cp-values.yaml
helm install zp-payments kuma/kuma-zone-proxy -f payments-values.yaml
helm install zp-orders kuma/kuma-zone-proxy -f orders-values.yaml
```

```yaml
# payments-values.yaml
meshes:
  - name: payments-mesh
    role: all
    replicas: 2
```

| Aspect | Analysis |
|--------|----------|
| **Simplicity** | ❌ Many releases to manage |
| **Upgrades** | ✅ Fine-grained control per mesh |
| **Lifecycle coupling** | ✅ Each mesh fully independent |
| **Failure blast radius** | ✅ Issues isolated to single mesh |
| **Cleanup** | ✅ `helm uninstall zp-payments` removes just that mesh |
| **GitOps** | ⚠️ Multiple files/releases to track |
| **Scaling** | ✅ Different resource profiles per mesh |

**Best for**: Large deployments, multi-team environments, meshes with different SLAs.

#### Namespace Placement Options

**K8s Naming Constraint**: Kubernetes resource names (Deployments, Services) are limited to 63 characters.
With the naming pattern `zone-proxy-<mesh-name>`, mesh names should be kept under ~50 characters to avoid truncation.
Helm templates truncate at 63 chars (see [`deployments/charts/kuma/templates/_helpers.tpl#L31-L46`](https://github.com/kumahq/kuma/blob/master/deployments/charts/kuma/templates/_helpers.tpl#L31-L46)) which may cause unexpected name collisions with very long mesh names.

##### Option A: kuma-system Namespace (Centralized)

```yaml
# All zone proxies in kuma-system
apiVersion: apps/v1
kind: Deployment
metadata:
  name: zone-proxy-payments-mesh
  namespace: kuma-system
```

| Aspect | Analysis |
|--------|----------|
| **Operations** | ✅ All Kuma components in one place |
| **RBAC** | ✅ Simple - one namespace to grant access |
| **Monitoring** | ✅ Single namespace to scrape metrics |
| **Isolation** | ❌ All meshes share failure domain |
| **Resource quotas** | ❌ Hard to enforce per-mesh limits |
| **Multi-tenancy** | ❌ Teams can't own their zone proxy |

**Best for**: Single-team operations, simpler environments.

##### Option B: Per-Mesh Namespace

```yaml
# Zone proxy in mesh's namespace
apiVersion: apps/v1
kind: Deployment
metadata:
  name: zone-proxy
  namespace: payments-mesh  # or payments-system
```

| Aspect | Analysis |
|--------|----------|
| **Operations** | ⚠️ Zone proxies distributed across namespaces |
| **RBAC** | ✅ Teams can own their mesh's infrastructure |
| **Monitoring** | ⚠️ Need to aggregate from multiple namespaces |
| **Isolation** | ✅ Mesh failure doesn't affect others |
| **Resource quotas** | ✅ Per-namespace quotas apply to zone proxy |
| **Multi-tenancy** | ✅ Clear ownership boundaries |

**Best for**: Multi-team, multi-tenant environments, strict isolation requirements.

##### Option C: Dedicated Zone Proxy Namespace

```yaml
# All zone proxies in dedicated namespace
apiVersion: apps/v1
kind: Deployment
metadata:
  name: zone-proxy-payments-mesh
  namespace: kuma-zone-proxies
```

| Aspect | Analysis |
|--------|----------|
| **Operations** | ✅ All zone proxies together, separate from CP |
| **RBAC** | ✅ Can grant zone proxy access without CP access |
| **Monitoring** | ✅ Single namespace for zone proxy metrics |
| **Isolation** | ⚠️ Zone proxies share namespace, but separate from CP |
| **Resource quotas** | ⚠️ Can limit total zone proxy resources |
| **Multi-tenancy** | ⚠️ Partial - zone proxies separate from apps |

**Best for**: Teams wanting separation from CP but not full per-mesh isolation.

##### Recommendation

**Default to kuma-system**, make namespace configurable:

```yaml
meshes:
  - name: payments-mesh
    namespace: kuma-system  # default, configurable via Helm values
  - name: orders-mesh
    namespace: orders-system  # override for this mesh
```

This allows gradual migration to per-mesh namespaces without breaking existing setups.

### Unified vs Separate Zone Proxies

Left to a separate MADR (MADR XXX).

### Question 1: Sidecar to Fake Container vs Standalone Deployment

#### Current Implementation

Zone proxies are deployed as standalone Kubernetes Deployments with dedicated pods containing only the kuma-dp container.
With a unified zone proxy model (see MADR XXX), this would consolidate to a single deployment template (or separate templates for operators choosing split deployments).

#### Options

| Option | Description |
|--------|-------------|
| **A. Sidecar to pause container** | Inject zone proxy as sidecar to a pod with a minimal "sleep infinity" or pause container |
| **B. Standalone deployment** | Current approach - dedicated deployment with only kuma-dp |
| **C. Direct pod without injection** | Create pods directly without using sidecar injection webhook |

#### Analysis

The primary concern raised was that "mesh updates won't be updated automatically" with standalone deployments.
This concern refers to **Kuma version upgrades** (e.g., `helm upgrade`):

- **Option A (sidecar to fake container)**: The sidecar injection webhook injects the proxy image at pod creation time.
  On `helm upgrade`, existing pods are **not automatically restarted** - the operator must manually trigger pod recreation to pick up the new version.
  This gives operators full control over upgrade timing, allowing careful rollouts that avoid dropping requests.

- **Option B (standalone deployment)**: The Deployment spec directly references the proxy image.
  On `helm upgrade`, the Deployment is updated, **triggering automatic pod rollout**.
  The rolling update strategy (`maxUnavailable`, `maxSurge`) controls how pods are replaced, preventing all-at-once restarts.

Option B is still recommended because zone proxies are infrastructure that should update with the control plane.
The rolling update strategy provides sufficient control over the upgrade process.

#### Advantages and Disadvantages

**Option A: Sidecar to fake container**
- Advantages:
  - Consistent with application sidecar pattern
  - Could leverage existing injection infrastructure
  - Manual upgrade control - operator decides when to roll pods
  - You don't need special `limit / resources` for egress / ingress you can just use containerpatches for regular sidecars.
- Disadvantages:
  - Adds unnecessary resource overhead (pause container)
  - More complex deployment model
  - Confusing operational model (what is the "application"?)

**Option B: Standalone deployment**
- Advantages:
  - Clear operational model - zone proxies are infrastructure
  - Independent scaling and lifecycle management
  - Familiar pattern for infrastructure operators
  - Lower resource overhead
  - Simpler debugging and monitoring
  - **Prometheus metrics support**: Unified zone proxy as Dataplane uses `DefaultProxyProfile` which includes `PrometheusEndpointGenerator` (see [`pkg/xds/generator/proxy_template.go#L78-L92`](https://github.com/kumahq/kuma/blob/master/pkg/xds/generator/proxy_template.go#L78-L92)).
    This enables Prometheus metrics scraping via the metrics hijacker.
    Current ZoneIngress/ZoneEgress profiles don't include this generator.
- Disadvantages:
  - Zone proxies are part of the helm chart so they restart at the same time as the CP which could cause newer proxies to connect to old CPs
  - Different pattern from application sidecars (minor)

#### Recommendation

No verdict yet.

### Question 2: Universal Deployment Model

#### Current Universal Deployment

On Universal (VM/bare metal), zone proxies are deployed by:
1. Creating `ZoneIngress` or `ZoneEgress` resources via `kumactl apply` or REST API
2. Running `kuma-dp run` with the appropriate flags

Resources are **global-scoped** (no mesh field), so a single deployment handles all meshes:

```yaml
type: ZoneEgress
name: egress-1
spec:
  zone: zone-1
  networking:
    address: 10.0.0.1
    port: 10002
```

#### New Mesh-Scoped Deployment

With the move to mesh-scoped zone proxies, the deployment changes:

1. Zone proxies become `Dataplane` resources with specific labels
2. Each resource must specify a `mesh` field
3. Deploy one zone proxy instance **per mesh** (not one per zone for all meshes)

```yaml
type: Dataplane
mesh: payments-mesh
name: zone-proxy-payments
labels:
  kuma.io/proxy-type: zoneproxy
networking:
  address: 10.0.0.1
  advertisedAddress: 203.0.113.1
  advertisedPort: 10001
```

Whether a zone proxy handles ingress, egress, or both (unified vs separate) is addressed in MADR XXX.

#### Migration Path

1. Deploy new mesh-scoped zone proxy for each mesh requiring cross-zone/external traffic
2. Control plane routes traffic to new proxy when MeshIdentity is enabled for that mesh
3. Scale down old global-scoped proxy after verification
4. Remove legacy ZoneIngress/ZoneEgress resources

#### Tooling Updates Required

- `kumactl` commands for zone proxy management
- Documentation for Universal deployment
- Migration guide from global to mesh-scoped

### Question 3: Require kuma.io/workload Annotation

#### Current Usage

From `pkg/plugins/runtime/k8s/metadata/annotations.go`:
- `kuma.io/workload`: "Specifies workload identifier associated with Pod"
- Used for correlating pods to parent workload objects (Deployments, StatefulSets, etc.)

#### Options

| Option | Description |
|--------|-------------|
| **A. Required** | Zone proxies must have `kuma.io/workload` annotation |
| **B. Not required** | Zone proxies don't need this annotation |

#### Analysis

**Option A: Required**
- Advantages:
  - Consistent with regular sidecars
  - Enables workload-based policy targeting
- Disadvantages:
  - Zone proxies don't have a "workload" in the application sense - they ARE the workload
  - Adds unnecessary configuration requirement
  - Conceptually confusing

**Option B: Not required (recommended)**
- Advantages:
  - Simpler configuration
  - Zone proxies are infrastructure, not application workloads
  - Targeted by the `kuma.io/proxy-type` label (e.g., `zoneegress`) or extended `proxyTypes` in targetRef
  - The Deployment itself is the workload - no parent correlation needed
- Disadvantages:
  - Slight inconsistency with application sidecar pattern (acceptable)

#### Recommendation

**Option B: Not required** - Zone proxies should not require manual `kuma.io/workload` configuration.
They should be identified by the `kuma.io/proxy-type: zoneproxy` label and the `kuma.io/zone-proxy-role` label for role-specific targeting.
Policies can use extended `proxyTypes` in targetRef or label-based selectors.
The auto-generated `kuma.io/workload` value serves as the security binding for DP tokens, replacing the coarse zone token scoping (see Token Implications below).

**Auto-generated workload identity**: On Universal, `kuma.io/workload` is used to create identity for a Dataplane (required for token generation).
For zone proxies, the workload will be auto-generated with the pattern: `zone-proxy-<mesh>-<role>`.
Example: `zone-proxy-payments-mesh-all`.
This enables token generation on Universal without manual configuration.

#### Token Implications

When zone proxies become mesh-scoped Dataplane resources, their authentication path changes from zone tokens to DP tokens. This has significant security implications:

1. **Current state**: Zone proxies use zone tokens with `Zone` + `Scope` claims ([`ZoneClaims`](https://github.com/kumahq/kuma/blob/master/pkg/tokens/builtin/zone/token.go)). The auth path is [`authZoneEntity()`](https://github.com/kumahq/kuma/blob/master/pkg/xds/auth/universal/authenticator.go#L76-L99) which checks [scope](https://github.com/kumahq/kuma/blob/master/pkg/tokens/builtin/zone/scope.go) (ingress/egress) and zone name. There is no mesh or workload binding — the token is coarse-grained.

   ```json
   {
     "Zone": "us-east",
     "Scope": ["ingress"],
     "exp": 1735689600,
     "jti": "a]b2c3d4-e5f6-7890-abcd-ef1234567890"
   }
   ```

2. **New state**: Zone proxies as Dataplanes will use DP tokens ([`DataplaneClaims`](https://github.com/kumahq/kuma/blob/master/pkg/tokens/builtin/issuer/token.go) with fields: `Name`, `Mesh`, `Tags`, `Type`, `Workload`). The auth path becomes [`authDataplane()`](https://github.com/kumahq/kuma/blob/master/pkg/xds/auth/universal/authenticator.go#L55-L73) which validates `Name`, `Mesh`, `Tags`, and `Workload`. This is strictly more secure — tokens are bound to a specific mesh and can be bound to a specific workload identity.

   ```json
   {
     "Name": "zone-proxy-payments-mesh-all-1",
     "Mesh": "payments-mesh",
     "Tags": {
       "kuma.io/proxy-type": ["zoneproxy"],
       "kuma.io/zone-proxy-role": ["all"]
     },
     "Type": "dataplane",
     "Workload": "zone-proxy-payments-mesh-all",
     "exp": 1735689600,
     "jti": "f6e5d4c3-b2a1-0987-fedc-ba0987654321"
   }
   ```

3. **How `kuma.io/workload` fits**: With the auto-generated workload `zone-proxy-<mesh>-<role>`, DP tokens can include this workload value. [`validateWorkload()`](https://github.com/kumahq/kuma/blob/master/pkg/xds/auth/universal/authenticator.go#L116-L128) enforces the match. This means:
   - A token generated for `workload: zone-proxy-payments-mesh-all` can ONLY be used by a zone proxy for `payments-mesh` with role `all`
   - This replaces the coarse `Scope: ["ingress"]` with fine-grained workload binding

4. **Zone token scope → workload mapping**: The old `Scope` concept (ingress/egress) is subsumed by the workload pattern:
   - `Scope: ["ingress"]` → `Workload: "zone-proxy-<mesh>-ingress"`
   - `Scope: ["egress"]` → `Workload: "zone-proxy-<mesh>-egress"`
   - `Scope: ["ingress", "egress"]` → `Workload: "zone-proxy-<mesh>-all"`

5. **Migration**: Existing zone tokens continue to work for legacy `ZoneIngress`/`ZoneEgress` resources (routed via the [`Authenticate()` switch](https://github.com/kumahq/kuma/blob/master/pkg/xds/auth/universal/authenticator.go#L42-L53)). New mesh-scoped zone proxies use DP tokens. The `authDataplane()` path is already production-proven for regular dataplanes.

Since zone proxies become mesh-scoped Dataplane resources, targeting a zone proxy for a specific mesh is straightforward: policies themselves are mesh-scoped, so a policy in `payments-mesh` automatically only applies to zone proxies in that mesh.
The `kuma.io/zone-proxy-role` label then filters to specific roles (ingress, egress, or all) within that mesh.

The auto-generated workload identity is not just a convenience for token generation — it is the security mechanism that replaces zone token scoping. By binding DP tokens to a specific workload value, we get mesh-scoped, role-specific authentication that is strictly more granular than the current zone token model.

### Question 4: Support kuma.io/ingress-public-address

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

### Question 5: Default Helm Installation Behavior

#### Recommendation

**Multi-zone ready with unified zone proxy** - Deploy a single zone proxy (handling all traffic, i.e., both ingress and egress) and skip default mesh creation:

```yaml
controlPlane:
  defaults:
    skipMeshCreation: true  # No default mesh

meshes:
  - name: my-mesh
    role: all  # or: ingress, egress
    replicas: 1
```

The `role` setting controls deployment topology:
- `all` (default): Single Deployment handling both ingress and egress
- `ingress`: Single Deployment handling only ingress
- `egress`: Single Deployment handling only egress
- `separate`: Two Deployments (one ingress, one egress) for operators needing independent scaling

This provides multi-zone readiness while requiring explicit mesh creation.
The unified default reduces resource usage and operational complexity for typical deployments.

## Decision

### Tooling Decisions

1. **Per-mesh Services**: Each mesh gets its own Service/LoadBalancer for proper mTLS isolation.
   Sharing a LoadBalancer would require SNI-based cert selection which adds complexity.

2. **Namespace placement**: Default to `kuma-system`, make namespace configurable per-mesh.
   This allows gradual migration to per-mesh namespaces without breaking existing setups.

### Design Questions

1. **Standalone deployment**: no verdict yet.

2. **Universal deployment**: Use mesh-scoped Dataplane resources with zone proxy labels instead of global ZoneIngress/ZoneEgress.
   Deploy one zone proxy per mesh.

3. **kuma.io/workload auto-generated**: Zone proxies will have `kuma.io/workload` auto-generated with the pattern `zone-proxy-<mesh>-<role>` (e.g., `zone-proxy-payments-mesh-all`).
   This enables token generation on Universal without manual configuration.
   Zone proxies should be targeted by the `kuma.io/proxy-type: zoneproxy` label and `kuma.io/zone-proxy-role` for role-specific targeting.

4. **Keep kuma.io/ingress-public-address**: Support the annotation as an escape hatch for complex network topologies, but document Service-based configuration as the primary method.

5. **Helm defaults**: Zone proxy deployment with `role: all`, with option to split via `role: separate`.
   No default mesh created (`skipMeshCreation: true`).

Note: Unified vs Separate zone proxies is addressed in a separate MADR (MADR XXX).

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
