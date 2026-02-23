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
- As a mesh operator, I want the data plane's OTel export configured from standard `OTEL_EXPORTER_OTLP_*` env vars injected into my pods so I don't duplicate collector config between application instrumentation and the mesh.

## Design

### Option A: New shared telemetry backend resource (recommended)

Introduce a new mesh-scoped resource `MeshOpenTelemetryBackend` that defines an OTel collector endpoint with connection settings. Observability policies reference it via a `backendRef` field.

#### Resource definition

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshOpenTelemetryBackend
metadata:
  name: main-collector
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
spec:
  # Collector endpoint
  endpoint:
    # Address of the collector (hostname or IP)
    address: otel-collector.observability
    # Port number
    port: 4317
    # Base path prefix for HTTP endpoints. The CP appends the
    # signal-specific suffix (/v1/traces, /v1/metrics, /v1/logs)
    # automatically, matching OTEL_EXPORTER_OTLP_ENDPOINT semantics.
    # Ignored for gRPC.
    # +optional
    path: ""
  # Transport protocol. Drives whether Envoy uses GrpcService or
  # HttpService in the generated xDS config.
  # Values: grpc, http. Default: grpc.
  # +optional
  protocol: grpc
```

No `type` discriminator - the resource name itself is the type. If other telemetry backend types are needed later (Zipkin, Datadog), they become separate resources (`MeshZipkinBackend`, etc.). The `backendRef.kind` field enforces type matching at the schema level - a policy's OTel backend can only reference a `MeshOpenTelemetryBackend`, not a hypothetical `MeshZipkinBackend`.

The `endpoint` is structured (address + port + path) instead of a raw string so we validate each component separately. Today MeshMetric/MeshAccessLog split the endpoint string on `:`, and MeshTrace has an unreleased URL parser for HTTP endpoints that's being removed. A structured format avoids these inconsistencies.

The `protocol` field selects gRPC or HTTP transport. Envoy's OTel extensions use separate `grpc_service` and `http_service` fields in their protobuf config - the CP needs to know which one to generate. Port-based inference (4317 = gRPC, 4318 = HTTP) would break for non-standard setups, so an explicit field is safer. Default is `grpc` for backward compatibility with existing Kuma behavior.

When `protocol: http`, the `path` field is a base path prefix. The CP appends the signal-specific suffix (`/v1/traces`, `/v1/metrics`, `/v1/logs`) during xDS generation, matching how the OTel SDK handles `OTEL_EXPORTER_OTLP_ENDPOINT`. An empty path means the standard paths are used directly. For gRPC, paths are irrelevant - gRPC routes by protobuf service name.

#### Policy reference

Each policy's OTel backend gets an optional `backendRef` field. When set, the endpoint comes from the referenced MeshOpenTelemetryBackend. Signal-specific fields remain inline.

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
          kind: MeshOpenTelemetryBackend
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
          kind: MeshOpenTelemetryBackend
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
          kind: MeshOpenTelemetryBackend
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

Each struct needs a new optional `backendRef` field. The inline `endpoint` remains supported but is deprecated starting in 2.14 and will be removed in 3.0. Validation enforces mutual exclusivity: either `endpoint` or `backendRef`, not both.

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

1. CP loads all MeshOpenTelemetryBackend resources into MeshContext (same as MeshService, MeshExternalService)
2. During policy plugin's `Apply()`:
   - If `backendRef` is set: look up MeshOpenTelemetryBackend by name from `ctx.Mesh.Resources.MeshLocalResources`
   - Extract endpoint from `spec.endpoint`
   - Convert to the same `*core_xds.Endpoint` struct used today
3. Proceed with existing cluster/listener creation - no changes downstream

This follows the same pattern as TargetRef resolution in policies (`pkg/plugins/policies/core/rules/resolve/targetref.go`).

#### Resource characteristics

| Property        | Value                                                                                |
|-----------------|--------------------------------------------------------------------------------------|
| `IsPolicy`      | false (it's a resource, not a policy)                                                |
| `Scope`         | Mesh                                                                                 |
| `HasStatus`     | false                                                                                |
| `KDSFlags`      | <code>GlobalToZonesFlag &#124; ZoneToGlobalFlag</code> (same as MeshExternalService) |
| `ShortName`     | `motb`                                                                               |
| `IsDestination` | false                                                                                |

Placed in `pkg/core/resources/apis/meshopentelemetrybackend/` following the same directory structure as MeshExternalService.

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

### Option B: Named backends on the Mesh resource (rejected)

Add OTel collector configurations to the Mesh spec. Policies reference by name.

This goes against the current direction. We're actively moving config OUT of Mesh (MeshMetric replaced `Mesh.spec.metrics`, MeshTrace replaced `Mesh.spec.tracing`, MeshAccessLog replaced `Mesh.spec.logging`). Adding more config to Mesh is not an option.

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
- Adding `metrics.backends` to Mesh was a mistake - the whole point of MeshMetric was to move this OUT of Mesh. We should not repeat it.

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

### Option A (MeshOpenTelemetryBackend)

- Secret references: When auth support is added (bearer tokens, client certs), the resource would reference Kubernetes Secrets or Kuma secrets. Standard secret handling patterns from MeshGateway TLS apply.
- RBAC: Separate resource means separate RBAC rules. An operator can grant "create MeshOpenTelemetryBackend" without granting "modify Mesh". This is better than Option B where Mesh modification is required.
- Cross-mesh isolation: Mesh-scoped resource ensures one mesh's collector config can't affect another mesh.

### Option B (Mesh-level)

- Modifying collector config requires Mesh write access, which is typically restricted to cluster admins. This may be too restrictive for teams that should manage their own observability.

### Options C & D

- No new security concerns.

## Reliability implications

### Option A

- Dangling references: If a MeshOpenTelemetryBackend is deleted while policies reference it, policies lose their backend config. This matches how Kuma handles all other dangling references today - when a `BackendRef` can't resolve, the reference is silently dropped and the policy proceeds without that backend. The CP should log at Info level when a referenced MeshOpenTelemetryBackend is not found during xDS generation. No cross-reference validation webhooks - Kuma's existing pattern is to not block deletion of referenced resources.
- Resource sync: MeshOpenTelemetryBackend syncs via KDS. If sync is delayed, newly created policies in a zone may not find their backend. Same behavior as MeshExternalService references today.

### Option B

- Mesh is a critical resource. Adding more config increases blast radius of Mesh modifications.

### Options C & D

- No new reliability concerns.

## Implications for Kong Mesh

- Option A: Kong Mesh would need to include MeshOpenTelemetryBackend in its resource list.
- Options B/C/D: No Kong Mesh-specific implications.

## Decision

Option A: New MeshOpenTelemetryBackend resource.

It's more work upfront but it's the right abstraction. The endpoint config is shared infrastructure that doesn't belong in any single policy or in the already-bloated Mesh resource. The Kuma tooling (`tools/policy-gen/bootstrap`, `make generate`) makes creating new resource types mechanical - most of the code is generated.

The resource is named after its backend type (`MeshOpenTelemetryBackend`) rather than using a generic name with a type discriminator. If other backend types are needed later, they become separate resources - and `backendRef.kind` enforces type matching at the schema level.

Start with a minimal spec (endpoint only, no TLS/auth). Add TLS and auth fields as follow-up work. HTTP endpoint support will only exist in MeshOpenTelemetryBackend - the unreleased HTTP support in MeshTrace is being removed, and it won't be added to MeshMetric or MeshAccessLog.

Signal-specific config (refreshInterval, attributes, body, sampling) stays in each policy. The shared resource only handles "where is the backend and how do I connect to it."

## Notes

### Initial scope (MVP)

The first implementation should be minimal:

1. MeshOpenTelemetryBackend resource with `endpoint` (address + port + path) and `protocol` (grpc/http)
2. `backendRef` field added to all three policy OTel backends
3. Validation: `endpoint` XOR `backendRef` (mutual exclusivity)
4. Resolution in each policy's Apply()
5. Inline `endpoint` remains supported but deprecated (removed in 3.0)

TLS and auth fields are follow-up work.

### Naming

`MeshOpenTelemetryBackend` names the resource after its backend type. Since the resource can only be referenced from the OTel backend section of policies, a generic name adds no value. Type matching is enforced at the schema level - `backendRef.kind: MeshOpenTelemetryBackend` can only appear in an OTel backend block. If Zipkin or Datadog backend types are needed later, they become separate resources (`MeshZipkinBackend`, etc.) with their own `backendRef.kind`.

Rejected alternatives:
- `MeshTelemetryBackend` (with `type` discriminator) - a generic name + type discriminator can't enforce type matching at the schema level. A policy's OTel backend could reference a Zipkin-typed backend, and this mismatch would only be caught by runtime validation, not by the API structure.
- `MeshOTelBackend` - abbreviation is less readable than the full name
- `MeshOTelCollector` - "collector" is one deployment model, the resource is for any telemetry-receiving endpoint
- `MeshObservabilityBackend` - too abstract, doesn't add clarity over the specific protocol name
- `MeshCollectorBackend` - not all backends are "collectors" (e.g., managed SaaS endpoints)

### Signal-specific fields stay in policies

MeshAccessLog's `body` field is logs-only (OTel LogRecord body). It holds Envoy access log command operators (`%START_TIME%`, `%UPSTREAM_HOST%`) that control what each log record says. This is per-record content formatting, not connection config - it stays in MeshAccessLog.

MeshAccessLog's `attributes` are also logs-only (mapped to OTel `KeyValueList` with command operator substitution). MeshTrace has `Tags` (mapped to Envoy `CustomTag` on the HCM tracing config). MeshMetric has no OTel attributes (uses kuma-dp `ExtraLabels`). Each signal uses a completely different Envoy mechanism for attributes, so there's no shared abstraction that works across all three.

Shared resource attributes (mesh name, zone) in MeshOpenTelemetryBackend are conceptually sound but practically complex - Envoy handles resource attributes differently per signal (direct `KeyValueList` on access logger, `resource_detectors` extension on tracer, kuma-dp labels for metrics). This is follow-up work.

### Standard OTel environment variables

The OTel specification defines environment variables (`OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_EXPORTER_OTLP_PROTOCOL`, `OTEL_EXPORTER_OTLP_HEADERS`, etc.) that configure where an OTel SDK sends telemetry. The OpenTelemetry Operator for Kubernetes injects these env vars into the first container in the pod spec by default (`Containers[0]`), which is typically the application container because mesh sidecar injectors append their containers. The Operator has no sidecar awareness - this targeting is a side effect of container ordering, not intentional filtering.

These env vars configure the application's OTel SDK, not the sidecar proxy. Envoy doesn't use the OTel SDK and doesn't read `OTEL_EXPORTER_OTLP_*` env vars - its telemetry export is configured through xDS pushed by the control plane. No service mesh today (Istio, Linkerd, or Kuma) reads pod env vars to configure sidecar proxy telemetry. This is an architectural split, not an oversight:

| | Application telemetry | Proxy telemetry |
|---|---|---|
| Config source | OTel env vars on app container | Control plane via xDS |
| Who exports | OTel SDK in app code | Envoy sidecar |
| What it captures | App-level spans, custom metrics | L4/L7 request metrics, access logs, mesh traces |

A mesh operator might want both paths pointing at the same collector without configuring each separately. The kuma-dp process already has a `DynamicMetadata` transport (`map[string]string`) that carries key-value pairs from the data plane to the control plane during bootstrap. This transport could carry `OTEL_EXPORTER_OTLP_*` values from the pod environment to the CP, which would then generate matching xDS config.

Env var auto-detection is out of scope for the MVP. MeshOpenTelemetryBackend is the explicit configuration path for proxy telemetry. Env var integration would build on top of it as follow-up work - the resource is still the canonical config for a backend's connection settings, whether set manually or populated from env vars.

### Relationship to CP metrics export

The control plane's own OTel metric export (via `KUMA_TRACING_OPENTELEMETRY_ENABLED` and standard `OTEL_EXPORTER_OTLP_*` env vars) is separate. MeshOpenTelemetryBackend is for data plane telemetry policies only. CP observability config is environment-level, not mesh-scoped.
