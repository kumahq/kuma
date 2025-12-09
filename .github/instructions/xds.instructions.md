---
applyTo:
  - "pkg/xds/**"
---

# XDS Development Guidelines

## XDS Overview

**Components:**
- **CDS** (Clusters) - Backend services
- **EDS** (Endpoints) - Service endpoints
- **LDS** (Listeners) - Inbound/outbound listeners
- **RDS** (Routes) - HTTP routes

## Performance Critical

XDS generation is **hot path** - performance matters:

- Minimize allocations (many dataplanes)
- Efficient ResourceSet usage
- Optimize cache (MeshContext dual-tier)
- No unnecessary goroutines
- Batch operations where possible

## Development Guidelines

**Configurers:**
- Strategy pattern for Envoy config
- Single responsibility
- Keep focused (one concern per configurer)
- Example: `TimeoutConfigurer`, `CircuitBreakerConfigurer`

**Caching:**
- MeshContext: dual-tier cache (short TTL + hash)
- Cache invalidation on policy changes
- Efficient cache key generation

**Protocol Handling:**
- HTTP, TCP, gRPC specific logic
- Protocol detection and routing
- Metadata for debugging

## Testing

- Golden files for Envoy config validation
- All protocol scenarios (HTTP, TCP, gRPC)
- Performance benchmarks for hot paths
- K8s + Universal mode compatibility

## Review Focus

- No allocations in hot paths
- Efficient ResourceSet usage
- Proper caching and invalidation
- All protocols handled
- Metadata included for debugging
