# Shared telemetry backend resource for observability policies

* Status: accepted

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

1. As a mesh operator, I want to deploy one OTel collector and point MeshMetric, MeshTrace, and MeshAccessLog at it without duplicating the endpoint in three places. I want to roll out signals incrementally - start with metrics, verify it works, then add tracing and access logging against the same backend.
2. As a mesh operator, I want to update my collector address in one place when the observability team moves it to a different namespace or changes the port.
3. As a mesh operator using a managed OTel backend (Grafana Cloud, Datadog, etc.), I want to configure connection settings (TLS, auth, protocol) for my collector in one place rather than across three policies.
4. As a mesh operator running multi-zone, I want each zone to have its own collector config without duplicating endpoints across zone-scoped policies.
5. As a mesh operator, I want to know when a policy references a backend that no longer exists so I don't silently lose telemetry.
6. As a mesh operator, I want the data plane's OTel export configured from standard [`OTEL_EXPORTER_OTLP_*`](https://opentelemetry.io/docs/specs/otel/protocol/exporter/) env vars injected into my pods so I don't duplicate collector config between application instrumentation and the mesh.

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

No `type` discriminator - the resource name itself is the type. If other telemetry backend types are needed later (Zipkin, Datadog), they become separate resources (`MeshZipkinBackend`, etc.). The `backendRef.kind` field enforces type matching via the admission webhook - the validator rejects any `backendRef` where `kind` is not `MeshOpenTelemetryBackend` inside an OTel backend block.

The `endpoint` is structured (address + port + path) instead of a raw string so we validate each component separately. Today MeshMetric/MeshAccessLog split the endpoint string on `:`, and MeshTrace has an unreleased URL parser for HTTP endpoints that's being removed. A structured format avoids these inconsistencies.

The `protocol` field selects gRPC or HTTP transport. Envoy's OTel extensions use separate `grpc_service` and `http_service` fields in their protobuf config - the CP needs to know which one to generate. Port-based inference (4317 = gRPC, 4318 = HTTP) would break for non-standard setups, so an explicit field is safer. Default is `grpc` for backward compatibility with existing Kuma behavior.

The OTLP/HTTP spec defines two content encodings: binary protobuf (`application/x-protobuf`) and JSON (`application/json`). Envoy's built-in OTel extensions (stats sink, access logger, tracer) only support protobuf encoding over HTTP. JSON encoding is not available in Envoy without a custom filter. So `protocol: http` means OTLP/HTTP with protobuf encoding. If JSON encoding support is needed later, a separate `encoding` field can be added to the spec.

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
    // OR reference to shared backend.
    // Uses TargetRef directly, not the routing BackendRef struct
    // (weight/port don't apply here).
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
| `HasStatus`     | true (surfaces unresolved backendRef conditions to referencing policies)             |
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

#### Configuration walkthroughs

Concrete configurations for each user story, all using Option A.

##### Story 1: Single collector, incremental rollout

Deploy the shared backend once, then add policies one signal at a time.

```yaml
# 1. Create the backend
apiVersion: kuma.io/v1alpha1
kind: MeshOpenTelemetryBackend
metadata:
  name: main-collector
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
spec:
  endpoint:
    address: otel-collector.observability
    port: 4317
  protocol: grpc
---
# 2. Start with metrics
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
        refreshInterval: 30s
---
# 3. Once metrics are verified, add tracing
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
      overall: 80
---
# 4. Then access logging
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
        attributes:
        - key: mesh
          value: "%KUMA_MESH%"
```

##### Story 2: Update collector address in one place

The observability team moves the collector to a new namespace. Update only the backend resource - all three policies pick up the change automatically.

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshOpenTelemetryBackend
metadata:
  name: main-collector
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
spec:
  endpoint:
    address: otel-collector.new-observability  # changed from otel-collector.observability
    port: 4317
  protocol: grpc
# MeshMetric, MeshTrace, and MeshAccessLog are unchanged.
```

##### Story 3: Managed backend with connection settings

Point at Grafana Cloud's OTLP gateway over HTTP. Initial scope covers endpoint and protocol. TLS and auth fields are follow-up work.

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshOpenTelemetryBackend
metadata:
  name: grafana-cloud
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
spec:
  endpoint:
    address: otlp-gateway-prod-us-east-0.grafana.net
    port: 443
    path: /otlp    # CP appends /v1/metrics, /v1/traces, /v1/logs per signal
  protocol: http
  # Follow-up work:
  # tls:
  #   mode: STRICT
  # auth:
  #   type: Bearer
  #   secretRef:
  #     name: grafana-cloud-token
---
# Policies reference it the same way as story 1
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
          name: grafana-cloud
        refreshInterval: 30s
```

