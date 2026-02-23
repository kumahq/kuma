# Shared telemetry backend resource for observability policies

* Status: accepted

Technical Story: https://github.com/Kong/kong-mesh/issues/9163

## Context and problem statement

Kuma has three observability policies that can export telemetry to an OpenTelemetry collector: MeshMetric, MeshTrace, and MeshAccessLog. Each policy defines the OTel collector endpoint independently in its own spec.

In practice, most deployments point all three policies at the same collector. The endpoint string (`otel-collector.observability:4317`) is duplicated across three policy instances per mesh. When the collector address changes (new namespace, different port), the operator updates three places. In multi-mesh setups, multiply by the number of meshes.

Each policy also has a slightly different OTel backend struct:

| Policy        | OTel backend fields              |
|---------------|----------------------------------|
| MeshMetric    | `endpoint`, `refreshInterval`    |
| MeshTrace     | `endpoint`                       |
| MeshAccessLog | `endpoint`, `attributes`, `body` |

The endpoint is the common denominator. Signal-specific fields (`refreshInterval`, `attributes`, `body`) are unique to each policy and don't belong in a shared resource.

Beyond endpoint duplication, there's a missing abstraction for collector connection settings. Today none of the policies support configuring:

- TLS settings for the collector connection (custom CA, client certs)
- Authentication headers (bearer tokens for managed OTel backends)
- Protocol preference (gRPC vs HTTP) - policies are gRPC-only today

These would need to be added to all three policies individually, tripling the work and the surface area.

### User stories

- As a mesh operator, I want to deploy one OTel collector (DaemonSet or Deployment) and point MeshMetric, MeshTrace, and MeshAccessLog at it without duplicating the endpoint in three places.
- As a mesh operator, I want to update my collector address in one place when the observability team moves it to a different namespace or changes the port.
- As a mesh operator using a managed OTel backend (Grafana Cloud, Datadog, etc.), I want to configure a bearer token for collector auth.
- As a mesh operator running multi-zone, I want each zone to have its own collector config without duplicating endpoints across zone-scoped policies.
- As a mesh operator, I want to roll out signals incrementally - start with metrics, verify it works, then add tracing and access logging against the same collector.

## Design

### Option A: New shared telemetry backend resource (recommended)

Introduce a new mesh-scoped resource `MeshTelemetryBackend` that defines an OTel collector endpoint with connection settings. Observability policies reference it via a `backendRef` field.

#### Resource definition

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshTelemetryBackend
metadata:
  name: main-collector
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
spec:
  # Only OpenTelemetry is supported.
  type: OpenTelemetry
  openTelemetry:
    # Collector endpoint - supports gRPC and HTTP/HTTPS
    endpoint:
      # Address of the collector (hostname or IP)
      address: otel-collector.observability
      # Port number
      port: 4317
      # Path for HTTP/HTTPS endpoints (e.g., /v1/traces, /v1/metrics, /v1/logs).
      # Ignored for gRPC.
      # +optional
      path: ""
```

The spec uses the same `type` discriminator + type-specific struct pattern as MeshMetric, MeshTrace, MeshAccessLog, and MeshLoadBalancingStrategy. Only `type: OpenTelemetry` is supported.

The `endpoint` is structured (address + port + path) instead of a raw string so we validate each component separately. Today MeshMetric/MeshAccessLog split the endpoint string on `:`, and MeshTrace has an unreleased URL parser for HTTP endpoints that's being removed. A structured format avoids these inconsistencies.

#### Policy reference

Each policy's OTel backend gets an optional `backendRef` field. When set, the endpoint comes from the referenced MeshTelemetryBackend. Signal-specific fields remain inline.

##### MeshMetric

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshMetric
metadata:
  name: all-metrics
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
spec:
  targetRef:
    kind: Mesh
  default:
    backends:
    - type: OpenTelemetry
      openTelemetry:
        backendRef:
          kind: MeshTelemetryBackend
          name: main-collector
        refreshInterval: 30s      # signal-specific, stays inline
```

