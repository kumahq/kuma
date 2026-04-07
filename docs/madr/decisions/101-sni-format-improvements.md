# SNI Format Improvements

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/16041

## Context and Problem Statement

### Background

SNI (Server Name Indication) is used in Kuma for cross-zone communication to route traffic through zone proxies.
The current SNI format was introduced in MADR 055 with the following structure:

```
<format-version><hash>.<resource-name>.<port>.<mesh-name>.<resource-type>
```

For example:
* `ae10a8071b8a8eeb8.backend.8080.demo.ms` - MeshService named `backend` in `demo` mesh
* `ae10a8071b8a8eeb8.google.0.default.mes` - MeshExternalService named `google` in `default` mesh
* `ae10a8071b8a8eeb8.backend.8080.demo.mzms` - MeshMultiZoneService

This format was designed to:
* Be a valid DNS hostname (avoiding characters like `,{/`)
* Stay within length limits (max ~157 chars)
* Support differentiation between resource types via suffixes (`ms`, `mes`, `mzms`)

MADR 065 further decided to store the SNI directly in the resource definition
rather than computing it dynamically.

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

#### 2. Opaque hash makes integration difficult

The current format includes a hash (`<format-version><hash>`) which:
* Makes the SNI opaque to users who cannot easily construct it from known resource attributes
* Complicates integrations that need to predict or construct SNIs programmatically
* Reduces debuggability when troubleshooting routing issues

The hash was originally introduced to handle uniqueness and tags,
but with the current resource model where SNI is stored in the resource itself,
this may no longer be necessary.

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
* Resources synced from global vs local zone resources
* Kubernetes namespaced resources vs Universal resources
* System namespace handling

MeshExternalService and MeshMultiZoneService are cluster-scoped on Kubernetes
and use a simpler `GetDisplayName()` approach,
but this inconsistency makes the codebase harder to maintain and reason about.


## Decision Drivers

* **Zone-awareness**:
  SNIs must uniquely identify resources across zones to prevent routing collisions,
  especially for MeshExternalService resources that may exist both globally and locally

* **Human-readability**:
  Users and integrations should be able to understand and construct SNIs from known resource attributes
  without needing to compute hashes

* **Consistency**:
  The format should handle MeshService, MeshExternalService, and MeshMultiZoneService uniformly
  without special-case code paths

* **Valid DNS hostname**:
  SNI must remain a valid DNS hostname per RFC 6066

* **Alignment with KRI**:
  Consider how the SNI format can align with or leverage the KRI format adopted in MADR 070

## Design

### Option 1: New SNI format derived from KRI

#### Format

Creating a mapping of all possible KRIs to SNI is tricky and doesn't make much sense.
Main problem is that KRI allows empty segments i.e. `kri_m____default_` is valid KRI,
however DNS hostnames don't allow empty segments, i.e. `kri.m....default.` is not a valid hostname.

The problem can be solved for a subset of KRIs that satisfy the constraints:
* resource types are `MeshService`, `MeshExternalService` and `MeshMultiZoneService`
* `mesh`, `name` and `sectionName` are always non-empty
* if `namespace` is non-empty then `zone` is non-empty 

In that case, the KRI `kri_<type>_<mesh>_<zone>_<namespace>_<name>_<sectionName>` can be unambiguously mapped to
`sni.<type>.<mesh>.<zone>.<namespace>.<name>.<sectionName>` omitting the empty segment.

It's unambiguous because different cases have different number of segments:
* global-originated resource - `sni.<type>.<mesh>.<name>.<sectionName>` (5 segments)
* zone-originated resource in system namespace - `sni.<type>.<mesh>.<zone>.<name>.<sectionName>` (6 segments)
* zone-originated resource in custom namespace - `sni.<type>.<mesh>.<zone>.<namespace>.<name>.<sectionName>` (7 segments)

#### DNS hostname limitation

[RFC 1035](https://datatracker.ietf.org/doc/html/rfc1035) enforces limitations:
* Each segment (DNS label) must be 63 characters or less
* Total length of the hostname must be 253 characters or less

There is no issue with each segment being 63 or less, however, total length is the issue.

The max length of the SNI would be:
* sni (3 chars)
* type (6 chars because of the "extsvc")
* mesh, zone, namespace, name, section name (63 chars each)
* dots (6 chars)

Total: 3 + 6 + 63*5 + 6 = 330

So why even consider this option even though it exceeds the DNS hostname limit?
I think there are few reasons to that:

* Single mesh future. We've discussed multiple times that single mesh should the way to run Kuma. 
In that case instead of 63 the mesh is likely to be 7 ("default").
* 63 char for port name is way too much (even though Service allows port names upto 63, port name on Container is only 15 chars max)

We don't need to limit individual segments, however,
when user creates `MeshService`, `MeshExternalService` or `MeshMultiZoneService`,
we can validate the length of `mesh`, `zone`, `namespace`, `name` and `port` + 15 (for "sni", "type" and dots) less than 253.

### Option 2: Improve existing SNI format

Existing format supports `additionalData`.
It's a `map[string]string` that's going to be added to the hash.

We can systematically add `zone` and `namespace` to the `additionalData`.

## Decision

{To be determined}

## Notes

Related MADRs:
* MADR 055 - Original SNI format for MeshService, MeshExternalService
* MADR 065 - SNI stored in resource definition
* MADR 070 - Resource Identifier (KRI) format

