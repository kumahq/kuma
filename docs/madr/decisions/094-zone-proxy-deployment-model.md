# Zone Proxy Deployment Model

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/9030

## Context and Problem Statement

Currently, zone proxies are **global-scoped** resources, meaning a single ZoneIngress or ZoneEgress instance handles traffic for all meshes in a zone. This global nature creates fundamental limitations:

1. **Cannot issue MeshIdentity for zone egress**: MeshIdentity is mesh-scoped, but a global zone egress serves multiple meshes. This creates identity conflicts and prevents proper mTLS certificate issuance. See [MADR 090](090-zone-egress-identity.md) for detailed analysis.

2. **Cannot apply policies on zone proxies**: Kuma policies (MeshTrafficPermission, MeshTimeout, etc.) are mesh-scoped. A global zone proxy cannot be targeted by mesh-specific policies, limiting observability and traffic control for cross-zone communication.

To resolve these limitations, zone proxies are being changed to **mesh-scoped** resources represented as Dataplane resources with specific tags. This architectural change requires revisiting the deployment model for zone proxies.

This document addresses the following questions:

1. Should zone proxies be deployed as sidecars to "fake" containers on Kubernetes?
2. How should zone proxies be deployed on Universal (VM/bare metal)?
3. Should `kuma.io/workload` annotation be required on zone proxies?
4. Should we continue supporting `kuma.io/ingress-public-address` annotation?
5. What should be the default Helm installation behavior for zone proxies?
6. Can ZoneIngress and ZoneEgress run together in the same pod, or alongside MeshGateway?

## Design

### Question 1: Sidecar to Fake Container vs Standalone Deployment

#### Current Implementation

Zone proxies are deployed as standalone Kubernetes Deployments with dedicated pods containing only the kuma-dp container. This is defined in:
- `deployments/charts/kuma/templates/ingress-deployment.yaml`
- `deployments/charts/kuma/templates/egress-deployment.yaml`

#### Options

| Option | Description |
|--------|-------------|
| **A. Sidecar to pause container** | Inject zone proxy as sidecar to a pod with a minimal "sleep infinity" or pause container |
| **B. Standalone deployment** | Current approach - dedicated deployment with only kuma-dp |
| **C. Direct pod without injection** | Create pods directly without using sidecar injection webhook |

#### Analysis

The primary concern raised was that "mesh updates won't be updated automatically" with standalone deployments. This concern refers to **Kuma version upgrades** (e.g., `helm upgrade`):

- **Option A (sidecar to fake container)**: The sidecar injection webhook injects the proxy image at pod creation time. On `helm upgrade`, existing pods are **not automatically restarted** - the operator must manually trigger pod recreation to pick up the new version. This gives operators full control over upgrade timing, allowing careful rollouts that avoid dropping requests.

- **Option B (standalone deployment)**: The Deployment spec directly references the proxy image. On `helm upgrade`, the Deployment is updated, **triggering automatic pod rollout**. The rolling update strategy (`maxUnavailable`, `maxSurge`) controls how pods are replaced, preventing all-at-once restarts.

Option B is still recommended because zone proxies are infrastructure that should update with the control plane. The rolling update strategy provides sufficient control over the upgrade process.

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

**Option B: Standalone deployment** - Zone proxies should remain as standalone deployments without a fake application container.

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

With the move to mesh-scoped zone proxies, the deployment model changes:

1. Zone proxies become `Dataplane` resources with specific labels
2. Each resource must specify a `mesh` field
3. Deploy one zone proxy instance **per mesh** (not one per zone for all meshes)

