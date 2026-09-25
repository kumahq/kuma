# Incremental, event-driven MeshContext and proxy-scoped regeneration

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/3685, https://github.com/kumahq/kuma/issues/3684, https://github.com/kumahq/kuma/issues/13273

## Context and Problem Statement

`MeshContext` (`pkg/xds/context`) holds everything a zone control plane needs, for one mesh, to generate xDS for any proxy in that mesh: resource lists, the destination index, VIP outbounds and domains, endpoint maps, CA bundles, zone egress instances.
It is produced by `meshContextBuilder.BuildIfChanged` and shared by every proxy of the mesh through `pkg/xds/cache/mesh`.

How it is used today:

1. Every connected proxy has a `DataplaneWatchdog` goroutine that ticks every `dataplaneConfigurationRefreshInterval` (1s).
2. Each tick calls `MeshCache.GetMeshContext`. The mesh cache has a 1s TTL, so about once per second per mesh the builder:
   * lists roughly 30 resource types from the read-only (1s TTL) cached manager
   * resolves DNS for every Dataplane and MeshZoneAddress
   * hashes every resource (a deterministic proto marshal of every Dataplane, a JSON marshal of every MeshService/MeshExternalService/MeshMultiZoneService)
   * builds a new `DestinationIndex`
   * only then compares the hash with the previous context and throws the work away when nothing changed
3. When the hash differs, the builder recomputes everything derived from the lists (endpoint maps, VIPs, CAs, zone egresses), even if only one type changed.
4. The watchdog regenerates the proxy's whole config when `MeshContext.Hash` differs from the last one it saw. That hash covers the Dataplane roster (full spec, including inbound health), secrets, every policy, and every `Mesh` resource on the CP.

As a result:

