# Zone Proxy Deployment Model

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/9030

## Context and Problem Statement

Currently, zone proxies are **global-scoped** resources (global in Kuma's naming convention, meaning cluster-scoped in Kubernetes terms — not global across multiple clusters), meaning a single ZoneIngress or ZoneEgress instance handles traffic for all meshes in a zone.
This global nature creates fundamental limitations:

1. **Cannot issue MeshIdentity for zone egress**: MeshIdentity is mesh-scoped, but a global zone egress serves multiple meshes.
   This creates identity conflicts and prevents proper mTLS certificate issuance.
   See [MADR 090](090-zone-egress-identity.md) for detailed analysis.

2. **Cannot apply policies on zone proxies**: Kuma policies (MeshTrafficPermission, MeshTimeout, etc.) are mesh-scoped.
   A global zone proxy cannot be targeted by mesh-specific policies, limiting observability and traffic control for cross-zone communication.

3. **Limited observability scoping**: With a global zone proxy, metrics, access logs, and traces cannot be scoped to a specific mesh — all mesh traffic is mixed. Mesh-scoped zone proxies enable per-mesh observability via policies like MeshAccessLog and MeshMetric.

To resolve these limitations, zone proxies are being changed to **mesh-scoped** resources represented as Dataplane resources with specific tags.
This architectural change requires revisiting the deployment model for zone proxies.

**Scope of this document**: This MADR focuses on **deployment tooling** — how users deploy zone proxies via Helm, Konnect UI, and Terraform.

**Multi-mesh support**: The `meshes` list supports deploying zone proxies for multiple meshes in a single Helm release. Single-mesh is the simplest case — one entry in the list.

This document addresses the following question:

1. What should be the default Helm installation behavior for zone proxies?

Note: Whether zone ingress and egress share a single deployment is addressed in a separate MADR. [^2]

### Decision Summary

| Tooling Decision | Choice |
|------------------|--------|
| Per-mesh Services | **Yes** - each mesh gets its own Service/LoadBalancer for mTLS isolation |
| Namespace placement | **kuma-system** |
| Deployment mechanism | **Helm-managed** (current pattern extended for mesh-scoped zone proxies) |
| Helm release structure | **Per-mesh templates** (each mesh entry rendered by per-mesh templates) |
| MADR 093 revert | **Yes** - allow multiple meshes per namespace, handle Workload collisions in controller |
| Additive migration | **Yes** - `meshes` config alongside existing `ingress`/`egress` keys |

| Question                  | Decision                                                            |
|:--------------------------|:--------------------------------------------------------------------|
| 1. Default Helm behavior? | `meshes: []` — explicit opt-in, no zone proxies deployed by default |

Note: Resource model (Dataplane representation, labels, tokens, workload identity) is in a separate MADR. [^1]
Note: Zone proxy deployment topology (shared vs separate ingress/egress) is addressed in a separate MADR. [^2]

### Document Structure

This document is organized in two parts:

1. **Tooling and User Flows** - Describes how users will deploy zone proxies using different tools (Konnect UI, Helm, Terraform). This covers the UX and configuration experience.

2. **Question 1** - Answers a deployment-related design question with options analysis and recommendation.

## Design

### Tooling and User Flows

With zone proxies becoming mesh-scoped, users configure zone proxies per mesh via the `meshes` list in `values.yaml`. Each entry in `meshes` is rendered by per-mesh templates that manage the zone proxy Deployment(s), Service(s), HPA, PDB, ServiceAccount, and optionally the Mesh resource and default policies.

For single-mesh deployments, the `meshes` list has one entry. Multi-mesh deployments add additional entries.

**Full mesh entry schema**:

```yaml
meshes:
  - name: <mesh-name>            # Required. Name of the mesh this entry targets.
    createMesh: false             # Optional. Render a Mesh resource. Default: false. Ignored when kdsGlobalAddress is set.
    createPolicies:               # Optional. List of default policies to create with the mesh. Default: []. Ignored when kdsGlobalAddress is set.
      - MeshCircuitBreaker        # Default circuit breaker for all traffic.
      - MeshRetry                 # Default retry config (TCP + HTTP/GRPC).
      - MeshTimeout               # Default timeouts for sidecars + gateways.

    # Zone proxy deployment — choose EITHER ingress/egress OR combinedProxies (not both).

    ingress:                      # Separate ingress deployment.
      enabled: true
      podSpec: {}
      hpa: {}
      pdb: {}
      service:
        name: ""                  # Optional override for auto-generated Service name.
      resources: {}

    egress:                       # Separate egress deployment.
      enabled: true
      podSpec: {}
      hpa: {}
      pdb: {}
      service:
        name: ""                  # Optional override for auto-generated Service name.
      resources: {}

    combinedProxies:              # Single deployment running both ingress + egress roles.
      enabled: true               # Mutually exclusive with ingress/egress above.
      podSpec: {}
      hpa: {}
      pdb: {}
      service:
        name: ""                  # Optional override for auto-generated Service name.
      resources: {}
```

**Mesh validation behavior**: Zone proxies can be deployed before their target mesh exists. Bootstrap already validates mesh existence for Dataplanes — the CP returns HTTP 422 with `mesh: mesh "<name>" does not exist`. With mesh-scoped zone proxies represented as Dataplanes, this validation applies automatically. The zone proxy logs the error and retries until the mesh is created. This "eventual consistency" model means users don't need to sequence mesh creation before zone proxy deployment, though tools like Terraform can enforce ordering via `depends_on`.

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

##### Proposed Flow

**Step 1: UI Enhancement**

```
┌──────────────────────────────────────────────────────────────────────────┐
│ Connect zone                                                             │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│ Detected meshes:                                                         │
│                                                                          │
│ ☑ default     Mode: [Separate ▼]  Ingress: [Yes ▼]  Egress: [Yes ▼]     │
│ ☑ payments    Mode: [Combined ▼]  Ingress: [--- ▼]  Egress: [--- ▼]     │
│ ☐ backend     Mode: [--------- ▼]  Ingress: [--- ▼]  Egress: [--- ▼]    │
│                                                                          │
│ Mode options:                                                            │
│   Separate — deploy ingress and egress as independent Deployments        │
│   Combined — deploy a single Deployment running both roles               │
│                                                                          │
└──────────────────────────────────────────────────────────────────────────┘
```

All detected meshes are auto-filled as entries. Each mesh has its own mode selector and ingress/egress toggles, mirroring the per-mesh structure in the `meshes` schema. When "Combined" mode is selected, the Ingress/Egress toggles are disabled (the combined Deployment handles both roles). Users can deselect meshes they don't need zone proxies for — deselected meshes are excluded from the generated `values.yaml`.

**Mesh-first requirement**: The Konnect UI requires at least one mesh to exist before generating zone proxy values.

**Step 2: Generated values.yaml**

The generated values.yaml uses the `meshes` list:

```yaml
kuma:
  controlPlane:
    mode: zone
    zone: zone-1
    kdsGlobalAddress: grpcs://us.mesh.sync.konghq.tech:443

  meshes:
    - name: default
      ingress:
        enabled: true
      egress:
        enabled: true
```

**Why this works for Konnect:**

- Konnect UI has API access to global CP
- Can fetch mesh list via `GET /meshes`
- Validation happens UI-side before generating values.yaml

#### Flow 2: Self-Hosted Global CP (Helm)

##### Single-Mesh (Default)

```yaml
kuma:
  controlPlane:
    mode: zone
    zone: zone-1
    kdsGlobalAddress: grpcs://global-cp:5685

  meshes:
    - name: default
      ingress:
        enabled: true
      egress:
        enabled: true
```

- Helm renders per-mesh templates for each entry in `meshes`
- Zone proxy connects to CP, requests config for the configured mesh
- If the mesh doesn't exist yet, the zone proxy retries (see Mesh validation behavior above)

##### Multi-Mesh Variant

For multi-mesh, add additional entries to the `meshes` list — each renders independently.

#### Flow 3: Unfederated Zone (Standalone)

##### Single-Mesh (Default)

```yaml
kuma:
  controlPlane:
    mode: zone
    zone: zone-1
    # No kdsGlobalAddress = unfederated

  meshes:
    - name: default
      createMesh: true
      ingress:
        enabled: true
      egress:
        enabled: true
```

**Mesh creation in unfederated zones**: Unlike federated zones (where the mesh is synced from the Global CP), unfederated zones create the mesh locally. The `createMesh` field controls this:

1. **`createMesh: true`**: The chart renders a Mesh resource. This is needed for unfederated zones where no Global CP syncs meshes.

2. **CP auto-creation**: When `skipMeshCreation: false` (the default), the CP creates the `default` mesh at startup (`EnsureDefaultMeshExists` in `pkg/defaults/mesh.go`). This only creates a mesh named `default` — non-default mesh names require `createMesh: true`.

3. **Conditional rendering**: The chart only renders the Mesh resource and default policies when no `kdsGlobalAddress` is configured (unfederated). For federated zones, both `createMesh` and `createPolicies` are ignored since meshes and policies are synced from the Global CP.

For the common case (`name: default` + `skipMeshCreation: false`), the user can omit `createMesh` — the CP handles default mesh creation automatically.

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

  meshes:
    - name: ${mesh}
      ingress:
        enabled: true
      egress:
        enabled: true
```

##### Existing Mesh Variant

For existing meshes, omit the `konnect_mesh` resource and pass the mesh name as a variable.

#### Design Decisions

##### 1. Helm Release Structure: Per-Mesh Templates

Each entry in the `meshes` list is rendered by **per-mesh templates**. The templates encapsulate all resources for a single mesh's zone proxy deployment.

**Per-mesh template responsibilities** (per mesh entry):

- Deployment(s) — zone ingress, zone egress, or combined
- Service(s) — per-mesh LoadBalancer/NodePort
- HPA — horizontal pod autoscaler
- PDB — pod disruption budget
- ServiceAccount
- Mesh resource (when `createMesh: true`)
- Default policies (controlled by `createPolicies` list)

| Aspect         | Analysis                                                                  |
|:---------------|:--------------------------------------------------------------------------|
| **Simplicity** | One `helm install` deploys everything; `meshes` list is declarative       |
| **Isolation**  | Each mesh entry renders independent resources; no cross-mesh interference |
| **GitOps**     | Single `values.yaml` is the source of truth for all meshes in a zone      |

**`createPolicies`** is a list of default policies to create alongside the mesh (see schema above for available policies).

**`combinedProxies`** merges ingress and egress into a single Deployment. Mutually exclusive with `ingress`/`egress` — Helm validates and errors if both are defined. References the separate deployment topology MADR [^2] for analysis.

**Implementation**: The per-mesh templates can be organized as a Helm library chart, a subchart, or named templates within the main chart — this is left to the implementation. Users interact purely through the `meshes` list in `values.yaml`.

From the user's perspective, the install experience is identical to today — only the `values.yaml` content changes:

```bash
helm install kuma kumahq/kuma -n kuma-system -f values.yaml
```

```yaml
# values.yaml
meshes:
  - name: default
    ingress:
      enabled: true
    egress:
      enabled: true
```

To add a mesh later, update `values.yaml` and upgrade:

```bash
helm upgrade kuma kumahq/kuma -n kuma-system -f values.yaml
```

###### Considered Alternatives

1. **Single release with flat `zoneProxy` config** — the previous recommended approach: flat `zoneProxy.enabled`/`zoneProxy.mesh` in the main chart, no per-mesh templates. Rejected because it doesn't naturally extend to multi-mesh and doesn't encapsulate mesh lifecycle (mesh creation, default policies).

2. **Separate zone-proxy chart** — a standalone `kuma-zone-proxy` chart for zone proxy releases. Rejected because it's too narrow (only handles zone proxies, not mesh lifecycle) and adds maintenance burden for a partial solution.

3. **Multi-mesh out of scope** — previous decision to leave multi-mesh to users. Reversed because the per-mesh template approach naturally supports it without additional complexity.

##### 2. Per-Mesh Services (Not Shared)

**Decision**: Each mesh gets its own Service/LoadBalancer.

**Rationale**: With mesh-scoped zone proxies:
- Each mesh has **different mTLS CA certificates** — sharing a LoadBalancer would require SNI-based cert selection (complex)
- Independent scaling and failover per mesh

**Service naming**: Follows the existing pattern from [`deployments/charts/kuma/templates/egress-service.yaml`](https://github.com/kumahq/kuma/blob/master/deployments/charts/kuma/templates/egress-service.yaml) but with per-mesh naming: `<release>-<mesh>-zoneproxy` (e.g., `kuma-payments-mesh-zoneproxy`).
This prevents name collisions when multiple meshes are deployed.

**Name length**: The 63-character limit only applies to **Service names** (DNS label, RFC 1123). Deployments, HPAs, PDBs, and ServiceAccounts use DNS subdomain names (253 chars) so they are not a concern. The zone proxy naming pattern `<release>-<mesh>-zoneproxy` should validate the Service name at install time:

```
{{ if gt (len (include "kuma.zoneProxy.serviceName" .)) 63 }} {{ fail "zone proxy service name exceeds 63 characters; use zoneProxy.service.name to set a shorter name" }} {{ end }}
```

To handle long mesh names, expose a `service.name` override per mesh entry (matching the existing `ingress.service.name` pattern in `_helpers.tpl:65-68`):

```yaml
meshes:
  - name: my-very-long-mesh-name-that-exceeds-limits
    ingress:
      enabled: true
      service:
        name: zp-long-mesh-ingress  # Override when auto-generated name is too long
    egress:
      enabled: true
      service:
        name: zp-long-mesh-egress
```

**Cost implication**: More LoadBalancers = higher cloud cost.
Users can use NodePort or Ingress controllers to reduce LB count if needed.

##### 3. Namespace Placement and MADR 093 Revert

All meshes' zone proxies are deployed in the `kuma-system` namespace.

> **This reverses [MADR 093](093-disallow-multiple-meshes-per-k8s-ns.md) (accepted).**

**Why**: The chart deploys zone proxies for multiple meshes into `kuma-system`. Requiring separate namespaces for each mesh's infrastructure components adds operational complexity with no benefit — zone proxies are infrastructure, not application workloads.

**What changes**: Zone proxies for different meshes coexist in `kuma-system`. More broadly, this is a general revert — multiple meshes per namespace are allowed everywhere, not just `kuma-system`.

**Collision handling**: The Workload controller fails with a clear error if a Workload name collision occurs across meshes. For zone proxies this is inherently avoided by the naming pattern `zone-proxy-<mesh>-<role>`, which guarantees unique Workload names per mesh.

For K8s naming constraints and `service.name` overrides, see Per-Mesh Services above.

| Aspect | Analysis |
|--------|----------|
| **Operations** | All Kuma components in one place |
| **RBAC** | Simple - one namespace to grant access |
| **Monitoring** | Single namespace to scrape metrics |
| **Multi-mesh** | MADR 093 revert allows coexistence; naming pattern avoids collisions |

**Note on sidecar injection**: Zone proxies run `kuma-dp` directly as standalone Deployments (the same way current ZoneIngress/ZoneEgress work). They connect to the CP via bootstrap, not via sidecar injection. The `kuma-system` namespace does **not** need sidecar injection enabled for zone proxies to function.

##### 4. Migration Path from Global Zone Proxies

**Additive migration**: The `meshes` config is added alongside the existing `ingress`/`egress` keys in the chart.

- **Phase 1**: Add `meshes` support, old keys still functional
- **Phase 2**: Add deprecation warnings for old keys
- **Phase 3**: Remove old keys in future major release

**Traffic migration** (switching live traffic from global to mesh-scoped zone proxies) is out of scope for this MADR and will be addressed separately.

##### 5. Mesh Deletion Handling

**Existing Protection**: Kuma already prevents mesh deletion when Dataplanes are attached.
The mesh validator returns an error: `"unable to delete mesh, there are still some dataplanes attached"`.
See [`pkg/core/managers/apis/mesh/mesh_validator.go#L70-L88`](https://github.com/kumahq/kuma/blob/master/pkg/core/managers/apis/mesh/mesh_validator.go#L70-L88).

**Mesh-Scoped Zone Proxy Benefit**: With zone proxies represented as Dataplane resources, this protection applies automatically.
No additional implementation is needed for mesh deletion handling - the existing safeguard covers the new deployment model.

**Note**: Current ZoneIngress/ZoneEgress resources are NOT covered by this protection (they're global-scoped).
The move to mesh-scoped Dataplanes resolves this gap.

### Question 1: Default Helm Installation Behavior

**`meshes: []`** — no zone proxies are deployed by default. Zone proxy deployment requires explicit opt-in by adding entries to the `meshes` list:

```yaml
controlPlane:
  mode: zone
  zone: zone-1

# Opt-in: add mesh entries to deploy zone proxies
meshes:
  - name: default
    ingress:
      enabled: true
    egress:
      enabled: true
```

This avoids a broken state where zone proxy Deployments target meshes that don't exist. The `skipMeshCreation` flag is orthogonal — it controls whether the CP auto-creates the `default` Mesh at startup, and is independent of zone proxy deployment.

## Decision

### Tooling Decisions

1. **Per-mesh templates**: Each entry in the `meshes` list is rendered by per-mesh templates that manage Deployment(s), Service(s), HPA, PDB, ServiceAccount, and optionally the Mesh resource and default policies.

2. **Per-mesh Services**: Each mesh gets its own Service/LoadBalancer for proper mTLS isolation.
   Sharing a LoadBalancer would require SNI-based cert selection which adds complexity.

3. **Namespace placement**: Deploy all meshes' zone proxies in `kuma-system` namespace.

4. **MADR 093 revert**: Allow multiple meshes per namespace everywhere (general revert, not scoped to zone proxies only). Handle Workload name collisions in the Workload controller with clear error messages rather than preventing the configuration.

5. **Additive migration**: The `meshes` config is added alongside existing `ingress`/`egress` keys. Old keys are deprecated and removed in a future major release.

6. **`combinedProxies` deployment option**: A map with the same shape as `ingress`/`egress` (enabled, podSpec, hpa, pdb, service, resources), merging ingress and egress into a single Deployment. Mutually exclusive with `ingress`/`egress` — Helm validates and errors if both are defined.

7. **Conditional mesh creation + `createPolicies` list**: `createMesh: true` renders a Mesh resource (for unfederated zones). `createPolicies` is a list of default policies to create with the mesh (see schema above for available policies). Both `createMesh` and `createPolicies` are ignored when `kdsGlobalAddress` is set (federated zones).

### Design Questions

1. **Helm defaults**: `meshes: []` — explicit opt-in, consistent with `ingress.enabled` / `egress.enabled`.
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
| Mesh proto (skipCreatingInitialPolicies) | `api/mesh/v1alpha1/mesh.proto` |
| Default mesh policies | `pkg/defaults/mesh/mesh.go` |
| K8s Ingress converter | `pkg/plugins/runtime/k8s/controllers/ingress_converter.go` |
| K8s Egress converter | `pkg/plugins/runtime/k8s/controllers/egress_converter.go` |
| Bootstrap generator (mesh validation) | `pkg/xds/bootstrap/generator.go` |
| Bootstrap handler (HTTP 422 response) | `pkg/xds/bootstrap/handler.go` |
| Helm ingress deployment | `deployments/charts/kuma/templates/ingress-deployment.yaml` |
| Helm egress deployment | `deployments/charts/kuma/templates/egress-deployment.yaml` |
| Helm values | `deployments/charts/kuma/values.yaml` |
| Annotations | `pkg/plugins/runtime/k8s/metadata/annotations.go` |

[^1]: Resource model MADR link will be backfilled when created.
[^2]: Zone proxy deployment topology MADR link will be backfilled when created.