##### Story 4: Per-zone collectors in multi-zone

When all zones run the same collector service name, create one backend on the Global CP. DNS resolves to the local collector in each zone.

```yaml
# Global CP - syncs to all zones via KDS
apiVersion: kuma.io/v1alpha1
kind: MeshOpenTelemetryBackend
metadata:
  name: zone-collector
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
spec:
  endpoint:
    address: otel-collector.observability.svc.cluster.local
    port: 4317
  protocol: grpc
# Policies also created on Global CP, synced to all zones.
# Each zone's DNS resolves the address to its local collector.
```

When zones need different collector addresses (separate cloud regions, different infrastructure), use zone-specific backend names:

```yaml
# Global CP
apiVersion: kuma.io/v1alpha1
kind: MeshOpenTelemetryBackend
metadata:
  name: collector-us-east
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
spec:
  endpoint:
    address: collector.us-east.internal
    port: 4317
  protocol: grpc
---
apiVersion: kuma.io/v1alpha1
kind: MeshOpenTelemetryBackend
metadata:
  name: collector-eu-west
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
spec:
  endpoint:
    address: collector.eu-west.internal
    port: 4317
  protocol: grpc
---
# Zone-targeted policy
apiVersion: kuma.io/v1alpha1
kind: MeshMetric
metadata:
  name: metrics-us-east
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
spec:
  targetRef:
    kind: MeshSubset
    tags:
      kuma.io/zone: us-east
  default:
    backends:
    - type: OpenTelemetry
      openTelemetry:
        backendRef:
          kind: MeshOpenTelemetryBackend
          name: collector-us-east
        refreshInterval: 30s
# Without MOTB, each zone's three policies would hardcode the endpoint.
# With MOTB, changing a zone's collector means updating one resource
# instead of three.
```

##### Story 5: Dangling reference detection

An operator deletes a backend that policies still reference. The CP detects the broken reference during xDS generation.

```yaml
# The backend was deleted. The policy still references it:
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
          name: main-collector    # no longer exists
        refreshInterval: 30s
```

CP behavior when the referenced backend is missing:

- Logs at Info level: `MeshOpenTelemetryBackend "main-collector" not found, referenced by MeshMetric "all-metrics"`
- Skips OTel backend config in xDS for that signal (no telemetry export)
- Status condition on the resource surfaces the unresolved reference so the operator can detect it via `kubectl` or the REST API

Unlike routing, where a dropped backend is one of many weighted destinations, a dropped telemetry backend means the signal is entirely lost. This is why `HasStatus: true` matters here.

##### Story 6: Env var auto-detection

The OTel Operator or a pod annotation sets `OTEL_EXPORTER_OTLP_*` env vars on the sidecar. kuma-dp reads them at startup and sends them to the CP via DynamicMetadata. Policies without an explicit endpoint or backendRef use them as fallback.

```yaml
# Pod annotation injects env vars into kuma-sidecar
apiVersion: v1
kind: Pod
metadata:
  annotations:
    kuma.io/sidecar-env-vars: >-
      OTEL_EXPORTER_OTLP_ENDPOINT=http://collector.observability:4317;
      OTEL_EXPORTER_OTLP_PROTOCOL=grpc
spec:
  containers:
  - name: my-app
    image: my-app:latest
---
# Policy with no endpoint and no backendRef
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
        # No backendRef, no endpoint.
        # kuma-dp reads OTEL_EXPORTER_OTLP_ENDPOINT from its env,
        # sends it to the CP via DynamicMetadata in the bootstrap
        # request. The CP uses it as the fallback endpoint.
        refreshInterval: 30s
# Priority: backendRef > inline endpoint > env var auto-detection
```

#### Naming

`MeshOpenTelemetryBackend` names the resource after its backend type. Since the resource can only be referenced from the OTel backend section of policies, a generic name adds no value. Type matching is enforced by the admission webhook - the validator rejects `backendRef.kind` values that don't match the enclosing backend type. If Zipkin or Datadog backend types are needed later, they become separate resources (`MeshZipkinBackend`, etc.) with their own `backendRef.kind`.

Rejected alternatives:
- `MeshTelemetryBackend` (with `type` discriminator) - a generic name + type discriminator can't enforce type matching via the API structure. A policy's OTel backend could reference a Zipkin-typed backend, and this mismatch would only be caught by runtime validation.
- `MeshOTelBackend` - abbreviation is less readable than the full name
- `MeshOTelCollector` - "collector" is one deployment model, the resource is for any telemetry-receiving endpoint
- `MeshObservabilityBackend` - too abstract, doesn't add clarity over the specific protocol name
- `MeshCollectorBackend` - not all backends are "collectors" (e.g., managed SaaS endpoints)

