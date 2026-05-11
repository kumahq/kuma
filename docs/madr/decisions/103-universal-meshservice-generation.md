# Auto-generated MeshService on Universal post-inbound-tag removal

* Status: accepted

Technical Story: https://github.com/Kong/kong-mesh/issues/9443

## Context and Problem Statement

On Universal, the control plane (CP) auto-generates `MeshService` from `Dataplane` inbound tags (`pkg/core/resources/apis/meshservice/generate/generator.go`). A field report exposed two unmet needs:

- Custom `Dataplane.metadata.labels` and inbound tags do not propagate to the auto-generated `MeshService`. MeshMultizoneService selects on `MeshService.metadata.labels`, so multi-zone selection by team/env is impossible for Universal today.
- Some operators (ECS/Fargate behind restricted networks) cannot reach the zone CP REST API. Their only channel is the `Dataplane` shipped via `kuma-dp run --dataplane-file`.

Inbound tags do two jobs: data plane (DP) grouping (which DPs are one logical service) and per-inbound service membership (which services this port serves). The `kuma.io/workload` label and existing `Workload` generator solve the first job. Many-to-many (M:M) cases (port carve-out: one workload to many MeshServices; workload aggregation: blue/green, canary) need the second.

### Use cases

1. Trivial: N tasks for one service, one MeshService, custom labels reach `MeshService.metadata.labels`.
2. Port carve-out: one workload exposes `http`, `admin`, `metrics` as separate MeshServices.
3. Restricted-network operator: declares everything via `Dataplane` template at `kuma-dp` startup.

## Decision Drivers

- Removing inbound tags must not regress M:M expressiveness.
- Restricted-network operators must declare intent and observe its effect through the `Dataplane`/`kuma-dp` channel only. Signals queryable solely via `kumactl inspect` are unreachable for them.
- Auto-generation has a single writer (CP). Distributed N-writer designs on a shared resource reintroduce CP-side coordination.
- Silent MeshMultizoneService misses are unacceptable.

### Release timeline

- Kuma 2.14: tag-free operation supported on Kubernetes (K8s) and Universal, opt-in via `inboundTagsDisabled`. The chosen path (Option D) ships here with a migration tool. The tactical label-propagation patch ships here.
- Kuma 3.0: tags removed by default. Downstream policies matching `kuma.io/service` break unless migrated.

## Design

### Option A: `kuma-dp run --meshservice-template`

DP renders MeshService template and submits via bootstrap; CP upserts.

* Bad. MeshService is N:1 with Dataplane: cold-start race, reconnect flap, rolling-restart skew, last-task orphan, env-var template drift.
* Bad. The mitigations (idempotent first-write, ref counting, primary DP) reintroduce CP coordination.
* Bad. Broadens DP token to write a shared resource; security regression.

### Option B: workload-only auto-generation

CP generates one MeshService per `kuma.io/workload` value; ports = union of all inbounds.

* Good. Covers the trivial case with no new spec fields.
* Bad. Cannot express M:M cases. Removing inbound tags silently drops port carve-out and workload aggregation.
* Bad. Blue/green deployments impossible.

### Option C: structured per-inbound `meshServices` field

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

CP groups inbounds across the mesh by each value in `inbound.meshServices[*]`. Per group: `name` is the membership value; `selector` is `kuma.io/workload` label match plus per-port-name filter; `ports` is the union by `inbound.name`; `labels` is the union of non-`kuma.io/*` keys from member Dataplanes.

Member DPs sorted by `(creation_time, name)` via `SortDataplanes`. On conflict:

| Conflict                                     | Policy                                                         |
|----------------------------------------------|----------------------------------------------------------------|
| Same port name, different port/protocol      | First-DP-wins; `WorkloadStatus.Conditions[PortConflict]=True`  |
| Same label key, different value              | First-DP-wins; `WorkloadStatus.Conditions[LabelConflict]=True` |
| Divergent port set across DPs                | Union; xDS endpoint filter drops DPs missing the port          |
| Missing `kuma.io/workload` or `inbound.name` | Reject at registration                                         |

Migration: `kuma.io/service: <name>` inbound tag becomes `inbound.meshServices: [<name>]`. Mechanical 1:1 lift.

