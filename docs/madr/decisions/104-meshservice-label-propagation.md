# MeshService label propagation (Universal)

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/16542

## Context and Problem Statement

When the MeshService generator creates or updates auto-generated MeshService
objects on Universal (see MADR 050 for the autogen decision), only system
labels are written: `kuma.io/mesh`, `kuma.io/managed-by`, `kuma.io/env`,
`kuma.io/zone`, `kuma.io/origin`, `kuma.io/display-name`.

Operator-defined metadata on Dataplane resources — inbound tags like `team`,
`tier`, or resource-level labels like `color` — is silently discarded. This
blocks two use cases:

1. `MeshMultiZoneService` selector-by-label: if the selectors reference
   user-defined keys (e.g. `team: payments`), they cannot match
   auto-generated MeshServices on Universal because those keys are absent.
2. Policy matching by label: policies that target MeshServices via label
   selectors have no user-defined keys to match against.

## Design

The propagation pipeline runs in three stages before the existing
`desiredLabels` call in the generator:

### Stage 1 — intra-DP inbound consensus

For a single Dataplane, each inbound can carry different tags. A key is
included in the contribution only when **all inbounds agree on its value**;
any key with conflicting values across inbounds is dropped for that DP.
This prevents a DP with heterogeneous inbounds from generating noise.

Reserved prefixes (`kuma.io/`, `k8s.kuma.io/`) are stripped before the
consensus check.

### Stage 2 — validation and allow-list filter

Each candidate key/value pair is validated with `apimachineryvalidation.
IsQualifiedName` (key) and `k8s.io/apimachinery/pkg/api/validation.
IsValidLabelValue` (value). Invalid pairs are dropped and a Prometheus
counter is incremented; they are never surfaced as Dataplane validation
errors (non-breaking).

If `meshService.labelPropagation.allowedLabelKeys` is non-empty, only keys
in that list survive.

### Stage 3 — DP resource-label overlay

After inbound tags, the Dataplane's own resource-level labels are merged in
using the same reserved-prefix filter and allow-list. DP-level labels take
precedence over inbound tags for the same key within a single DP.

### Cross-DP merge

When multiple Dataplanes contribute to the same MeshService name:

- **Per-key majority-wins**: the value held by the most Dataplanes wins.
- **Newest-tiebreak**: when two values tie on count, the one from the
  Dataplane with the most recent creation time wins.
- **Lexicographic final tiebreak**: if creation timestamps are also equal,
  the lexicographically smaller value wins (deterministic across restarts).

This limits label churn during rolling deployments: as long as the majority
of Dataplanes agree on a value, the MeshService label is stable.

### Generator wiring

The propagation call site is `generator.go:227-240`. The function
`propagatedLabelsFor` is a no-op when:

- `g.labelPropagationEnabled == false` (flag off, default).
- The mesh `meshServices.mode` is not `Exclusive` (generator short-circuits
  before reaching label code for other modes).

The result is passed into `desiredLabels`, which merges propagated keys
with the mandatory system keys. System keys always override propagated keys
with the same name (e.g. `kuma.io/mesh` cannot be hijacked).

### External-label preservation

On Update, keys present in the existing MeshService but absent from
`desired` are copied back (`generator.go:268-275`). This preserves labels
set out-of-band by operators without overwriting them on every reconcile
cycle.

### No-op Update guard

The generator skips the Update call when `desired` equals `existingLabels`
(`generator.go:276`). Under steady state this means zero label-related KDS
churn.

## Security implications and review

Label values are operator-controlled input. The reserved-prefix gate
(`kuma.io/`, `k8s.kuma.io/`) ensures external metadata cannot overwrite
platform-owned keys on the MeshService. Invalid label keys or values are
silently dropped (logged at V(1)) and never raised as validation errors on
the originating Dataplane, so malformed metadata cannot cause DP registration
to fail.

## Reliability implications

The per-key majority-wins merge bounds the MeshService label churn during
rolling deploys to a single flip per key (old majority → new majority).
Combined with the no-op Update guard, steady-state KDS traffic is unchanged.

The eventually-consistent reconcile loop (default interval 2 s) means newly
registered Dataplanes take up to one interval to influence MeshService labels.
This is acceptable for label use cases (policy matching, multi-zone selectors)
because those paths are not latency-sensitive.

## Implications for Kong Mesh

Flag default-off in this minor release. Kong Mesh may enable it in the same
release cycle once integration tests cover the propagated-label path. No API
or wire-format changes are required beyond setting the feature flag.

## Decision

Adopt the three-stage compute (intra-DP consensus → validation/allow-list →
DP label overlay) plus cross-DP majority-wins merge described above. Ship
behind `KUMA_MESH_SERVICE_LABEL_PROPAGATION_ENABLED` (default `false`).
Complements MADR 050 (auto-generated MeshService on Universal) by adding
user-metadata propagation to the objects 050 decided to auto-generate.

## Notes

Sub-issues for audit trail:
- [#16559](https://github.com/kumahq/kuma/issues/16559): pure label compute helpers
- [#16564](https://github.com/kumahq/kuma/issues/16564): generator wiring
- [#16547](https://github.com/kumahq/kuma/issues/16547): Universal e2e (this PR)

Removal-on-tag-drop semantics (removing a propagated label when no Dataplane
contributes it any more) require a separate ownership-tracking design to
distinguish "previously propagated" from "operator-set" labels. Tracked as a
follow-up; not addressed here.
