# Finalized schemes for unified Envoy resources and stats naming formats (KRI, contextual, system)

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/13256

## Context and Problem Statement

Envoy resource and stat names were inconsistent, hard to correlate with Kuma resources, and often increased cardinality. Earlier MADRs defined: the KRI string format for user resources, a contextual format for proxy-local resources, and a system format for internal resources. During implementation we adjusted the contextual format to include a short scope so we can clearly distinguish sidecar dataplanes from ZoneIngress and ZoneEgress.  

## Decision

Use three complementary naming formats:

1. **KRI** for resources that directly map to a Kuma resource (MeshService, MeshExternalService, MeshMultiZoneService, policies, gateways, ingress, egress, etc.)
2. **Contextual `self_…`** for resources local to a single proxy. This includes inbounds and transparent-proxy passthrough, and now carries an explicit **scope** token to mark the proxy kind
3. **System `system_…`** for internal resources. When a system resource still originates from a Kuma resource, use `system_<kri>`; otherwise use a descriptive `system_<namespace>…` name.

The formats below define the exact strings, character sets, and examples. This MADR is a summary of naming. It does not describe rollout or feature flags.

## Common rules

- Avoid `:` in names. Prefer `_`. This keeps xDS, admin output, and metrics consistent
- Use `_` as the delimiter. Fields may be empty but their position is kept
- **sectionName** is either a port number or a DNS-like label. It must match:

  ```
  (([1-9][0-9]{0,4})|([a-z0-9](?!.*--)(?!.*\.\.)[a-z0-9.-]*[a-z0-9]))
  ```

  This merges Kubernetes Service and Container port-name rules, and allows `.` for future cases

- Resource names and the emitted stat name should be identical to keep a 1:1 mapping

## KRI format (user resources)

### Purpose

Resources that directly map to user-facing Kuma resources.

### Shape

```
kri_<resourceType>_<mesh>_<zone>_<namespace>_<resourceName>_<sectionName>
```

If a field is not present, keep its slot as an empty string so positions remain stable.

### Allowed characters

KRI uses a conservative charset suitable for URLs, xDS, and metrics. See "Constraints" in the KRI MADR.

### Examples

```
kri_msvc_mesh-1_us-east-2_kuma-demo_backend_httpport
kri_extsvc_mesh-1__kuma-system_es1_
kri_mhttpr_mesh-1_us-east-2_kuma-demo_route-1_
kri_zi__us-east-2_kuma-system_zi1_
kri_ze__us-east-2_kuma-system_ze1_
```

These examples show user resources such as MeshService, MeshExternalService, MeshHTTPRoute, ZoneIngress, and ZoneEgress.

### Notes

- KRI avoids IPs and other high-cardinality bits.

## Contextual format (proxy-local)

### Purpose

Some resources live only in the context of one proxy. The contextual format keeps names short and stable, and avoids mesh, zone, namespace, and proxy identity to prevent cardinality spikes. The chosen keyword is `self`.

### Descriptor structure with the implementation adjustment

We extend the descriptor with a **scope** token that marks the proxy kind:

```
self_<category>_<scope>_<rest>
```

- `<category>`: the feature area, for example `inbound` or `transparentproxy_passthrough`
- `<scope>`: one of `dp` (sidecar dataplane), `zi` (ZoneIngress), `ze` (ZoneEgress)
- `<rest>`: category-specific tail that may include `sectionName`, direction, IP version, etc.

This adjustment replaces earlier examples like `self_inbound_httpport` with `self_inbound_dp_httpport`, so Zone can use the same category while staying unambiguous. *(Undocumented change made during implementation)*

### Categories and shapes

#### Inbounds

```
self_inbound_<scope>_<sectionName>
```

Examples:

```
self_inbound_dp_httpport
self_inbound_dp_8080
self_inbound_zi_10001
self_inbound_ze_10001
```

The old `self_inbound_<sectionName>` form is now scoped. This keeps names short and avoids KRI for inbounds, which would raise metric cardinality.

#### Transparent-proxy passthrough

```
self_transparentproxy_passthrough_<scope>_<direction>_ipv<ipVersion>
```