* **Fan-out.** One Dataplane joining, leaving, or flipping inbound health makes every proxy in the mesh run a full `Build`, `GenerateSnapshot` and snapshot marshal. Benchmarks on a 500-DP mesh with 100 flapping DPs measured 130–220 full regenerations per second (`project_policy_cache_17017` benchmark), and mesh rebuild (`List` + unmarshal + hash) at 14–21% of CP CPU.
* **Cross-mesh coupling.** A version bump on any `Mesh` (annotations, status on Kubernetes) invalidates every mesh on the CP.
* **Polling.** The unchanged path is paid every second per mesh, whether or not anything changed. Cost is proportional to the size of the mesh, not to the rate of change.
* **Goroutine per proxy.** With thousands of proxies, a mesh-wide change wakes thousands of goroutines that all allocate a full snapshot at the same time, causing CPU and memory spikes (issue #3684).
* **Per-proxy duplication.** Outbound rules for a destination are recomputed per proxy even when every client of a destination ends up with the same producer config (issue #13273).

A first round of work (quick wins, and splitting `MeshContext` into parts with their own hashes that are rebuilt independently) makes each rebuild cheaper, but keeps the polling model and the mesh-wide regeneration gate.
This MADR covers the changes to the model itself.

### Use cases

1. A Deployment with 50 replicas rolls out in a 2000-proxy mesh. Only clients of that service should do any work, and only for EDS.
2. A user edits one MeshTimeout targeting service `payments`. Only proxies that match that policy should regenerate.
3. Nothing changes for an hour. The CP should do close to zero work for that mesh.
4. A Mesh resource in mesh `a` is annotated. Proxies in mesh `b` should not notice.
5. 5000 proxies connect at once after a CP restart. Memory should stay bounded.

## Design

The design builds on the split `MeshContext` from the first round:

| Part | Contents | Rebuilt when |
|---|---|---|
| MeshConfig | the mesh's own `Mesh` resource | the Mesh changes |
| PolicyContext | policies, MeshService/MeshExternalService/MeshMultiZoneService, `DestinationIndex`, VIP outbounds and domains | a policy or destination changes |
| TrustContext | MeshIdentity, MeshTrust, Secrets, CAs by trust domain, zone egress SANs | identity or trust material changes |
| TopologyContext | Dataplanes, MeshZoneAddresses, endpoint maps (outbound, zone ingress, zone egress) | a Dataplane, MeshZoneAddress or endpoint-relevant destination changes |

There are four independent parts, C1–C4. Each can ship on its own; C2 benefits the most from C1.

### C1. Event-driven invalidation instead of polling

Subscribe the mesh context component to the resource event bus (`pkg/events`), which KDS already uses for its event-based watchdog.
Each event marks one part of one mesh dirty, based on the resource type:

| Resource type | Marks dirty |
|---|---|
| `Mesh` | MeshConfig of that mesh only |
| policies, MeshService, MeshExternalService, MeshMultiZoneService | PolicyContext (and TopologyContext for destinations, since endpoint maps depend on them) |
| MeshIdentity, MeshTrust, Secret | TrustContext (and TopologyContext for Secrets used by MeshExternalService TLS) |
| Dataplane, MeshZoneAddress | TopologyContext |

A per-mesh loop flushes dirty parts on a short debounce (for example 100ms–1s), rebuilds only those parts, and publishes a new immutable `MeshContext` version.
A periodic full resync (for example every 1m, with jitter) protects against missed events.
DNS-resolved addresses (Dataplane and MeshZoneAddress hostnames) are re-resolved on their own timer, and mark TopologyContext dirty only when the resolved IP changes.

Watchdogs stop polling the mesh cache and receive a notification with the new version instead (see C3).

Pros:
* No work at all for a mesh while nothing changes (use case 3).
* Work is proportional to the rate of change, not to mesh size.
* The same mechanism, and its lessons, already exist in KDS: flush ticks aligned to interval boundaries, jitter below 2s, carrying unchanged types over without rebuilding them.

Cons:
* The event bus is best-effort. A dropped event means stale config until the next full resync. KDS has run a 1m resync with `events_dropped` at 0 in production.
* On Kubernetes, events come from informers and on Universal from Postgres `LISTEN/NOTIFY`. Both already feed the event bus, but every resource type in `MeshResourceTypes()` needs to be covered. A test should enforce this.
* More moving parts than a TTL cache: a debouncer per mesh, lifecycle handling for deleted meshes.

Alternative considered: keep polling, but make the unchanged path cheaper by memoizing per-resource hashes by `(type, mesh, name, version)`.
This is simpler and a reasonable interim step, but it still lists every type every second and still needs special handling for DNS-resolved objects, whose content changes without a version bump.

### C2. Proxy-scoped regeneration

Replace the single mesh-wide gate (`meshCtx.Hash != lastHash`) with a per-proxy decision based on what the proxy actually depends on.

After each full generation, a proxy records its **dependency set**:
* the hashes of the MeshConfig, PolicyContext and TrustContext parts
* its own Dataplane `XDSHash`
* the set of endpoint keys it read (the services behind its generated clusters)
* whether it read the whole topology (direct access, embedded zone ingress/egress listeners)

TopologyContext keeps a hash per endpoint key, so a topology rebuild produces the set of **changed keys**.

On a new `MeshContext` version, each proxy does one of:

1. **Skip.** Only topology changed and none of the changed keys are in the proxy's dependency set.
   With reachable backends configured, this is the common case during rollouts (use case 1 for non-clients).
2. **EDS-only.** Only topology changed and some dependent keys changed.
   Regenerate only the `ClusterLoadAssignment`s for those keys, replace them in the previous snapshot, and re-version.
   This avoids policy matching, listener, route and cluster generation, and marshaling of unchanged types.
3. **Full.** A config part changed, the proxy's own Dataplane changed, or the proxy reads the whole topology.
   This is the path everything takes today.

The EDS-only path is not a plain copy of `EndpointMap` into CLAs. Two pieces of per-proxy logic post-process CLAs today, and would have to be replayed:
* `meshloadbalancingstrategy` rewrites CLAs through `claConfigurer`: locality priorities, affinity to pod labels, overprovisioning factor. The input is the proxy's matched MeshLoadBalancingStrategy conf, which does not change on a topology-only update. It can be stored with the dependency set and replayed.
* `pkg/xds/generator/zoneproxy` builds CLAs for embedded zone proxies. These proxies are marked as reading the whole topology and always take the full path.

The EDS-only path needs its own generator entry point, and a test that runs both the EDS-only and full paths over the same inputs and requires byte-identical EDS output. That test is the main guard against the two paths drifting apart.

Recording which endpoint keys were read needs accessors instead of direct `EndpointMap` map reads, for example `TopologyContext.Endpoints(key)`, which records the key in a per-generation tracker.

Pros:
* Scales with the number of affected proxies instead of mesh size (use cases 1 and 2).
* The skip path is free.
* Policy-only changes still take the full path, so the risk is limited to topology-only updates.

Cons:
* Two generation paths for EDS that have to stay consistent.
* The dependency tracking has to see every endpoint read. A missed read means stale endpoints. Mitigated by the equivalence test and a periodic forced full regeneration (for example every 5m per proxy, with jitter).
* Without reachable backends every proxy depends on every destination, so option 1 never applies and option 2 is the whole gain.

Alternative considered: narrowing only the watchdog gate to a "config hash" plus the proxy's own Dataplane hash, and ignoring topology.
This is incorrect, because endpoints must still be pushed.

### C3. Worker pool instead of a goroutine per proxy

Keep one lightweight registration per connected proxy (key, metadata, last dependency set, last snapshot), but move generation to a bounded pool of workers (default `GOMAXPROCS`):

* A new `MeshContext` version enqueues the affected proxies with the decision from C2 (skip proxies are never enqueued).
* The queue deduplicates per proxy and always processes it against the newest version, so a burst of changes collapses into one generation per proxy.
* Identity rotation (`ExpiringSoon`) and connect or reconnect of a proxy enqueue that proxy only.
* The queue is ordered by priority: new connections first (they have no config), then full, then EDS-only.

Pros:
* Bounded memory under storms (use case 5). A snapshot is only allocated for the proxies being worked on.
* Better CPU use: no thundering herd of thousands of goroutines.
* Replaces the per-tick work that happens today even when nothing changed (`SelectedIdentity` plus a JSON hash per proxy per second).

Cons:
* Changes how latency is spread. With a queue, some proxies get config later than others. Queue depth and time-in-queue have to be exposed as metrics.
* `DataplaneWatchdog` also drives inbound health updates, OTel status and metrics. These have to move.

### C4. Shared producer outbound rules

Precompute, as part of PolicyContext, the outbound configuration contributed by **producer** policies for each destination. These are policies whose top-level `targetRef` selects the destination's owner rather than the client.
Per proxy, only **consumer** policies (those that select the client) are matched and merged on top.

This follows issue #13273. The data structure maps `(policy type, destination)` to a list of confs.
A proxy merges the shared producer list with its own consumer matches, so memory scales with #destinations × #policy types rather than #proxies × #destinations.

Pros:
* Removes the largest remaining per-proxy cost when there are many policies. At 2000 MeshTrafficPermissions, `MatchedPolicies`/`BuildFromRules` was measured at about 46% of CP CPU.
* Pairs with C2: a policy change can tell which destinations it affects.

Cons:
* The merge order of producer vs consumer policies has to be exactly the same as the existing `BuildRules`. This needs golden-file equivalence tests over the whole policy suite.
* The biggest change to the policy plugin API of the four parts. Every plugin that uses `ToRules.ResourceRules` has to switch.

### Rollout order

1. C1 behind a flag (`KUMA_XDS_SERVER_EVENT_BASED_MESH_CONTEXT`), defaulting to off. Validate it against the polling path by hashing both and alerting on divergence in tests and e2e.
2. C3 in the same release as C1, since both change who drives generation.
3. C2: skip path first, then the EDS-only path with the equivalence test.
4. C4 independently, when policy counts make it worth doing.

## Security implications and review

No new API or input surface.
The risk to watch is stale security config.
All of the following always take the full regeneration path and never the skip or EDS-only paths:
* MeshTrafficPermission and other RBAC-affecting policies (PolicyContext)
* trust bundle and identity changes (TrustContext)
* a proxy's own Dataplane changes

A missed event (C1) delays such a change until the next full resync. The resync interval bounds how long a revoked permission can stay active, so it must stay short (≤ 1m) and be documented.

## Reliability implications

* **Staleness bounds:**
  * event debounce (sub-second)
  * the full resync interval (C1)
  * forced periodic full regeneration per proxy (C2)
  
  All three are configurable, with metrics.
* **New metrics:**
  * `mesh_context_build_seconds{part,result}`
  * `mesh_context_dirty_total{part,resource_type}`
  * `xds_regeneration_total{mode=skip|eds|full}`
  * `xds_generation_queue_depth`
  * `xds_generation_queue_wait_seconds`
* The polling path stays available behind a flag for at least one release as a fallback.
* A dropped event, a missed endpoint read, or a queue that falls behind all show up as stale config, not as crashes. The equivalence tests (C2, C4) and a divergence check between event-driven and polling contexts (C1) are what keep these from shipping.

## Implications for Kong Mesh

* The enterprise fork reads `MeshContext` fields directly. For example, the meshopa plugin uses `DataSourceLoader`. The split must keep fields that downstream code reads, or provide accessors on `MeshContext`.
* Enterprise policies that post-process `ClusterLoadAssignment`s must either register a replay hook for the EDS-only path (C2) or opt their proxies into the full path.
* Enterprise resource types that feed `MeshContext` must declare which part they belong to, so C1 can mark the right part dirty. An unknown type should default to marking every part dirty.

## Decision

Adopt C1 (event-driven, per-part invalidation with a periodic full resync) and C3 (a bounded worker pool) together, behind a flag, with the polling path kept as a fallback.
Follow with C2: first the skip path, then the EDS-only path, guarded by a test that requires EDS-only and full generation to produce identical EDS output.
C4 is accepted as the direction for per-proxy policy cost and will be planned separately.

## Notes

* The first round of work (fixing the reuse of the previous global and base contexts, hashing each list once, scoping the global context to the mesh's own `Mesh`, computing endpoint-map inputs once, adding build metrics, and splitting `MeshContext` into independently rebuilt parts) is a prerequisite and is tracked separately.
* Open questions:
  * Should the MeshService/Workload generators and MADS consume a thinner resource snapshot instead of the full `MeshContext`?
  * Should the store cache TTL (1s) and the mesh cache stay once watchdogs no longer poll?
  * Is 1m the right full resync interval for security-relevant config?