##### MeshTrace

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshTrace
metadata:
  name: all-traces
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
spec:
  targetRef:
    kind: Mesh
  default:
    backends:
    - type: OpenTelemetry
      openTelemetry:
        backendRef:
          kind: MeshTelemetryBackend
          name: main-collector
    sampling:
      overall: 80             # signal-specific, stays inline
```

##### MeshAccessLog

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshAccessLog
metadata:
  name: all-access-logs
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
spec:
  targetRef:
    kind: Mesh
  default:
    backends:
    - type: OpenTelemetry
      openTelemetry:
        backendRef:
          kind: MeshTelemetryBackend
          name: main-collector
        attributes:              # signal-specific, stays inline
        - key: mesh
          value: "%KUMA_MESH%"
```

#### Required policy changes

None of the current policy OTel backend structs have a `backendRef` field. Today they only have an inline `endpoint` string:

- MeshMetric `OpenTelemetryBackend`: `endpoint` + `refreshInterval`
- MeshTrace `OpenTelemetryBackend`: `endpoint`
- MeshAccessLog `OtelBackend`: `endpoint` + `attributes` + `body`

Each struct needs a new optional `backendRef` field. The inline `endpoint` remains supported. Validation enforces mutual exclusivity: either `endpoint` or `backendRef`, not both.

```go
// In each policy's OTel backend struct
type OpenTelemetryBackend struct {
    // Inline endpoint (existing, backward compatible)
    Endpoint string `json:"endpoint,omitempty"`
    // OR reference to shared backend
    BackendRef *common_api.TargetRef `json:"backendRef,omitempty"`
    // ... signal-specific fields unchanged
}
```

#### Resolution during xDS generation

1. CP loads all MeshTelemetryBackend resources into MeshContext (same as MeshService, MeshExternalService)
2. During policy plugin's `Apply()`:
   - If `backendRef` is set: look up MeshTelemetryBackend by name from `ctx.Mesh.Resources.MeshLocalResources`
   - Extract endpoint from the type-specific config (e.g., `spec.openTelemetry.endpoint`)
   - Convert to the same `*core_xds.Endpoint` struct used today
3. Proceed with existing cluster/listener creation - no changes downstream

This follows the same pattern as BackendRef resolution in routing policies (`pkg/plugins/policies/core/rules/resolve/backendref.go`).

#### Resource characteristics

