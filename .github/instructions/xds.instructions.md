---
applyTo:
  - "pkg/xds/**"
---

# XDS Development Guidelines

## Architecture Overview

**Key Directories:**
```
pkg/xds/
├── cache/          # Caching system (once, mesh, cla)
├── context/        # MeshContext building
├── envoy/          # Envoy builders/configurers
│   ├── clusters/   # Cluster builders
│   ├── listeners/  # Listener builders
│   └── routes/     # Route builders
├── generator/      # xDS resource generators
└── server/         # xDS server (callbacks, v3)
```

**Data Flow:**
```
Server → Sync → Generator (Composite) → Builders → Configurers → Envoy Resources
```

**Core Patterns:**
- **Builder** - Fluent API for constructing Envoy resources
- **Configurer** - Strategy pattern for modifying Envoy config
- **Generator** - Creates xDS resources for dataplanes
- **Adapter** - Function-to-interface converters

## XDS Components

- **CDS** (Clusters) - Backend services
- **EDS** (Endpoints) - Service endpoints
- **LDS** (Listeners) - Inbound/outbound listeners
- **RDS** (Routes) - HTTP routes

## Performance Critical

XDS generation is **hot path** - performance matters:

- Minimize allocations (many dataplanes)
- Efficient ResourceSet usage
- Optimize cache (three-tier context hierarchy)
- No unnecessary goroutines
- Batch operations where possible
- Single-flight pattern prevents thundering herd

## Core Patterns

### Builder Pattern

**Location:** `pkg/xds/envoy/clusters/`, `pkg/xds/envoy/listeners/`

**Interfaces:**
```go
// Build with error handling
func (b *ListenerBuilder) Build() (envoy.NamedResource, error)

// Build with panic on error (tests only)
func (b *ClusterBuilder) MustBuild() envoy.NamedResource
```

**Usage:**
```go
cluster := envoy_clusters.NewClusterBuilder(apiVersion, name).
    Configure(envoy_clusters.ProvidedEndpointCluster(false, endpoint)).
    Configure(envoy_clusters.Timeout(timeout, protocol)).
    ConfigureIf(condition, envoy_clusters.Http2()).  // Conditional
    Build()

listener := envoy_listeners.NewListenerBuilder(apiVersion, name).
    Configure(envoy_listeners.InboundListener(ip, port, protocol)).
    Configure(envoy_listeners.TransparentProxying(proxy)).
    MustBuild()  // Tests only
```

### Configurer Pattern (Strategy)

**Location:** `pkg/xds/envoy/*/v3/*_configurer.go`, `pkg/plugins/policies/*/xds/`

**Interfaces:**
```go
type ListenerConfigurer interface {
    Configure(listener *envoy_listener.Listener) error
}

type ClusterConfigurer interface {
    Configure(cluster *envoy_cluster.Cluster) error
}

type FilterChainConfigurer interface {
    Configure(filterChain *envoy_listener.FilterChain) error
}
```

**Examples:**
- `pkg/xds/envoy/clusters/v3/timeout_configurer.go` - Connection/idle timeouts
- `pkg/xds/envoy/listeners/v3/tcp_proxy_configurer.go` - TCP proxy filter
- `pkg/plugins/policies/meshtimeout/plugin/xds/configurer.go` - Policy timeout

**Rules:**
- Single responsibility (one concern per configurer)
- Protocol-aware (switch on HTTP/TCP/gRPC)
- Return error, don't panic
- Composable via builder

### Adapter Pattern

**Location:** `pkg/xds/envoy/listeners/v3/configurer.go`

Convert functions to configurer interfaces:
```go
FilterChainConfigureFunc       // func → FilterChainConfigurer
FilterChainMustConfigureFunc   // non-failing variant
ListenerConfigureFunc          // func → ListenerConfigurer
ListenerMustConfigureFunc      // non-failing variant
ClusterMustConfigureFunc       // func → ClusterConfigurer
```

**Usage:**
```go
builder.Configure(v3.FilterChainConfigureFunc(func(chain *envoy_listener.FilterChain) error {
    // Custom logic
    return nil
}))
```

### Generator Pattern

**Location:** `pkg/xds/generator/`

**Interface:**
```go
type ResourceGenerator interface {
    Generate(ctx, *model.ResourceSet, xds_context.Context, *model.Proxy) (*model.ResourceSet, error)
}
```

**Implementations:**
- `pkg/xds/generator/inbound_proxy_generator.go`
- `pkg/xds/generator/outbound_proxy_generator.go`
- `pkg/xds/generator/admin_proxy_generator.go`
- `pkg/xds/generator/egress/generator.go`

**CompositeResourceGenerator:** Chains multiple generators, accumulates resources

## Caching System

### Three-Tier Context Hierarchy

**Location:** `pkg/xds/context/mesh_context_builder.go`, `pkg/xds/cache/mesh/`

**Hierarchy:**
```
GlobalContext (zone-scoped)
  ├── Meshes, ZoneIngress, ZoneEgress
  └── hash: FNV128a
      ↓
BaseMeshContext (mesh policies, changes less often)
  ├── Mesh, policies, gateways, external services, VIPs
  └── hash: FNV128a
      ↓
MeshContext (dataplanes, secrets, changes frequently)
  ├── Dataplanes, endpoints, secrets
  └── Hash: base64(globalHash + baseHash + resourcesHash)
```