`<direction>` is `inbound` or `outbound`; `<ipVersion>` is `4` or `6`. Examples:

```
self_transparentproxy_passthrough_dp_inbound_ipv4
self_transparentproxy_passthrough_dp_outbound_ipv6
```

Originally defined without a scope token; we append `<scope>` to align with the adjusted contextual scheme.

### Why contextual here

- Keeps names local to a proxy and avoids mesh/zone fields that are already metric labels
- Prevents churn and excessive label sets in Prometheus and similar systems
- Matches how policies refer to inbounds by `sectionName`

### sectionName

When present in contextual names, `sectionName` follows the same regex and rules as KRI. Use the port name if set, otherwise the port number.

## System format (internal resources)

### Purpose

Use `system_…` so users can filter system metrics quickly and we can namespace internal xDS entities. There are two shapes.

### Shapes

#### System resources that still come from a Kuma resource

```
system_<kri>
```

Example:

```
system_kri_mgrl___kong-mesh-system_global-rate-limit-policy_
```

This covers cases like the MeshGlobalRateLimit cluster.

#### Purely internal components

Use a namespaced, descriptive prefix:

```
system_dynamicconfig_dns
system_metrics_prometheus
system_dns_builtin
system_kube_api_server_bypass
system_envoy_admin
```

### Regex

A simple regex for these names is `^system_([a-z0-9-]*_?)+$`.

### Examples list

The MADR on system naming includes a backfill table for current components such as Dynamic Config, MeshMetric, MeshTrace, MeshAccessLog, readiness, DNS, and others.

## Format selection by proxy and resource type

### Sidecar dataplane

- Outbound listeners, clusters, routes: **KRI** of the target MeshService or policy
- Inbounds: **Contextual** `self_inbound_dp_<sectionName>`. *(Adjusted form.)*
- Transparent-proxy passthrough: **Contextual** `self_transparentproxy_passthrough_<direction>_ipv<ipVersion>`
- Internal helpers like readiness or DNS: **System**.

### ZoneIngress

- Resources local to the ingress proxy: **Contextual** using `…_zi_…` scope, for example `self_inbound_zi_10001`. *(Adjusted from earlier tables that showed KRI.)*
- Anything derived directly from a Kuma resource and not proxy-local: **KRI**.
- Internal helpers: **System**.

### ZoneEgress

- Resources local to the egress proxy: **Contextual** using `…_ze_…` scope. *(Adjusted.)*
- Anything derived directly from a Kuma resource and not proxy-local: **KRI**
- Internal helpers: **System**

## Rationale highlights

- **KRI** gives a precise, stable identifier for user resources and lets tools correlate xDS, APIs, and metrics.
- **Contextual** names keep proxy-local entities small and readable, and avoid large cardinality from embedding mesh, zone, namespace, or proxy name.
- **System** names make it trivial to drop or filter internal metrics and keep dashboards focused.

## Examples at a glance

```
# KRI
kri_msvc_mesh-1_us-east-2_kuma-demo_backend_httpport
kri_extsvc_mesh-1__kuma-system_es1_

# Contextual (adjusted with scope)
self_inbound_dp_httpport
self_inbound_zi_10001
self_transparentproxy_passthrough_dp_inbound_ipv4
self_transparentproxy_passthrough_dp_outbound_ipv6

# System
system_dynamicconfig_dns
system_metrics_prometheus
system_kri_mgrl___kong-mesh-system_global-rate-limit-policy_
```

References for the shapes and examples above are in the original MADRs for KRI, contextual naming, and system naming.  

## Non-goals and scope notes

* Legacy `kuma.io/service` naming and migration steps are out of scope here. This summary only specifies the string formats.

## Appendix: quick mapping cheat-sheet

- **User-facing, directly tied to a Kuma resource** → `kri_…`
- **Proxy-local**
  - sidecar dataplane → `self_…_dp_…`
  - ZoneIngress → `self_…_zi_…`
  - ZoneEgress → `self_…_ze_…`
- **Internal plumbing**
  - tied to a Kuma resource → `system_<kri>`
  - generic internal component → `system_<namespace>[_…]`
