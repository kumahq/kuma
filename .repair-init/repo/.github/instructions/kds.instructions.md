---
applyTo:
  - "pkg/kds/**"
---

# KDS (Kuma Discovery Service) Guidelines

## Multi-Zone Architecture

**KDS:** Kuma Discovery Service for multi-zone synchronization

**Protocol:** Delta xDS (v3) with incremental sync
- Bidirectional gRPC streams: `GlobalToZoneSync`, `ZoneToGlobalSync`
- ACK/NACK flow for error handling
- Hash-based resource versioning (SHA-256)

**Flow:** Global CP ↔ Zone CPs via KDS
- Zone Ingress/Egress for cross-zone traffic
- K8s + Universal (VM) support
- Eventual consistency model

**Key Packages:**
- `v2/reconcile/` - Snapshot generation, diffing, filters, mappers
- `v2/server/` - Delta xDS server, event-based watchdog (experimental)
- `v2/client/` - Delta xDS client, ACK/NACK handling
- `v2/store/` - Resource synchronization to local store
- `mux/` - Connection multiplexing, health monitoring, version negotiation
- `context/` - Filters, mappers, feature flags

## Development Considerations

**Resource Processing Pipeline:**
```
Global → Zone:
  Filter (GlobalProvidedFilter) → Map (GlobalResourceMapper) → Validate → Store (ZoneSyncCallback)

Zone → Global:
  Filter (ZoneProvidedFilter) → Map (ZoneResourceMapper) → Validate → Store (GlobalSyncCallback)
```

**Key Patterns:**
- Composable mappers: `CompositeResourceMapper(mapper1, mapper2, ...)`
- Conditional mapping: `reconcile.If(condition, mapper)`
- Hash naming: `HashedName(mesh, name, zone, namespace)` for uniqueness
- Two-error return: `(fatalError, nackError)` - fatal cancels connection, NACK continues

**Consistency:**
- Eventual consistency between zones
- Incremental sync via delta xDS (only changed resources)
- Version maps track hash per resource
- Snapshot caching for unchanged resources

**Cross-Zone Isolation:**
- Hash-based naming prevents conflicts: `resource-name-a1b2c3d4`
- `kuma.io/zone` tag auto-added to all zone resources
- `kuma.io/origin` label: `global` vs `zone`
- Filter prevents syncing resources back to origin zone
- `IsLocallyOriginated()` check in filters

**Version Compatibility:**
- Feature flags in gRPC metadata: `kds.Features{hash-suffix, zone-ping-health, producer-policy-flow, ...}`
- Check with `features.HasFeature(kds.FeatureX)` before applying logic
- Graceful degradation when features absent
- KDS version negotiation: v2 (legacy) vs v3 (current)

**Failure Handling:**
- ResilientComponent: exponential backoff (5s base → 1m max)
- NACK backoff: 5s delay after NACK to prevent CP overload
- Zone health check: 5min interval, stale connection cleanup
- Store conflicts: 5 retries with jitter
- gRPC keepalive: 15s timeout

**Adding New Features:**
1. Define constant in `pkg/kds/features.go`
2. Send in client metadata: `kds.FeaturesMetadataKey`
3. Check in filters/mappers: `features.HasFeature(kds.FeatureNewThing)`
4. Ensure backward compatibility if feature absent
5. Document migration path

## Testing Requirements

**Unit Tests:**
```go
var _ = Describe("Feature", func() {
    var store core_store.ResourceStore
    var reconciler reconcile.Reconciler

    BeforeEach(func() {
        store = memory.NewStore()
        generator := reconcile.NewSnapshotGenerator(
            core_manager.NewResourceManager(store),
            myFilter,  // or reconcile.Any
            myMapper,  // or reconcile.NoopResourceMapper
        )
        reconciler = reconcile.NewReconciler(hasher, generator, cache, types)
    })

    It("should sync resource", func() {
        // given
        Expect(store.Create(..., resource, ...)).To(Succeed())

        // when
        err, changed := reconciler.Reconcile(ctx, node, changedTypes, log)

        // then
        Expect(err).ToNot(HaveOccurred())
        Expect(changed).To(BeTrue())
    })
})
```

**Multi-zone scenarios:**
- Global CP + multiple Zone CPs
- Zone Ingress/Egress traffic
- Cross-zone communication
- K8s + Universal mixed environments
- Feature flag combinations (with/without features)

**Failure cases:**
- Network failures (connection drops, retries)
- Sync interruptions (NACK, validation errors)
- Version mismatches (old Global CP, old Zone CP)
- Resource conflicts (hash collisions, duplicate names)
- Store transaction rollback

## Performance Considerations

**Optimization:**
- Event-based watchdog: reconcile only on resource changes (experimental, `EventBasedWatchdog=true`)
- Simple watchdog: timer-based reconciliation (default, `RefreshInterval=1s`)
- Version maps: skip unchanged resources in delta response
- Snapshot caching: reuse snapshots when no changes
- NACK backoff: prevent control plane overload

**Metrics:**
- `kds_delta_generation{reason, result}` - generation time
- `kds_delta_generation_errors` - failures
- `kds_resources_sync` - sync duration

## Review Focus

- Delta xDS protocol correct (ACK/NACK flow)
- Feature flag handling (backward compatibility)
- Resource filtering (zone boundaries, origin labels)
- Hash naming applied correctly (when `hash-suffix` feature present)
- Two-error pattern used (fatal vs NACK)
- Event-based watchdog integration (if applicable)
- Zone health check handling
- Store transaction safety (rollback on errors)
- Cross-zone isolation validated (no leaks)
- Testing covers feature flag combinations