| Property        | Value                                                                                |
|-----------------|--------------------------------------------------------------------------------------|
| `IsPolicy`      | false (it's a resource, not a policy)                                                |
| `Scope`         | Mesh                                                                                 |
| `HasStatus`     | false                                                                                |
| `KDSFlags`      | <code>GlobalToZonesFlag &#124; ZoneToGlobalFlag</code> (same as MeshExternalService) |
| `ShortName`     | `mtb`                                                                                |
| `IsDestination` | false                                                                                |

Placed in `pkg/core/resources/apis/meshtelemetrybackend/` following the same directory structure as MeshExternalService.

#### Advantages

- Purpose-built API for the use case
- Single place for endpoint + connection settings
- Extensible: TLS, auth, protocol preferences added once, used by all policies
- Follows established resource patterns (tooling, KDS, REST API, kumactl come for free)
- Multi-zone: syncs via KDS, can be zone-scoped
- Backward compatible with inline endpoints
- Separates infrastructure config (where the collector is) from policy config (what to send)

#### Disadvantages

- New resource type: CRD, REST API, kumactl support, documentation
- Three policy plugins need modification (add backendRef + resolution logic)
- Users learn a new concept
- More indirection (policy -> resource -> endpoint)

### Option B: Named backends on the Mesh resource

Add OTel collector configurations to the Mesh spec. Policies reference by name.

```yaml
kind: Mesh
metadata:
  name: default
spec:
  # Existing fields...
  otelCollectors:
  - name: main-collector
    endpoint: otel-collector.observability:4317
  - name: traces-only
    endpoint: https://jaeger.observability:4318/v1/traces
```

Policy reference:

```yaml
kind: MeshMetric
spec:
  default:
    backends:
    - type: OpenTelemetry
      openTelemetry:
        collectorRef: main-collector   # references Mesh.spec.otelCollectors[].name
        refreshInterval: 30s
```

#### Advantages

- No new resource type - simpler implementation
- Mesh is already loaded in MeshContext, resolution is trivial
- Precedent: `Mesh.spec.metrics.backends` uses a name-reference pattern (though this is legacy)
- Faster to implement

#### Disadvantages

- Mesh resource is already large (networking, mtls, metrics, tracing, logging, routing, meshServices sections). Adding more config makes it harder to manage.
- Can't apply separate RBAC (who can modify Mesh vs. who can configure collectors)
- All collector configs in one Mesh object - doesn't scale for different teams owning different collectors
- Mesh is proto-based (`api/mesh/v1alpha1/mesh.proto`), adding structured config there requires proto changes + regeneration
- Adding the metrics.backends precedent was arguably a mistake - the whole point of MeshMetric policy was to move this OUT of Mesh

### Option C: Reference MeshExternalService

Model the OTel collector as a MeshExternalService. Policies reference it as a destination.

```yaml
kind: MeshExternalService
metadata:
  name: otel-collector
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
spec:
  match:
    type: HostnameGenerator
    port: 4317
    protocol: grpc
  endpoints:
  - address: otel-collector.observability
    port: 4317
```

Policy reference:

```yaml
kind: MeshMetric
spec:
  default:
    backends:
    - type: OpenTelemetry
      openTelemetry:
        backendRef:
          kind: MeshExternalService
          name: otel-collector
          port: 4317
        refreshInterval: 30s
```

#### Advantages

- No new resource type
- Reuses existing TLS, hostname, and multi-zone infrastructure
- MeshExternalService already syncs via KDS

#### Disadvantages

- Semantically wrong: an OTel collector is infrastructure, not an "external service" being called by application traffic. Users will be confused.
- MeshExternalService is designed for traffic routing (VIP allocation, passthrough, etc.) - concepts that don't map to a telemetry backend
- Doesn't capture OTel-specific semantics: protocol preference (gRPC vs HTTP on the same collector), path for HTTP endpoints, signal-specific routing
- An OTel collector typically has multiple ports (4317 gRPC, 4318 HTTP). MeshExternalService.match.port is singular - you'd need multiple MeshExternalService resources for the same collector.
- If the collector is NOT meshed (recommended - disable sidecar injection), having a MeshExternalService creates routing expectations that conflict with direct connectivity.

### Option D: Keep inline configuration (status quo)

Accept the endpoint duplication. Each policy manages its own OTel backend config.

#### Advantages

- Zero effort
- No new concepts
- Each policy is independently deployable

#### Disadvantages

- Endpoint duplication (3 policies per mesh, more in multi-mesh)
- Coordinated updates when collector changes
- No shared place for TLS/auth (when we add these, we add them 3x)
- Error-prone at scale

## Security implications and review

### Option A (MeshTelemetryBackend)

- Secret references: When auth support is added (bearer tokens, client certs), the resource would reference Kubernetes Secrets or Kuma secrets. Standard secret handling patterns from MeshGateway TLS apply.
- RBAC: Separate resource means separate RBAC rules. An operator can grant "create MeshTelemetryBackend" without granting "modify Mesh". This is better than Option B where Mesh modification is required.
- Cross-mesh isolation: Mesh-scoped resource ensures one mesh's collector config can't affect another mesh.

### Option B (Mesh-level)

- Modifying collector config requires Mesh write access, which is typically restricted to cluster admins. This may be too restrictive for teams that should manage their own observability.

### Options C & D

- No new security concerns.

## Reliability implications

### Option A

- Dangling references: If a MeshTelemetryBackend is deleted while policies reference it, policies lose their backend config. Mitigation: validation webhook that prevents deletion of in-use backends, or status field listing referencing policies.
- Resource sync: MeshTelemetryBackend syncs via KDS. If sync is delayed, newly created policies in a zone may not find their backend. Same behavior as MeshExternalService references today.

### Option B

- Mesh is a critical resource. Adding more config increases blast radius of Mesh modifications.

### Options C & D

- No new reliability concerns.

## Implications for Kong Mesh

- Option A: Kong Mesh would need to include MeshTelemetryBackend in its resource list.
- Options B/C/D: No Kong Mesh-specific implications.

## Decision

Option A: New MeshTelemetryBackend resource.

It's more work upfront but it's the right abstraction. The endpoint config is genuinely shared infrastructure that doesn't belong in any single policy or in the already-bloated Mesh resource. The Kuma tooling (`tools/policy-gen/bootstrap`, `make generate`) makes creating new resource types mechanical - most of the code is generated.

The resource uses a `type` discriminator, same as every other Kuma policy with multiple backend types. Only `type: OpenTelemetry` is supported, with OTel-specific connection info nested under `openTelemetry`.

Start with a minimal spec (endpoint only, no TLS/auth). Add TLS, auth, and protocol fields as follow-up work. HTTP endpoint support will only exist in MeshTelemetryBackend - the unreleased HTTP support in MeshTrace is being removed, and it won't be added to MeshMetric or MeshAccessLog.

Signal-specific config (refreshInterval, attributes, body, sampling) stays in each policy. The shared resource only handles "where is the backend and how do I connect to it."

## Notes

### Initial scope (MVP)

The first implementation should be minimal:

1. MeshTelemetryBackend resource with `type: OpenTelemetry` and `endpoint` (address + port) only
2. `backendRef` field added to all three policy OTel backends
3. Validation: `endpoint` XOR `backendRef` (mutual exclusivity)
4. Resolution in each policy's Apply()
5. Inline `endpoint` remains fully supported

TLS, auth, and protocol fields are follow-up work.

### Type mismatch between resource and policy

Since OpenTelemetry is the only supported type, this MADR does not cover what happens when a MeshTelemetryBackend's type doesn't match the referencing policy's backend type (e.g., a MeshTrace with `type: Zipkin` pointing at a MeshTelemetryBackend with `type: OpenTelemetry`). That's a separate design decision for when additional types are introduced.

### Naming

`MeshTelemetryBackend` was chosen to avoid vendor-locking the resource name to OpenTelemetry. The `type` discriminator makes it clear that current config is OTel-specific (nested under `openTelemetry`), while the resource name stays generic.

Rejected alternatives:
- `MeshOTelBackend` - locks the name to OpenTelemetry; if we add Zipkin or Datadog backend support, the name becomes misleading
- `MeshOpenTelemetryBackend` - same vendor-lock problem, plus too long
- `MeshOTelCollector` - "collector" is one deployment model; the resource represents any telemetry-receiving endpoint
- `MeshObservabilityBackend` - too abstract (24 chars), doesn't add clarity over "telemetry"
- `MeshSignalBackend` - "signal" is OTel jargon, not widely understood outside the OTel community
- `MeshCollectorBackend` - not all backends are "collectors" (e.g., managed SaaS endpoints)

`MeshTelemetryBackend` follows Kuma's established naming: infrastructure resources describe what they manage (`MeshExternalService`, `MeshIdentity`, `MeshTrust`). This resource manages telemetry backend infrastructure.

### Relationship to CP metrics export

The control plane's own OTel metric export (via `KUMA_TRACING_OPENTELEMETRY_ENABLED` and standard `OTEL_EXPORTER_OTLP_*` env vars) is separate. MeshTelemetryBackend is for data plane telemetry policies only. CP observability config is environment-level, not mesh-scoped.
