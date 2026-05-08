# Auto-generated MeshService on Universal post-inbound-tag removal

* Status: proposed

Technical Story: https://github.com/Kong/kong-mesh/issues/9443

## Context and Problem Statement

On Universal, the CP auto-generates `MeshService` from `Dataplane` inbound tags
(`pkg/core/resources/apis/meshservice/generate/generator.go`). Two in-flight
changes collide:

1. Remove auto-generation of `MeshService` on Universal.
2. Remove `Dataplane` inbound tags in favour of `MeshService` as primary identity.

A field report exposed two unmet needs:

- Custom `Dataplane.metadata.labels` and inbound tags do not propagate to the
  auto-generated `MeshService`. MMZS selects on `MeshService.metadata.labels`,
  so multi-zone selection by team/env is impossible for Universal today.
- Some operators (ECS/Fargate behind restricted networks) cannot reach the
  zone CP REST API. Their only channel is the `Dataplane` shipped via
  `kuma-dp run --dataplane-file`.

Inbound tags do two jobs: DP grouping (which DPs are one logical service) and
per-inbound service membership (which services this port serves). The
`kuma.io/workload` label and existing `Workload` generator solve the first
job. M:M cases (port carve-out: one workload to many MeshServices; workload
aggregation: blue/green, canary) need the second.

### Use cases

1. Trivial: N tasks for one service, one MeshService, custom labels reach `MeshService.metadata.labels`.
2. Port carve-out: one workload exposes `http`, `admin`, `metrics` as separate MeshServices.
3. Workload aggregation: two workloads (blue/green) back the same MeshService.
4. Restricted-network operator: declares everything via `Dataplane` template at `kuma-dp` startup.

## Decision Drivers

- Removing inbound tags must not regress M:M expressiveness.
- Restricted-network operators must declare intent and observe its effect through the `Dataplane`/`kuma-dp` channel only. Signals queryable solely via `kumactl inspect` are unreachable for them.
- Auto-generation has a single writer (CP). Distributed N-writer designs on a shared resource reintroduce CP-side coordination.
- Silent MMZS misses are unacceptable.
- Decision-blocking inputs: empirical M:M usage frequency (Options C vs D), and Workload-to-MeshService lifecycle ownership (cascading vs grace-period delete).

### Release timeline (hard)

- Kuma 2.14: tag-free operation supported on K8s and Universal, opt-in via `inboundTagsDisabled`. The structural replacement (C or D) ships here with a migration tool. The tactical label-propagation patch ships here.
- Kuma 3.0: tags removed by default. Downstream policies matching `kuma.io/service` break unless migrated.
- Exit criteria below are 2.14-cycle deliverables.

### Path-dependent constraints

- KDS replication: any auto-gen must round-trip zone to global cleanly.
- MMZS resolution: needs a deterministic label source-of-truth across reconciles.
- Unified resource naming: validator enforces `inbound.name` stability.

## Design

### Option A: `kuma-dp run --meshservice-template`

DP renders MeshService template and submits via bootstrap; CP upserts.

* Bad. MeshService is N:1 with Dataplane: cold-start race, reconnect flap, rolling-restart skew, last-task orphan, env-var template drift.
* Bad. The mitigations (idempotent first-write, ref counting, primary DP) reintroduce CP coordination.
* Bad. Broadens DP token to write a shared resource; security regression.

### Option B: operator-authored MeshService, no auto-generation

CP no longer generates. Operators run `kumactl apply -f meshservice.yaml`.

* Good. Deletes the polling generator, `checkMeshServicesConsistency`, grace-period delete, and ownership tracking. Smallest CP code.
* Good. No conflict resolution (single writer). No `WorkloadStatus.Conditions` to maintain.
* Good. Symmetric with Kubernetes (operator authors the service-identity resource).
* Good. No new fields on `Dataplane` spec.
* Bad. Blocks the restricted-network operator entirely.
* Bad. Universal users must run a parallel CI/GitOps surface for MeshService.
* Bad. The tactical label-propagation patch cannot ship under it.

### Option C: workload-only auto-generation

CP generates one MeshService per `kuma.io/workload` value; ports = union of all inbounds.

* Good. Covers the trivial case with no new spec fields.
* Bad. Cannot express M:M cases. Removing inbound tags silently drops port carve-out and workload aggregation.

### Option D: structured per-inbound `meshServices` field

Replace the free-form `inbound.tags["kuma.io/service"]` with a typed list:

```yaml
networking:
  inbound:
    - port: 8080
      name: http
      meshServices: [checkout]
    - port: 7000
      name: metrics
      meshServices: [checkout-metrics, observability]   # multi-valued, M:M
```

CP groups inbounds across the mesh by each value in `inbound.meshServices[*]`.
Per group: `name` is the membership value; `selector` is `kuma.io/workload`
label match plus per-port-name filter; `ports` is the union by `inbound.name`;
`labels` is the union of non-`kuma.io/*` keys from member Dataplanes.

Member DPs sorted by `(creation_time, name)` via `SortDataplanes`. On conflict:

| Conflict | Policy |
|---|---|
| Same port name, different port/protocol | First-DP-wins; `WorkloadStatus.Conditions[PortConflict]=True` |
| Same label key, different value | First-DP-wins; `WorkloadStatus.Conditions[LabelConflict]=True` |
| Divergent port set across DPs | Union; xDS endpoint filter drops DPs missing the port |
| Missing `kuma.io/workload` or `inbound.name` | Reject at registration |