#### Signal-specific fields stay in policies

MeshAccessLog's `body` field is logs-only (OTel LogRecord body). It holds Envoy access log command operators (`%START_TIME%`, `%UPSTREAM_HOST%`) that control what each log record says. This is per-record content formatting, not connection config - it stays in MeshAccessLog.

MeshAccessLog's `attributes` are also logs-only (mapped to OTel `KeyValueList` with command operator substitution). MeshTrace has `Tags` (mapped to Envoy `CustomTag` on the HCM tracing config). MeshMetric has no OTel attributes (uses kuma-dp `ExtraLabels`). Each signal uses a completely different Envoy mechanism for attributes, so there's no shared abstraction that works across all three.

Shared resource attributes (mesh name, zone) in MeshOpenTelemetryBackend are conceptually sound but practically complex - Envoy handles resource attributes differently per signal (direct `KeyValueList` on access logger, `resource_detectors` extension on tracer, kuma-dp labels for metrics). This is follow-up work.

#### OTel environment variable auto-detection

The OTel specification defines environment variables (`OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_EXPORTER_OTLP_PROTOCOL`, `OTEL_EXPORTER_OTLP_HEADERS`, etc.) that configure where an OTel SDK sends telemetry. The OpenTelemetry Operator for Kubernetes injects these env vars into the first container in the pod spec by default (`Containers[0]`), which is typically the application container because mesh sidecar injectors append their containers. The Operator has no sidecar awareness - this targeting is a side effect of container ordering, not intentional filtering.

These env vars configure the application's OTel SDK, not the sidecar proxy. Envoy doesn't use the OTel SDK and doesn't read `OTEL_EXPORTER_OTLP_*` env vars - its telemetry export is configured through xDS pushed by the control plane. No service mesh today (Istio, Linkerd, or Kuma) reads pod env vars to configure sidecar proxy telemetry. This is an architectural split, not an oversight:

|                  | Application telemetry           | Proxy telemetry                                 |
|------------------|---------------------------------|-------------------------------------------------|
| Config source    | OTel env vars on app container  | Control plane via xDS                           |
| Who exports      | OTel SDK in app code            | Envoy sidecar                                   |
| What it captures | App-level spans, custom metrics | L4/L7 request metrics, access logs, mesh traces |

A mesh operator might want both paths pointing at the same collector without configuring each separately. This is user story 6, and the implementation is straightforward because the transport already exists.

##### DynamicMetadata transport

kuma-dp has a `DynamicMetadata` field (`map[string]string`) in the bootstrap request. During startup, kuma-dp collects key-value pairs and sends them to the CP via `POST /bootstrap`. The CP embeds them in Envoy's `node.metadata.dynamicMetadata`, which Envoy includes in every xDS discovery request back to the CP. Policy plugins can then read these values when generating xDS config.

Today this transport only carries the CoreDNS version. Adding `OTEL_*` env vars means reading them during kuma-dp startup and adding them to the same map. The rest of the pipeline handles arbitrary metadata already.

##### Getting env vars into kuma-dp

On Kubernetes, the sidecar injector does not copy env vars from the application container. Three paths exist today without code changes:

1. Pod annotation `kuma.io/sidecar-env-vars`: semicolon-separated key=value pairs added to the kuma-sidecar container. Example: `OTEL_EXPORTER_OTLP_ENDPOINT=http://collector:4317;OTEL_EXPORTER_OTLP_PROTOCOL=grpc`
2. `ContainerPatch` CRD: JSON patch that adds env vars to the sidecar container, applied via `kuma.io/container-patches` annotation.
3. Helm `containerConfig.envVars` on the sidecar injector config.

Auto-copying `OTEL_*` vars from the app container is possible but depends on container ordering and creates an implicit coupling between app and sidecar config. The annotation or ContainerPatch paths are explicit and less surprising. Having kuma-dp read its own process env vars is the cleaner approach since it works the same way on both Kubernetes and Universal.

On Universal, set `OTEL_*` env vars in the process environment (systemd unit, Docker run, shell) before starting kuma-dp. kuma-dp only processes `KUMA_*` vars for its own config, so `OTEL_*` vars sit in the environment until the startup code reads them.

##### Priority and override semantics

Explicit config wins. If a policy references a MeshOpenTelemetryBackend via `backendRef`, that's the endpoint - env vars are ignored. This matches how the OTel SDK itself works: programmatic config overrides env vars.

The inline `endpoint` field (deprecated) also takes precedence over env vars. Priority order: `backendRef` > inline `endpoint` > env var auto-detection. Operators can adopt env vars as a default and override specific policies with explicit config when needed.

##### Which env vars to capture

Supported:

