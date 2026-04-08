# SNI Format Improvements

- Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/16041

## Context and Problem Statement

### Background

SNI (Server Name Indication) is used in Kuma for cross-zone communication to route traffic through zone proxies.
The current SNI format was introduced in [MADR 055](055-sni.md) with the following structure:

```
<format-version><hash>.<resource-name>.<port>.<mesh-name>.<resource-type>
```

For example:

- `ae10a8071b8a8eeb8.backend.8080.demo.ms` - MeshService named `backend` in `demo` mesh
- `ae10a8071b8a8eeb8.google.0.default.mes` - MeshExternalService named `google` in `default` mesh
- `ae10a8071b8a8eeb8.backend.8080.demo.mzms` - MeshMultiZoneService

This format was designed to:

- Be a valid DNS hostname (avoiding characters like `,{/`)
- Stay within length limits (max ~157 chars)
- Support differentiation between resource types via suffixes (`ms`, `mes`, `mzms`)

[MADR 065](065-sni-in-the-resource.md) further decided to store the SNI directly in the resource definition
rather than computing it dynamically, but it was never implemented.

### Current limitations

The current SNI format has several issues:

#### 1. Zone is not included in the SNI

The SNI does not include the zone where the resource was created.
This can lead to naming collisions when `MeshExternalService` resources with the same name
are both synced from the global control plane and created locally in a zone.

For example,
if zone `zone-a` and the global CP both have a MeshExternalService named `external-api`,
their SNIs would be identical,
causing routing conflicts.

#### 2. Opaque hash prevents SNI <-> KRI conversion

The current format includes a hash (`<format-version><hash>`) which:

- Makes the SNI opaque — users cannot construct it from known resource attributes
- Makes it impossible to reverse an SNI back into a KRI, since the hash is a lossy transformation
- Complicates integrations that need to map between SNIs and resource identifiers programmatically
- Reduces debuggability when troubleshooting routing issues

A bidirectional SNI <-> KRI mapping would allow components to resolve a resource
from an observed SNI (e.g. in Envoy logs or access logs) without needing a lookup table.

#### 3. Inconsistent handling between resource types

The code has special handling for MeshService vs MeshExternalService/MeshMultiZoneService
when computing the resource name for SNI:

```go
// From pkg/xds/generator/zoneproxy/destinations.go
switch r := any(dest).(type) {
case *meshservice_api.MeshServiceResource:
    rName = r.SNIName(systemNS)
default:
    rName = core_model.GetDisplayName(dest.GetMeta())
}
```

MeshService uses a special `SNIName()` method that accounts for namespace scoping on Kubernetes
and various resource origin scenarios:

- Resources synced from global vs local zone resources
- Kubernetes namespaced resources vs Universal resources
- System namespace handling

MeshExternalService and MeshMultiZoneService are cluster-scoped on Kubernetes
and use a simpler `GetDisplayName()` approach,
but this inconsistency makes the codebase harder to maintain and reason about.

## Decision Drivers