Migration: `kuma.io/service: <name>` inbound tag becomes `inbound.meshServices: [<name>]`. Mechanical 1:1 lift.

* Good. Full M:M expressiveness; the multi-valued list fits port carve-out and aggregation.
* Good. The channel is the existing `Dataplane`; restricted-network operators are unblocked.
* Good. Typed and validated; typos fail at registration, not silently at MMZS.
* Good. Composes with the existing `kuma.io/workload` label and `Workload` generator.
* Good. The tactical label-propagation patch ships under it; the field report closes immediately.
* Bad. Concedes per-inbound service membership is load-bearing; that's a walk-back from "remove inbound tags entirely."
* Bad. Adds a hard-to-delete field on `Dataplane`. The polling generator and `inboundTagsDisabled` branching stay.
* Bad. `meshServices` (plural, on inbound) vs `MeshService` (resource) creates support confusion.
* Bad. First-DP-wins may not match operator intuition for blue/green (newest-wins).
* Bad. Largest implementation surface: validator, generator, condition lifecycle, xDS filter, migration tool, downstream policy audits.

### Tactical patch (independent, ships in 2.14)

Generator propagates non-`kuma.io/*` keys from `Dataplane.metadata.labels` to
`MeshService.metadata.labels`, first-DP-wins. Closes the field report. Inbound
tags do not propagate, which signals deprecation. Once shipped, MMZS selectors
in the wild depend on it; Option B becomes materially harder.

### Migration window behavior

A fleet in transition carries both forms. `checkMeshServicesConsistency`
oscillates each tick under split fleets. The chosen option must enforce one
of:

- Validator: a Dataplane MUST NOT carry both `inbound.tags["kuma.io/service"]` and the new field.
- Per-mesh flag: `inboundTagsDisabled` flips per-mesh, exactly one generator path per mesh.

### Alternatives not considered

- Event-driven generator off `DataplaneLifecycle`: revisit if the performance budget is exceeded.
- MeshService-as-view (not stock): foreclosed by the current KDS/MMZS replication contract.

## Security implications and review

- DP token scope unchanged; CP remains sole `MeshService` writer.
- No new privileged DP-to-CP channel. Option A's auth-broadening is avoided.
- `kuma.io/workload` and `inbound.meshServices` are validated at Dataplane registration.

## Reliability implications

- Silent MMZS-miss is the dominant failure. CP emits a structured warning and
  metric whenever an MMZS resolves to zero MeshServices in a zone. Lands with
  the tactical patch.
- `WorkloadStatus.Conditions[PortConflict|LabelConflict]` are set and cleared
  on every reconcile pass; stale `True` values are unacceptable and must be
  tested.
- Conflict signals must mirror to `DataplaneInsight` so `kuma-dp` logs surface
  them locally for restricted-network operators.
- `inboundTagsDisabled` is opt-in in 2.14 and default in 3.0; no follow-up
  MADR. Migration is forward-only.

### Required failure-mode handling

| Failure | Required behavior |
|---|---|
| Cross-operator `kuma.io/workload` collision | Validator checks uniqueness within mesh, or document accepted merge |
| Auto-gen vs operator-authored MeshService same name | Operator-authored wins via `ManagedByLabel`; generator MUST NOT overwrite |
| Operator removes `kuma.io/workload` from running DP | Treat as workload departure; grace-period delete only when no DP claims it |
| Conflict resolved between reconciles | Set conditions to `False` explicitly on clean pass |

### Performance budget

Polling cost at scale is unmeasured. Publish p50/p99 of
`component_meshservice_generator` at 1k, 5k, and 20k Dataplanes before the
structural option lands. Option D adds an O(M) factor (mean `meshServices[]`
cardinality) on top of the existing O(D·I).

## Implications for Kong Mesh

Significant in 3.0. Every downstream policy matching on `kuma.io/service`
inbound tags breaks at upgrade unless migrated. The downstream project must
audit policies, run the migration tool, and document the 2.14-to-3.0 upgrade.

## Decision

TBD. The tactical patch and MMZS observability ship in 2.14 ahead of the structural decision.

## Exit criteria (close inside Kuma 2.14)

1. M:M telemetry: count unique `kuma.io/service` values per workload-grouped DP set. Threshold of >=X% with >1 service/workload picks Option D; below picks Option C.
2. Performance run at 1k, 5k, and 20k Dataplanes.
3. Operator interviews: 3-5 ECS/Fargate restricted-network operators, ACTA-style.
4. Lifecycle commitment: Workload-owns-MeshService vs generator-owns-MeshService.
5. Migration tool ships in 2.14 alongside opt-in.
6. Conflict policy locked: first-DP-wins vs newest-wins vs explicit priority, decided on data, not aesthetics.

## Notes

- Open: precise selector encoding for "workload labels union per-port filter". May live in xDS endpoint generator; confirm against `pkg/xds/generator/endpoint*`.
- Open: field naming to avoid `meshServices` (field) vs `MeshService` (resource) confusion.
- Open: whether to widen the bootstrap channel to accept operator-authored MeshService for the restricted-network cohort, narrowing the gap between Options B and D.
