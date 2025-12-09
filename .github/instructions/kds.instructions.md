---
applyTo:
  - "pkg/kds/**"
---

# KDS (Kuma Discovery Service) Guidelines

## Multi-Zone Architecture

**KDS:** Kuma Discovery Service for multi-zone synchronization

**Flow:** Global CP → Zone CPs via KDS
- Zone Ingress/Egress for cross-zone traffic
- K8s + Universal (VM) support
- Eventual consistency model

## Development Considerations

**Consistency:**
- Eventual consistency between zones
- Handle sync failures gracefully
- Resource versioning critical
- Conflict resolution strategy

**Cross-Zone Isolation:**
- Validate zone boundaries
- Resource isolation per zone
- `kuma.io/zone` tag enforcement
- No accidental cross-zone leaks

**Version Compatibility:**
- Backward/forward compatibility
- Protocol version handling
- Graceful degradation
- Migration paths

**Failure Handling:**
- Network partitions
- Temporary disconnections
- Sync retry logic
- State reconciliation

## Testing Requirements

**Multi-zone scenarios:**
- Global CP + multiple Zone CPs
- Zone Ingress/Egress traffic
- Cross-zone communication
- K8s + Universal mixed environments

**Failure cases:**
- Network failures
- Sync interruptions
- Version mismatches
- Resource conflicts

## Review Focus

- Consistency guarantees maintained
- Cross-zone isolation validated
- Version compatibility preserved
- Failure scenarios handled
- Resource mapping correct