- **Zone-awareness**:
SNIs must uniquely identify resources across zones to prevent routing collisions,
especially for MeshExternalService resources that may exist both globally and locally
- **SNI <-> KRI reversibility**:
It should be possible to convert between SNI and KRI in both directions —
construct an SNI from a KRI, and recover a KRI from an observed SNI —
without requiring a lookup table
- **Consistency**:
The format should handle MeshService, MeshExternalService, and MeshMultiZoneService uniformly
without special-case code paths
- **Valid DNS hostname**:
SNI must remain a valid DNS hostname per [RFC 6066](https://datatracker.ietf.org/doc/html/rfc6066)
- **Alignment with KRI**:
Consider how the SNI format can align with or leverage the KRI format adopted in MADR 070

## Design

### Option 1: New SNI format derived from KRI

#### Format

Creating a mapping of all possible KRIs to SNI is tricky and doesn't make much sense.
Main problem is that KRI allows empty segments i.e. `kri_m____default_` is valid KRI,
however DNS hostnames don't allow empty segments, i.e. `kri.m....default.` is not a valid hostname.

However, the problem can be solved for a subset of KRIs that satisfy the constraints:

- resource types are `MeshService`, `MeshExternalService` and `MeshMultiZoneService`
- `mesh`, `name` and `sectionName` are always non-empty
- if `namespace` is non-empty then `zone` is non-empty

These constraints follow naturally from Kuma's resource model and are reasonable in the context of SNI matching.

In that case, the KRI `kri_<type>_<mesh>_<zone>_<namespace>_<name>_<sectionName>` can be unambiguously mapped to
`sni.<type>.<mesh>.<zone>.<namespace>.<name>.<sectionName>` omitting the empty segment.

It's unambiguous because different cases have different number of segments:

- global-originated resource - `sni.<type>.<mesh>.<name>.<sectionName>` (5 segments)
- zone-originated resource in system namespace - `sni.<type>.<mesh>.<zone>.<name>.<sectionName>` (6 segments)
- zone-originated resource in custom namespace - `sni.<type>.<mesh>.<zone>.<namespace>.<name>.<sectionName>` (7 segments)

#### DNS hostname limitation

[RFC 1035](https://datatracker.ietf.org/doc/html/rfc1035) enforces limitations:

- Each segment (DNS label) must be 63 characters or less
- Total length of the hostname must be 253 characters or less

There is no issue with each segment being 63 or less, however, total length is the issue.

The max length of the SNI would be:

- sni (3 chars)
- type (6 chars because of the "extsvc")
- mesh, zone, namespace, name, section name (63 chars each)
- dots (6 chars)

Total: 3 + 6 + 63*5 + 6 = 330

So why even consider this option even though it exceeds the DNS hostname limit?
I think there are few reasons to that:

- Single mesh future. We've discussed multiple times that single mesh should the way to run Kuma. 
In that case instead of 63 the mesh is likely to be 7 ("default").
- 63 char for port name is way too much (even though Service allows port names upto 63, port name on Container is only 15 chars max)

We don't need to limit individual segments, however,
when user creates `MeshService`, `MeshExternalService` or `MeshMultiZoneService`,
we can validate the length of `mesh`, `zone`, `namespace`, `name` and `port` + 15 (for "sni", "type" and dots) less than 253.

#### Migration

Only new mesh-scoped zone proxies adopting a new format.

### Option 2: Improve existing SNI format

#### Format

Keep the existing SNI format unchanged:

```
<format-version><hash>.<resource-name>.<port>.<mesh-name>.<resource-type>
```

The existing `SNIForResource` function already accepts an `additionalData map[string]string` parameter
that gets mixed into the FNV hash:

```go
func SNIForResource(resName string, meshName string, resType model.ResourceType, port int32, additionalData map[string]string) string
```

Today, all callers pass `nil` for `additionalData`.
The fix is to systematically pass `zone` and `namespace` (when present) in this map.

#### Solving the zone collision problem

The zone collision described in [limitation 1](#1-zone-is-not-included-in-the-sni) happens because
resources from different zones with the same name produce identical SNIs.
By including `zone` (and `namespace` when applicable) in `additionalData`,
the hash will differ even when resource name, mesh, port, and type are the same.

For example, both global CP and zone `zone-a` having a MeshExternalService named `external-api`
would currently produce:

```
ae10a8071b8a8eeb8.external-api.443.default.mes
```

With zone added to the hash, they produce different SNIs:

```
ae10a8071b8a8eeb8.external-api.443.default.mes   (global, no zone)
b72f3a9c1d4e58a01.external-api.443.default.mes   (zone-a)
```

The visible part of the SNI (resource name, port, mesh, type) stays the same — only the hash prefix changes.

#### Unifying resource name computation

Currently, `MeshService` uses a special `SNIName(systemNamespace)` method
that encodes namespace and origin information into the resource name segment of the SNI.
`MeshExternalService` and `MeshMultiZoneService` use `GetDisplayName()`.

With `zone` and `namespace` moved into `additionalData`,
the `SNIName()` special-casing can be removed.
All resource types can consistently use `GetDisplayName()` for the resource name segment,
and differentiation happens via the hash.
This resolves the inconsistency described in [limitation 3](#3-inconsistent-handling-between-resource-types).

#### What this does NOT solve

This option does **not** address [limitation 2](#2-opaque-hash-prevents-sni--kri-conversion).
The SNI remains hash-based, so:

- It is impossible to reverse an SNI back into a KRI — the hash is a lossy transformation
- Users and integrations still cannot construct an SNI from known resource attributes
- Mapping between SNIs and resources still requires a lookup against stored SNI values

#### Migration

Changing what goes into the hash produces different SNIs for existing resources.
This requires:

- A format version bump (e.g. `a` → `b`) so old and new SNIs are distinguishable
- Only new mesh-scoped zone proxies adopting a new format

### Pros and Cons

#### Option 1: New SNI format derived from KRI

- Good, because SNI <-> KRI conversion is bidirectional —
  given an SNI (e.g. from Envoy logs),
  the originating resource KRI can be recovered without a lookup table
- Good, because SNIs are fully human-readable —
  users can construct and interpret them from known resource attributes
- Good, because it aligns with the KRI format [MADR 070](070-resource-identifier.md),
  giving a consistent identifier scheme across the system
- Good, because zone, namespace, and all identifying attributes are explicit segments,
  not hidden in a hash
- Good, because it eliminates the `SNIName()` special-casing —
  all resource types use the same KRI-to-SNI mapping
- Bad, because it can exceed the 253-character DNS hostname limit
  when multiple segments are at their maximum length (up to 330 chars)
- Bad, because it requires validation at resource creation time
  to reject combinations whose SNI would exceed 253 characters
- Bad, because it is a completely new format,
  requiring more complex migration logic (dual-format support across all zones)
- Neutral, because the 253-char limit is unlikely to be hit in practice
  (single-mesh deployments, short port names)

#### Option 2: Improve existing SNI format

- Good, because it is a minimal, low-risk change —
  the format and generation logic stay the same,
  only the hash inputs change
- Good, because it solves the zone collision problem (limitation 1)
  by including zone and namespace in the hash
- Good, because it removes the `SNIName()` special-casing,
  unifying resource name computation across all resource types (limitation 3)
- Good, because SNI length stays bounded by the existing format
  (no risk of exceeding DNS hostname limits)
- Bad, because SNI → KRI reversal is impossible —
  the hash is lossy,
  so an observed SNI cannot be mapped back to a resource without a lookup
- Bad, because users and integrations still cannot construct or predict SNIs
  from resource attributes
- Bad, because debugging routing issues still requires looking up the stored SNI
  rather than reasoning about it from the resource name

## Decision

Option 1: New SNI format derived from KRI.

The key deciding factor is SNI <-> KRI reversibility.
Option 2 makes this impossible by design,
since the hash is a lossy transformation.
With option 1,
any component observing an SNI
(access logs, debugging tools, observability pipelines)
can resolve the originating resource
without a lookup table.

The 253-char DNS hostname limit is a theoretical concern
that is unlikely to matter in practice —
single mesh is the recommended deployment model,
and container port names are capped at 15 chars.
Creation-time validation provides a safety net
for the rare edge cases.

## Notes

Related MADRs:

- [MADR 055](055-sni.md) - Original SNI format for MeshService, MeshExternalService
- [MADR 065](065-sni-in-the-resource.md) - SNI stored in resource definition
- [MADR 070](070-resource-identifier.md) - Resource Identifier (KRI) format