* Good. Full M:M expressiveness; the multi-valued list fits port carve-out and aggregation.
* Good. The channel is the existing `Dataplane`; restricted-network operators are unblocked.
* Good. Typed and validated; a typo in the `meshServices` field is rejected at Dataplane registration, instead of silently producing a MeshMultizoneService that selects zero MeshServices later.
* Good. Composes with the existing `kuma.io/workload` label and `Workload` generator.
* Good. The tactical label-propagation patch (see "Tactical patch" section below) is compatible with this option, so the user-reported issue that initiated this MADR can be closed without waiting for the full structural change.
* Bad. Re-introduces per-inbound service membership under a new name. The original plan was to remove inbound tags entirely; this option keeps the same concept (which inbound serves which logical service) under a typed field instead of a free-form tag. We get a cleaner schema but admit the concept itself cannot be removed.
* Bad. Adds a hard-to-delete field on `Dataplane`. The polling generator and `inboundTagsDisabled` branching stay.
* Bad. First-DP-wins conflict policy is wrong for blue/green: when blue and green Dataplanes disagree on a label or port for the same service, the older (blue) one wins because it registered first, but operators usually expect the newer (green) deployment to take precedence.
* Bad. Largest implementation surface: validator, generator, condition lifecycle, xDS filter, migration tool, downstream policy audits.

#### Migration window behavior

During the migration to the new field, some Dataplanes in a mesh still use the old `kuma.io/service` inbound tag while others already use `inbound.meshServices`. If both forms reference the same service name, the existing `checkMeshServicesConsistency` logic picks a different "winner" on each reconcile pass and the generated `MeshService` spec flips back and forth. To prevent that, the implementation must enforce one of:

- Validator: a Dataplane MUST NOT carry both `inbound.tags["kuma.io/service"]` and the new field.
- Per-mesh flag: `inboundTagsDisabled` flips per-mesh, exactly one generator path per mesh.

#### Security implications and review

- DP token scope unchanged; CP remains sole `MeshService` writer.
- No new privileged DP-to-CP channel. Option A's auth-broadening is avoided.
- `kuma.io/workload` and `inbound.meshServices` are validated at Dataplane registration.

#### Reliability implications

- `WorkloadStatus.Conditions[PortConflict|LabelConflict]` are set and cleared on every reconcile pass; stale `True` values are unacceptable and must be tested.
- Conflict signals must mirror to `DataplaneInsight` so `kuma-dp` logs surface them locally for restricted-network operators.

### Option D: operator-authored MeshService, no auto-generation

CP no longer generates. Operators run `kumactl apply -f meshservice.yaml`.

* Good. Deletes the polling generator, `checkMeshServicesConsistency`, grace-period delete, and ownership tracking. Smallest CP code.
* Good. No conflict resolution (single writer). No `WorkloadStatus.Conditions` to maintain.
* Good. Symmetric with Kubernetes (operator authors the service-identity resource).
* Good. No new fields on `Dataplane` spec.
* Bad. Blocks the restricted-network operator entirely.
* Bad. Universal users must run a parallel CI/GitOps surface for MeshService.
* Bad. The tactical label-propagation patch cannot ship under it.

## Tactical patch (independent, ships in 2.14)

The "tactical patch" is a small immediate change that ships ahead of the structural decision and closes the user-reported issue (the field report). It is independent of which option (A-D) is ultimately chosen.

- Generator propagates non-`kuma.io/*` keys from `Dataplane.metadata.labels` to `MeshService.metadata.labels`, first-DP-wins. Closes the field report. Inbound tags do not propagate, which signals deprecation. Once shipped, MeshMultizoneService selectors in the wild depend on it.
- A silent MeshMultizoneService selector miss (the selector resolves to zero MeshServices) is the dominant failure. CP emits a structured warning and metric whenever this happens in a zone. Lands with the tactical patch.

## Implications for Kong Mesh

None specific to the downstream project. The 2.14-to-3.0 migration applies uniformly to all Kuma users; any downstream-specific code that derives MeshServices from inbound tags or matches policies on `kuma.io/service` must be updated alongside Kuma 3.0.

## Decision

Not generating MeshService on Universal is most clean solution. It removes all the ambiguities that come with MeshService generation.
It leaves full control over MeshService to mesh operator, they can label it as they need for grouping in MeshMultizoneService.