- `OTEL_EXPORTER_OTLP_ENDPOINT` - where to send telemetry
- `OTEL_EXPORTER_OTLP_PROTOCOL` - grpc or http/protobuf
- `OTEL_EXPORTER_OTLP_COMPRESSION` - gzip or none
- `OTEL_EXPORTER_OTLP_TIMEOUT` - export timeout in milliseconds

Signal-specific endpoint vars (`OTEL_EXPORTER_OTLP_TRACES_ENDPOINT`, `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT`, `OTEL_EXPORTER_OTLP_LOGS_ENDPOINT`) are deferred - MeshOpenTelemetryBackend already handles per-signal routing via path suffixes.

`OTEL_EXPORTER_OTLP_HEADERS` is deferred until auth support lands (user story 3) because it may contain bearer tokens and needs redaction work - see security section below.

##### Security considerations for env var forwarding

`OTEL_EXPORTER_OTLP_HEADERS` may contain bearer tokens. The bootstrap request travels over HTTPS (same trust boundary as the DP token), so transport is encrypted. But the values end up in Envoy's `node.metadata`, which shows up in config dumps and could leak into CP debug logs. Sensitive header values should be redacted from config dump output and CP log messages, similar to how DP tokens are handled today.

### Option B: Named backends on the Mesh resource (rejected)

Add OTel collector configurations to the Mesh spec. Policies reference by name.

This goes against the current direction. We're actively moving config OUT of Mesh (MeshMetric replaced `Mesh.spec.metrics`, MeshTrace replaced `Mesh.spec.tracing`, MeshAccessLog replaced `Mesh.spec.logging`). Adding `metrics.backends` to Mesh was a mistake - the whole point of MeshMetric was to move this OUT of Mesh. We should not repeat it.

The Mesh resource is already large (networking, mtls, metrics, tracing, logging, routing, meshServices). Adding collector config there means no separate RBAC, no per-team ownership, and proto changes for every new field. Faster to implement but wrong direction.

### Option C: Reference MeshExternalService (rejected)

Model the OTel collector as a MeshExternalService. Policies reference it as a destination.

Semantically wrong - an OTel collector is infrastructure config, not an external service in the traffic routing sense. MeshExternalService is built for traffic routing (VIP allocation, passthrough) and doesn't capture OTel-specific semantics (protocol preference, path for HTTP endpoints, signal-specific routing). An OTel collector has multiple ports (4317 gRPC, 4318 HTTP) but MeshExternalService.match.port is singular. If the collector isn't meshed (recommended), the MeshExternalService creates routing expectations that conflict with direct connectivity.

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

- Dangling references: If a MeshOpenTelemetryBackend is deleted while policies reference it, policies lose their backend config. Unlike routing where a dropped backend is one of many weighted destinations, a dropped telemetry backend means the signal is entirely lost. The CP logs at Info level when a referenced backend is not found during xDS generation. With `HasStatus: true`, the resource surfaces unresolved backendRef conditions so operators can detect missing backends (user story 5) rather than silently losing telemetry. No cross-reference validation webhooks - Kuma's existing pattern is to not block deletion of referenced resources.
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

The resource is named after its backend type (`MeshOpenTelemetryBackend`) rather than using a generic name with a type discriminator. If other backend types are needed later, they become separate resources - and `backendRef.kind` enforces type matching via webhook validation.

Start with a minimal spec (endpoint only, no TLS/auth). Add TLS and auth fields as follow-up work. HTTP endpoint support will only exist in MeshOpenTelemetryBackend - the unreleased HTTP support in MeshTrace is being removed, and it won't be added to MeshMetric or MeshAccessLog.

Signal-specific config (refreshInterval, attributes, body, sampling) stays in each policy. The shared resource only handles "where is the backend and how do I connect to it."

### Scope

The first implementation covers:

1. MeshOpenTelemetryBackend resource with `endpoint` (address + port + path), `protocol` (grpc/http), and `HasStatus: true`
2. `backendRef` field added to all three policy OTel backends
3. Validation: `endpoint` XOR `backendRef` (mutual exclusivity)
4. Resolution in each policy's Apply()
5. Status conditions on the resource to surface unresolved backendRefs (user story 5)
6. Inline `endpoint` remains supported but deprecated (removed in 3.0)
7. `OTEL_EXPORTER_OTLP_*` env var auto-detection via DynamicMetadata transport (user story 6)

TLS and auth fields are follow-up work.

## Notes

The control plane's own OTel metric export (via `KUMA_TRACING_OPENTELEMETRY_ENABLED` and standard `OTEL_EXPORTER_OTLP_*` env vars) is separate. MeshOpenTelemetryBackend is for data plane telemetry policies only. CP observability config is environment-level, not mesh-scoped.
