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
This aligns with [#15429](https://github.com/kumahq/kuma/issues/15429) where label-based service matching uses `Dataplane.Meta.Labels` to determine behavior, not resource type or inbound tags.

This document addresses the following questions:

1. Should zone ingress and egress be unified into a single zone proxy deployment?
2. Should zone proxies be deployed as sidecars to "fake" containers on Kubernetes?
3. How should zone proxies be deployed on Universal (VM/bare metal)?
4. Should `kuma.io/workload` annotation be required on zone proxies?
5. Should we continue supporting `kuma.io/ingress-public-address` annotation?
6. What should be the default Helm installation behavior for zone proxies?

## Design

### Question 1: Unified vs Separate Zone Proxies

#### Core Concept

Zone proxies are Envoy instances that receive listener configurations from the control plane.
The distinction between "ingress" and "egress" is fundamentally about **which listeners are generated**:

- **Ingress listeners**: Accept traffic from other zones and route to local services
- **Egress listeners**: Accept traffic from local services and route to other zones or external services

Since both are just listener configurations, a single Envoy instance can handle both roles simultaneously when configured with both listener types.

#### Options

| Option | Description |
|--------|-------------|
| **A. Unified zone proxy** | Single Dataplane type with `kuma.io/zone-proxy-role` label determining capabilities |
| **B. Separate proxies** | Maintain distinct ZoneIngress and ZoneEgress resource types (current implementation) |

#### Analysis

**Option A: Unified zone proxy (recommended)**

A unified zone proxy uses a single Dataplane resource with labels controlling behavior:

```yaml
type: Dataplane
mesh: payments-mesh
name: zone-proxy-1
labels:
  kuma.io/proxy-type: zoneproxy
  kuma.io/zone-proxy-role: both  # or: ingress, egress
networking:
  address: 10.0.0.1
  advertisedAddress: 203.0.113.1
  advertisedPort: 10001
```

Note: No inbound tags are required—service identification uses `Dataplane.Meta.Labels` per [#15429](https://github.com/kumahq/kuma/issues/15429).

The `kuma.io/zone-proxy-role` label controls which listeners the control plane generates:
- `both`: Generate both ingress and egress listeners (default)
- `ingress`: Generate only ingress listeners
- `egress`: Generate only egress listeners

Operators who need separate scaling for ingress vs egress traffic can deploy multiple zone proxies with different role labels.
This provides the same flexibility as separate types while simplifying the conceptual model.

- Advantages:
  - Simpler conceptual model—one proxy type instead of two
  - Reduced operational complexity—single deployment to manage
  - Flexible—can deploy combined or separate based on needs
  - Lower resource footprint when combined
  - Aligns with the reality that both are just Envoy with different listeners
  - Easier policy application—target all zone traffic or specific roles via labels
- Disadvantages:
  - Requires migration from current ZoneIngress/ZoneEgress resources
  - Combined deployment means single failure point (mitigated by replicas)

**Option B: Separate proxies (current implementation)**

Maintains distinct ZoneIngress and ZoneEgress resource types, each generating their specific listener configurations.

- Advantages:
  - No migration required
  - Independent scaling already understood by operators
  - Clear separation of concerns
- Disadvantages:
  - Two resource types that are conceptually the same thing (Envoy with listeners)
  - More complex mental model
  - Higher resource usage for typical deployments
  - Duplicated deployment configuration

#### Recommendation

**Option A: Unified zone proxy** - Zone proxies should be a single Dataplane type where the `kuma.io/zone-proxy-role` label determines capabilities.
This aligns with the technical reality that ingress/egress are listener configurations, not fundamentally different proxy types.

Operators can deploy:
- **Combined** (`role: both`): Single deployment handling all cross-zone traffic—recommended default
- **Separate** (`role: ingress` / `role: egress`): Independent deployments for scaling needs

### Question 2: Sidecar to Fake Container vs Standalone Deployment

#### Current Implementation

Zone proxies are deployed as standalone Kubernetes Deployments with dedicated pods containing only the kuma-dp container.
With the unified zone proxy model, this would consolidate to a single deployment template (or separate templates for operators choosing split deployments).

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
- Disadvantages:
  - Adds unnecessary resource overhead (pause container)
  - More complex deployment model
  - Confusing operational model (what is the "application"?)

**Option B: Standalone deployment (recommended)**
- Advantages:
  - Clear operational model - zone proxies are infrastructure
  - Independent scaling and lifecycle management
  - Familiar pattern for infrastructure operators
  - Lower resource overhead
  - Simpler debugging and monitoring
- Disadvantages:
  - Different pattern from application sidecars (minor)

**Option C: Direct pod without injection**
- Advantages:
  - Most minimal approach
- Disadvantages:
  - Loses benefits of injection webhook (consistent configuration)
  - Harder to maintain across versions

#### Recommendation

**Option B: Standalone deployment** - Zone proxies (whether unified or separate) should remain as standalone deployments without a fake application container.

### Question 3: Universal Deployment Model

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

With the move to mesh-scoped zone proxies and the unified model, the deployment changes:

1. Zone proxies become `Dataplane` resources with specific labels
2. Each resource must specify a `mesh` field
3. Deploy one zone proxy instance **per mesh** (not one per zone for all meshes)
4. The `kuma.io/zone-proxy-role` label determines ingress/egress/both capabilities

**Unified zone proxy (recommended)**:
```yaml
type: Dataplane
mesh: payments-mesh
name: zone-proxy-payments
labels:
  kuma.io/proxy-type: zoneproxy
  kuma.io/zone-proxy-role: both  # or: ingress, egress
networking:
  address: 10.0.0.1
  advertisedAddress: 203.0.113.1
  advertisedPort: 10001
```

Note: No inbound tags are required—service identification uses `Dataplane.Meta.Labels` per [#15429](https://github.com/kumahq/kuma/issues/15429).

**Separate deployments** (for operators needing independent scaling):
```yaml
# Ingress-only proxy
type: Dataplane
mesh: payments-mesh
name: zone-ingress-payments
labels:
  kuma.io/proxy-type: zoneproxy
  kuma.io/zone-proxy-role: ingress
networking:
  address: 10.0.0.1
  advertisedAddress: 203.0.113.1
  advertisedPort: 10001
---
# Egress-only proxy
type: Dataplane
mesh: payments-mesh
name: zone-egress-payments
labels:
  kuma.io/proxy-type: zoneproxy
  kuma.io/zone-proxy-role: egress
networking:
  address: 10.0.0.2
  advertisedAddress: 203.0.113.2
  advertisedPort: 10002
```

#### Migration Path

1. Deploy new mesh-scoped zone proxy for each mesh requiring cross-zone/external traffic
2. Control plane routes traffic to new proxy when MeshIdentity is enabled for that mesh
3. Scale down old global-scoped proxy after verification
4. Remove legacy ZoneIngress/ZoneEgress resources

#### Tooling Updates Required

- `kumactl` commands for zone proxy management
- Documentation for Universal deployment
- Migration guide from global to mesh-scoped

### Question 4: Require kuma.io/workload Annotation

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

**Option B: Not required** - Zone proxies should not require `kuma.io/workload` annotation.
They should be identified by the `kuma.io/proxy-type: zoneproxy` label and the `kuma.io/zone-proxy-role` label for role-specific targeting.
Policies can use extended `proxyTypes` in targetRef or label-based selectors.

Since zone proxies become mesh-scoped Dataplane resources, targeting a zone proxy for a specific mesh is straightforward: policies themselves are mesh-scoped, so a policy in `payments-mesh` automatically only applies to zone proxies in that mesh.
The `kuma.io/zone-proxy-role` label then filters to specific roles (ingress, egress, or both) within that mesh.

### Question 5: Support kuma.io/ingress-public-address

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

### Question 6: Default Helm Installation Behavior

#### Recommendation

**Multi-zone ready with unified zone proxy** - Deploy a single zone proxy (handling both ingress and egress) and skip default mesh creation:

```yaml
controlPlane:
  defaults:
    skipMeshCreation: true  # No default mesh

zoneProxy:
  enabled: true
  role: both  # or: ingress, egress, separate
```

The `role` setting controls deployment topology:
- `both` (default): Single Deployment handling both ingress and egress
- `ingress`: Single Deployment handling only ingress
- `egress`: Single Deployment handling only egress
- `separate`: Two Deployments (one ingress, one egress) for operators needing independent scaling

This provides multi-zone readiness while requiring explicit mesh creation.
The unified default reduces resource usage and operational complexity for typical deployments.

## Decision

1. **Unified zone proxy model**: Zone proxies should be a single Dataplane type where the `kuma.io/zone-proxy-role` label determines capabilities (ingress, egress, or both).
   Operators can deploy combined (`role: both`) or separate (`role: ingress` / `role: egress`) based on scaling needs.

2. **Standalone deployment**: Zone proxies should be deployed as standalone Kubernetes Deployments, not as sidecars to fake containers.
   The xDS protocol handles configuration updates regardless of deployment model.

3. **Universal deployment**: Use mesh-scoped Dataplane resources with zone proxy labels instead of global ZoneIngress/ZoneEgress.
   Deploy one zone proxy per mesh.

4. **kuma.io/workload not required**: Zone proxies are infrastructure components and should be targeted by the `kuma.io/proxy-type: zoneproxy` label and `kuma.io/zone-proxy-role` for role-specific targeting.

5. **Keep kuma.io/ingress-public-address**: Support the annotation as an escape hatch for complex network topologies, but document Service-based configuration as the primary method.

6. **Helm defaults**: Single unified zone proxy deployment with `role: both`, with option to split via `role: separate`.
   No default mesh created (`skipMeshCreation: true`).

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