Example (exact fields and label names determined by MADR for [issue #9028](https://github.com/kumahq/kuma/issues/9028)):
```yaml
type: Dataplane
mesh: payments-mesh
name: zone-egress-payments
labels:
  kuma.io/proxy-type: zone-egress
  kuma.io/zone: zone-1         # or kuma.io/zone-name
networking:
  address: 10.0.0.1
  advertisedAddress: 203.0.113.1
  advertisedPort: 10002
```

Label options for zone identification:
- `kuma.io/zone`: Consistent with existing zone label patterns
- `kuma.io/zone-name`: Explicit new label for zone proxies

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

**Option B: Not required** - Zone proxies should not require `kuma.io/workload` annotation. They should be identified by the `kuma.io/proxy-type` label (with values `zoneingress` or `zoneegress`) and targeted via policies using extended `proxyTypes` in targetRef or label-based selectors.

Since zone proxies become mesh-scoped Dataplane resources, targeting a zone proxy for a specific mesh is straightforward: policies themselves are mesh-scoped, so a policy in `payments-mesh` automatically only applies to zone proxies in that mesh. The `kuma.io/proxy-type` label then filters to just zone ingress or egress proxies within that mesh.

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

#### Current Defaults

From `deployments/charts/kuma/values.yaml`:
```yaml
controlPlane:
  mode: zone
  defaults:
    skipMeshCreation: false  # Creates "default" mesh

ingress:
  enabled: false  # NOT deployed by default

egress:
  enabled: false  # NOT deployed by default
```

#### Options

| Option | Default Mesh | Zone Ingress | Zone Egress | Description |
|--------|--------------|--------------|-------------|-------------|
| **A. Minimal (current)** | Yes | No | No | Opt-in zone proxies |
| **B. Multi-zone ready** | Yes | Yes | Yes | Include zone proxies |
| **C. No defaults** | No | No | No | Fully explicit |

#### Analysis

**Option A: Minimal**
- Advantages:
  - Most users start with single-zone; zone proxies unnecessary
  - Lower resource usage for simple deployments
- Disadvantages:
  - Requires explicit configuration for multi-zone
  - Users may not realize they need zone proxies

**Option B: Multi-zone ready (recommended)**
- Advantages:
  - Works out of the box for multi-zone
  - No additional configuration needed when adding zones
  - Consistent experience - default mesh is fully functional
  - Zone proxies are lightweight when not actively used
- Disadvantages:
  - Slightly higher resource usage for single-zone users

**Option C: No defaults**
- Advantages:
  - Maximum explicitness
- Disadvantages:
  - Poor getting-started experience
  - Breaks existing workflows

#### Recommendation

**Option B: Multi-zone ready** - Install zone-ingress and zone-egress with the default mesh:

```yaml
controlPlane:
  defaults:
    skipMeshCreation: false  # Creates "default" mesh

ingress:
  enabled: true  # Deployed with default mesh

egress:
  enabled: true  # Deployed with default mesh
```

This ensures the default mesh is fully functional for multi-zone scenarios out of the box.

### Question 6: Combined Deployments (Ingress + Egress or with Gateway)

#### Options

| Option | Description | Scaling | Resource Usage |
|--------|-------------|---------|----------------|
| **A. Separate deployments** | Independent ZoneIngress and ZoneEgress | Independent | Higher |
| **B. Combined pod** | Single pod with both ingress and egress | Together | Lower |
| **C. Collocate with Gateway** | Zone proxy in same pod as MeshGateway | Shared | Lowest |

#### Analysis

**Option A: Separate deployments (recommended default)**
- Advantages:
  - Independent scaling based on actual traffic patterns
  - Simpler debugging and monitoring
  - Clear failure domains
  - Traffic patterns differ (ingress vs egress load)
- Disadvantages:
  - Higher total resource usage

**Option B: Combined pod (optional)**
- Advantages:
  - Lower resource usage
  - Simpler for small deployments
  - Useful for resource-constrained environments
- Disadvantages:
  - Cannot scale independently
  - Mixed concerns in single pod

**Option C: Collocate with Gateway**
- Advantages:
  - Lowest resource usage
- Disadvantages:
  - High complexity
  - Policy conflicts between gateway and zone proxy roles
  - Very different traffic patterns and scaling needs
  - Debugging nightmare

#### Recommendation

**Option A: Separate deployments by default**, with Option B available as an opt-in for resource-constrained environments.

Implementation approach:
- Default: Separate ZoneIngress and ZoneEgress deployments
- Optional Helm flag: `zoneProxy.combined: true` for single deployment with both roles
- Do NOT support collocation with MeshGateway due to complexity

## Security Implications and Review

### mTLS and Identity

With mesh-scoped zone proxies:
- Each zone proxy has a single mesh identity (vs. current multi-mesh identity)
- Cleaner trust model - proxy identity matches the mesh it serves
- Better alignment with SPIRE and external identity providers
- See MADR 090 (Zone Egress Identity) for detailed security analysis

### Network Exposure

- Zone proxies expose network endpoints for cross-zone communication
- `kuma.io/ingress-public-address` override requires trust in annotation values
- Recommendation: Validate annotation values match expected patterns

### Resource Access

- Zone proxies are privileged infrastructure components
- Should be deployed in dedicated namespace with restricted RBAC
- No change from current security model

## Reliability Implications

### High Availability

- Separate deployments enable independent scaling and failover
- Recommendation: Document minimum replica count for production (2+)
- Pod disruption budgets should be configured

### Failure Domains

- Mesh-scoped proxies isolate failures to single mesh
- Current global-scoped proxy failure affects all meshes
- This is an improvement in fault isolation

### Migration Risk

- Migration from global to mesh-scoped requires careful coordination
- Recommendation: Support running both during migration period
- Provide clear rollback procedures

## Implications for Kong Mesh

### MinK (Mesh in Konnect)

- MinK is always multi-zone by design
- Zone proxies enabled by default aligns with MinK requirements
- Charts at Kong/mink-charts repository use the same defaults
- No special overrides needed

### Kong Mesh Enterprise

- Same deployment model applies
- May need documentation updates for enterprise-specific guidance
- No special handling required

## Decision

1. **Standalone deployment**: Zone proxies should be deployed as standalone Kubernetes Deployments, not as sidecars to fake containers. The xDS protocol handles configuration updates regardless of deployment model.

2. **Universal deployment**: Use mesh-scoped Dataplane resources instead of global ZoneIngress/ZoneEgress. Deploy one zone proxy per mesh.

3. **kuma.io/workload not required**: Zone proxies are infrastructure components and should be targeted by the `kuma.io/proxy-type` label (e.g., `zoneegress`).

4. **Keep kuma.io/ingress-public-address**: Support the annotation as an escape hatch for complex network topologies, but document Service-based configuration as the primary method.

5. **Multi-zone ready Helm defaults**: Zone proxies (ingress and egress) enabled by default with the default mesh, ensuring a fully functional multi-zone setup out of the box.

6. **Separate deployments**: ZoneIngress and ZoneEgress should be separate deployments by default, with an optional combined mode for resource-constrained environments. Do not support collocation with MeshGateway.

## Notes

### Related Issues and MADRs

- [Issue #9028](https://github.com/kumahq/kuma/issues/9028): Dataplane fields for zone proxies
- [Issue #9029](https://github.com/kumahq/kuma/issues/9029): Policies on zone proxies
- [Issue #9032](https://github.com/kumahq/kuma/issues/9032): MeshIdentity for zone egress
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