**Hash-Based Invalidation:**
```go
// Rebuild only if hash changed
newHash := base64(fnv128a(globalContext.hash + baseMeshContext.hash + resourcesHash))
if latestMeshCtx != nil && newHash == latestMeshCtx.Hash {
    return latestMeshCtx  // Reuse existing
}
```

### Dual-Tier Cache

**Location:** `pkg/xds/cache/mesh/cache.go`, `pkg/xds/cache/once/`

**Layers:**
1. **Short TTL Cache** (`once.Cache`): ~1s, ignores changes during burst
2. **Long TTL Cache** (`hashCache`): ~1min, stores rebuilt contexts

**Single-Flight Pattern:** `once.Cache` ensures only one concurrent rebuild per key (prevents thundering herd)

**CLA Cache:** `pkg/xds/cache/cla/` - SHA256 key from `(apiVersion, meshName, clusterHash, meshHash)`

### Cache Key Generation

Efficient, deterministic:
- FNV128a for speed
- Include mesh name, API version, resource hashes
- SHA256 for CLA cache (content-addressable)

## Protocol Handling

**Supported Protocols:**
- HTTP, HTTP2, gRPC (L7 routing, HCM filters)
- TCP, Kafka (L4 routing, TCP proxy)
- ProtocolUnknown (fallback to TCP)

**Protocol-Specific Logic:**

**Timeouts:**
```go
switch protocol {
case ProtocolHTTP, ProtocolHTTP2, ProtocolGRPC:
    // Configure HCM: IdleTimeout, StreamIdleTimeout, MaxStreamDuration
case ProtocolUnknown, ProtocolTCP, ProtocolKafka:
    // Configure TCP Proxy: IdleTimeout
}
```

**Cluster Config:**
```go
switch protocol {
case ProtocolHTTP:
    builder.Configure(envoy_clusters.Http())
case ProtocolHTTP2, ProtocolGRPC:
    builder.Configure(envoy_clusters.Http2())
}
```

**Protocol Inference:** `pkg/xds/context/mesh_context_builder.go` - from endpoint tags (`kuma.io/protocol`) or external service config

**Metadata:** Include for debugging (service name, protocol, mesh, zone)

## Common Patterns

**Conditional Configuration:**
```go
builder.ConfigureIf(protocol == ProtocolHTTP2, envoy_clusters.Http2())
```

**Error Handling:**
```go
cluster, err := builder.Build()  // Production
cluster := builder.MustBuild()   // Tests only (panics)
```

**ResourceSet Usage:**
```go
rs := model.NewResourceSet()
rs.Add(&model.Resource{Name: "x", Resource: cluster})
rs.AddSet(otherSet)  // Efficient merging, deduplicates by type+name
```

**Adapter for Custom Logic:**
```go
v3.FilterChainMustConfigureFunc(func(chain *envoy_listener.FilterChain) {
    // Custom filter chain modification
})
```

## Testing

**Golden Files:** `pkg/xds/generator/testdata/`

```go
DescribeTable("Generate Envoy xDS resources",
    func(given testCase) {
        resources, err := generator.Generate(ctx, xdsCtx, proxy)
        actual, _ := util_proto.ToYAML(resources)
        Expect(actual).To(MatchGoldenYAML("testdata", "inbound-proxy", given.expected))
    },
    Entry("basic", testCase{dataplane: "01.yaml", expected: "01.golden.yaml"}),
    Entry("mtls", testCase{dataplane: "02.yaml", expected: "02.golden.yaml"}),
)
```

**Update Golden Files:**
```bash
UPDATE_GOLDEN_FILES=true make test
```

**Protocol Scenarios:**
- Test all protocols: HTTP, HTTP2, gRPC, TCP, Kafka
- K8s + Universal mode compatibility
- Multi-zone scenarios

**Cache Tests:**
- Concurrent access (single-flight verification)
- Hash stability
- Metrics (hit/miss/wait)

## Review Focus

**Performance:**
- No allocations in hot paths (builders, configurers)
- Efficient ResourceSet usage (AddSet, deduplication)
- Proper caching (MeshContext rebuild only on hash change)
- No unnecessary goroutines

**Correctness:**
- All protocols handled (HTTP/HTTP2/gRPC/TCP/Kafka)
- Metadata included for debugging
- Error handling (return errors, don't panic)
- Golden files updated

**Architecture:**
- Single responsibility configurers
- Builder pattern for composability
- Generator pattern for extensibility
- Adapter pattern for flexibility

## Quick Reference

**Key Interfaces:**
- `ListenerConfigurer` - `pkg/xds/envoy/listeners/v3/configurer.go:10`
- `ClusterConfigurer` - `pkg/xds/envoy/clusters/v3/configurer.go:10`
- `ResourceGenerator` - `pkg/xds/generator/core/resource_generator.go:15`
- `MeshContextBuilder` - `pkg/xds/context/mesh_context_builder.go:40`

**Common Configurers:**
- Timeout: `pkg/xds/envoy/clusters/v3/timeout_configurer.go`
- TLS: `pkg/xds/envoy/tls/v3/sni_configurer.go`
- TCP Proxy: `pkg/xds/envoy/listeners/v3/tcp_proxy_configurer.go`
- HTTP: `pkg/xds/envoy/listeners/v3/http_connection_manager_configurer.go`

**Generators:**
- Inbound: `pkg/xds/generator/inbound_proxy_generator.go`
- Outbound: `pkg/xds/generator/outbound_proxy_generator.go`
- Admin: `pkg/xds/generator/admin_proxy_generator.go`
- Egress: `pkg/xds/generator/egress/generator.go`
