# Naming Envoy Resources and Stats from Multiple Kuma Policies

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/13886

## Context and Problem Statement

### Background: Unified Naming Strategy

Kuma has established a unified naming strategy for Envoy resources and stats through four foundational MADRs:

- [MADR-070: Resource Identifier](070-resource-identifier.md)
- [MADR-077: Standardized Naming for System xDS Resources](077-naming-system-envoy-resources.md)
- [MADR-078: Consistent Naming for Non-System Envoy Resources and Stats](078-migrating-to-consistent-and-well-defined-naming-for-non-system-envoy-resources-and-stats.md)
- [MADR-084: Finalized Schemes for Unified Envoy Resources and Stats Naming](084-finalized-schemes-for-unified-envoy-resources-and-stats-naming-formats.md)

The fundamental principle is straightforward: **the name of an Envoy resource should correspond to its origin** using one of three complementary formats:

- **KRI format** (`kri_<type>_<mesh>_<zone>_<namespace>_<name>_...`) - for resources created by Kuma policies or services
- **Contextual format** (`self_<component>_<scope>_...`) - for proxy-local resources not owned by a specific policy
- **System format** (`system_<component>_...`) - for internal Kuma infrastructure resources

This naming strategy enables users to:

- Trace metrics back to the Kuma resource that generated them
- Understand which policy or component is responsible for specific Envoy configuration
- Filter and query metrics by Kuma resource attributes (mesh, zone, namespace, name)
- Debug and troubleshoot by correlating Envoy behavior with Kuma configuration

### The Problem: Multi-Policy Resources

This principle works well when there is a **one-to-one relationship** between a Kuma resource and an Envoy resource. However, **some Envoy resources are the result of multiple Kuma policies being merged together**.

**Common scenarios where multiple policies merge:**
- Routes from multiple `MeshHTTPRoute` policies targeting the same service
- RBAC filters combining rules from multiple `MeshTrafficPermission` policies
- HTTP connection managers where multiple `MeshTimeout`, `MeshRetry`, or `MeshRateLimit` policies apply
- Clusters where multiple `MeshHealthCheck` or `MeshCircuitBreaker` policies merge

These multi-policy scenarios were explicitly marked as out of scope in previous MADRs. **This MADR addresses that gap** by defining how to handle naming for Envoy resources and components that result from multiple policies.

### Why This Matters

**Observability impact**: Current ad-hoc naming (hashes, fixed strings, undefined names) creates:

- Traceability gaps - cannot identify which policy caused observed behavior
- Inconsistency with single-policy resources
- Complex troubleshooting
- Difficult dashboard and alert creation

**Design constraints**: Solutions must balance:

- **Traceability** vs. **cardinality** - including all policy details risks metric explosion
- **Stability** vs. **completeness** - names should be stable across policy changes
- **Future-proofing** - must accommodate new policies and merge scenarios

**Tooling impact**: Naming affects Kuma GUI, Prometheus, OpenTelemetry exporters, and third-party tools.

## Scope

### In Scope

- Envoy resources and components that **generate stats and metrics** where multiple policies merge
- Defining naming approach for multi-policy scenarios
- Migration considerations

### Out of Scope

- Already-solved naming patterns (KRI-based, contextual, system)
- Resources without observability impact (no stats/metrics)
- Non-xDS resources (endpoints, secrets)
- Deprecated `kuma.io/service` tag-based naming
- Default routes (will be addressed in separate MADR)
- **MeshPassthrough** (will be addressed in separate MADR)

## Envoy Resources That Generate Stats

Naming primarily matters for resources generating stats and metrics.

### Primary Resources

1. **Listeners** - Generate `listener.<name>.*` stats
   - Examples: `downstream_cx_active`, `downstream_cx_total`, `downstream_rq_active`

2. **Clusters** - Generate `cluster.<name>.*` stats
   - Examples: `upstream_cx_active`, `upstream_cx_total`, `upstream_rq_total`, `health_check.success`

3. **Routes** - Generate stats within HTTP connection manager using virtual host name
   - Individual route entries tracked through tracing and access logs

### Nested Components

**Network Filters** (applied to listeners):

- **HTTPConnectionManager** - Uses `stat_prefix` for stats like `http.<stat_prefix>.downstream_cx_active`
- **TCPProxy** - Uses `stat_prefix` for stats like `tcp.<stat_prefix>.downstream_cx_total`
- **RBAC** - Uses `stat_prefix` for stats like `<stat_prefix>.allowed`, `<stat_prefix>.denied`
  - **Often merged from multiple `MeshTrafficPermission` policies**

**HTTP Filters** (applied within HTTPConnectionManager):

- **RBAC** - Uses `stat_prefix`, similarly merged from multiple policies
- Various other filters with stat prefixes

**Route Configuration**:

- Contains virtual hosts and routes
- **Route entries can merge from multiple `MeshHTTPRoute` policies**

## Envoy Resource Composition in Kuma

The diagrams below show Envoy resource hierarchy. Each component is annotated with "Affected by:" to indicate which policies influence its configuration, identifying where multiple policies merge.

### Listener Composition

```
Listener
├── name
│   └── Affected by: MeshGateway (gateway), Dataplane (sidecar)
├── address/port
│   └── Affected by: Dataplane, MeshGateway, networking.transparentProxying
└── filter_chains → filters
    ├── HTTPConnectionManager (HTTP traffic)
    │   ├── stat_prefix
    │   │   └── Affected by: MeshService/MeshExternalService/MeshMultiZoneService
    │   ├── route_config or RDS
    │   │   ├── routes[]
    │   │   │   ├── name (uses policy KRI for attribution)
    │   │   │   │   └── Affected by: MeshHTTPRoute (MULTIPLE POLICIES MERGE)
    │   │   │   ├── route.timeout
    │   │   │   │   └── Affected by: MeshTimeout (multiple policies)
    │   │   │   └── route.retry_policy
    │   │   │       └── Affected by: MeshRetry (multiple policies)
    │   │   └── virtual_hosts[].name
    │   │       └── Affected by: MeshService/MeshExternalService/MeshMultiZoneService
    │   └── http_filters
    │       ├── RBAC filter (optional)
    │       │   └── Affected by: MeshTrafficPermission (MULTIPLE POLICIES MERGE)
    │       ├── RateLimit filter (optional)
    │       │   └── Affected by: MeshRateLimit (multiple policies)
    │       ├── FaultInjection filter (optional)
    │       │   └── Affected by: MeshFaultInjection (multiple policies)
    │       └── Router filter (always present)
    └── TCPProxy (TCP traffic)
        ├── stat_prefix
        │   └── Affected by: MeshService/MeshExternalService/MeshMultiZoneService
        └── cluster reference
            └── Affected by: MeshTCPRoute, target service
```

### Cluster Composition

```
Cluster
├── name
│   └── Affected by: MeshService, MeshExternalService, MeshMultiZoneService
├── alt_stat_name
│   └── Affected by: same as name (should match for consistency)
├── load_balancing_policy
│   └── Affected by: MeshLoadBalancingStrategy
├── health_checks
│   └── Affected by: MeshHealthCheck (MULTIPLE POLICIES)
├── circuit_breakers
│   └── Affected by: MeshCircuitBreaker (MULTIPLE POLICIES)
└── outlier_detection
    └── Affected by: MeshCircuitBreaker
```

## Where Multiple Policies Create Naming Challenges

Based on analysis of Kuma's policy application and Envoy resource generation, these are the key areas where multiple policies merge.

### 1. Routes Merged from Multiple MeshHTTPRoute Policies

**Problem**: Multiple `MeshHTTPRoute` policies targeting the same outbound service merge into a single Envoy `RouteConfiguration` with individual route entries from different policies.

**Current behavior**:
- The `RouteConfiguration`, `VirtualHost`, `HTTPConnectionManager` `stat_prefix`, `Listener`, and `Cluster` all use the target service's KRI
- Each route entry uses its source policy's KRI + `_rule_<index>` suffix

**Example**: Listener for `backend-api` with three merged `MeshHTTPRoute` policies:

```yaml
Listener:
  name: kri_msvc_default_zone-1_demo-app_backend-api_httpport
  filter_chains:
    - filters:
        - name: envoy.filters.network.http_connection_manager
          typed_config:
            stat_prefix: kri_msvc_default_zone-1_demo-app_backend-api_httpport
            route_config:
              name: kri_msvc_default_zone-1_demo-app_backend-api_httpport
              virtual_hosts:
                - name: kri_msvc_default_zone-1_demo-app_backend-api_httpport
                  domains: ["*"]
                  routes:
                    - name: kri_mhttpr_default_zone-1_demo-app_backend-routing_rule_0
                      match: { prefix: / }
                    - name: kri_mhttpr_default__kuma-system_client-to-backend_rule_0
                      match: { prefix: /multizone }
                    - name: kri_mhttpr_default_zone-1_demo-app_backend-with-timeout_rule_0
                      match: { prefix: /api }

Cluster:
  name: kri_msvc_default_zone-1_demo-app_backend-api_httpport
```

**Three policies merged**: `backend-routing`, `client-to-backend`, `backend-with-timeout`

**Naming challenge**:

- HTTP stats like `http.<stat_prefix>.downstream_rq_timeout` or `http.<stat_prefix>.downstream_rq_2xx` aggregate across all policies without attribution
- Users cannot determine which specific `MeshHTTPRoute`, `MeshTimeout`, or `MeshRetry` policy contributed to observed behavior
- Individual routes are identifiable in access logs and tracing by policy KRI

**Stats affected**:

- ✅ Access logs and tracing: individual routes identified
- ❌ HTTP connection manager stats: `http.<stat_prefix>.*` aggregate without attribution

### 2. RBAC Filters Merged from Multiple MeshTrafficPermission Policies

**Problem**: Multiple `MeshTrafficPermission` policies applying to the same inbound or outbound merge into a single `RBAC` filter.

**Current behavior**:
- Multiple restrictive `MeshTrafficPermission` policies merge into a single `RBAC` filter (network or HTTP)
- `stat_prefix` uses contextual format for inbounds (e.g., `self_inbound_dp_<sectionName>`) or target service KRI for outbounds

**Example**: Three policies (`allow-from-frontend`, `allow-from-gateway`, `allow-admin-access`) merge into one RBAC filter with a single `stat_prefix`.

**Naming challenge**:
- The `stat_prefix` doesn't reflect which specific `MeshTrafficPermission` policies contributed
- Stats like `<stat_prefix>.allowed` and `<stat_prefix>.denied` aggregate across all merged policies
- Impossible to attribute permission decisions to individual policies

**Stats affected**:
- `<stat_prefix>.allowed`
- `<stat_prefix>.denied`
- `<stat_prefix>.shadow_allowed`
- `<stat_prefix>.shadow_denied`

### 3. Multi-Policy Filter Chains

**Problem**: Listener configurations with TLS/SNI routing or protocol detection may have filter chains where selection and configuration depend on multiple policies.

**Example**: Multiple `MeshTimeout` or `MeshRetry` policies with different selectors affect the same HTTP connection manager, whose `stat_prefix` represents all merged configurations.

**Stats affected**: `http.<stat_prefix>.*`, `tcp.<stat_prefix>.*`

### 4. Configuration Policies That Modify Shared Components

**Problem**: Several policy types modify shared Envoy components rather than creating separately-named resources. When multiple policies apply, stats aggregate without individual policy attribution.

**Policy categories:**

1. **Route-level configuration policies** - Modify individual route entries:
   - `MeshTimeout` - Sets `route.timeout` on each route
   - `MeshRetry` - Configures `route.retry_policy` on each route

2. **HTTP filter policies** - Create HTTP filters inserted before the router:
   - `MeshRateLimit` - Creates `envoy.filters.http.local_ratelimit` filter
   - `MeshFaultInjection` - Creates `envoy.filters.http.fault` filter

3. **Cluster-level policies** - Modify cluster configuration:
   - `MeshHealthCheck` - Configures `cluster.health_checks[]`
   - `MeshCircuitBreaker` - Configures `cluster.circuit_breakers`

**Example**: Route with multiple timeout/retry policies merged:

```yaml
route_config:
  routes:
    - name: kri_mhttpr_default_zone-1_demo-app_backend-routing_rule_0
      route:
        timeout: 15s  # From MeshTimeout policies
        retry_policy:  # From MeshRetry policies
          num_retries: 5
          per_try_timeout: 16s
```

**Naming challenges by policy type:**

**Route-Level Policies:**
- **`MeshTimeout`**
  - Where: `route.timeout`
  - Challenge: Stats aggregate at HCM level across all routes with different timeout policies
  - Stats: `http.<stat_prefix>.downstream_rq_timeout`

- **`MeshRetry`**
  - Where: `route.retry_policy`
  - Challenge: Stats aggregate at HCM level across all routes with different retry policies
  - Stats: `http.<stat_prefix>.retry.*`, `http.<stat_prefix>.upstream_rq_retry`

**HTTP Filter-Level Policies:**
- **`MeshRateLimit`**
  - Where: HTTP filter with `stat_prefix`
  - Challenge: Creates filter but multiple policies merge; filter stat_prefix doesn't reflect which policies
  - Stats: `http.<stat_prefix>.ratelimit.*`

- **`MeshFaultInjection`**
  - Where: HTTP filter with `stat_prefix`
  - Challenge: Creates filter but multiple policies merge; filter stat_prefix doesn't reflect which policies
  - Stats: `http.<stat_prefix>.fault.*`

**Cluster-Level Policies:**
- **`MeshHealthCheck`**
  - Where: `cluster.health_checks[]`
  - Challenge: Cluster stats don't reflect which policies configured the checks
  - Stats: `cluster.<name>.health_check.*`

- **`MeshCircuitBreaker`**
  - Where: `cluster.circuit_breakers`
  - Challenge: Cluster stats don't reflect which policies set thresholds
  - Stats: `cluster.<name>.circuit_breakers.*`

**Key insight**: While `MeshRateLimit` and `MeshFaultInjection` create HTTP filters, all these policies share a common challenge: their stats aggregate at the HTTPConnectionManager or Cluster level using the service KRI, without individual policy attribution.

## Design

### Option 1: Encode All Contributing Policies in Names

**Description**: Include KRIs or identifiers for all contributing policies in the resource name.

**Example**: `kri_msvc_..._backend-api_httpport__policies__mhttpr_policy1__mhttpr_policy2__mtimeout_policy3`

**Pros:**
- Complete policy attribution
- Full traceability

**Cons:**
- **Extreme metric cardinality** - names change with every policy addition/removal
- **Unstable names** - configuration churn causes metric discontinuity
- **Very long names** - difficult to read, may hit length limits
- **Impractical** - defeats the purpose of aggregate metrics

**Verdict**: Technically complete but operationally impractical. Not recommended.

### Option 2: Introduce Multi-Policy Marker Format

**Description**: Use a special marker in names to indicate multi-policy merge without listing all policies.

**Example**: `kri_msvc_..._backend-api_httpport__multipolicy` or `kri_msvc_..._backend-api_httpport__merged`

**Pros:**
- Indicates multi-policy nature
- Stable names
- Low cardinality

**Cons:**
- Still no policy attribution
- Doesn't actually solve the traceability problem
- Adds complexity without significant value

**Verdict**: Provides awareness of multi-policy scenarios but doesn't enable debugging. Not recommended.

### Option 3: Use Aggregate Naming + Policy Labels/Tags (Recommended)

**Description**:
- Use existing naming (service KRI or contextual) for multi-policy resources
- Rely on **access logs, tracing, and per-route/per-filter metrics** for policy-level attribution
- Document which resources aggregate vs. provide attribution

**Approach:**
1. **Aggregate resources** (listeners, clusters, HTTPConnectionManager): Use service KRI or contextual naming
2. **Attributable resources** (individual routes, per-route configurations): Use policy KRI
3. **Observability strategy**:
   - Use access logs with route names for request-level attribution
   - Use distributed tracing for end-to-end policy correlation
   - Use per-route metrics where available
   - Document aggregate vs. attributable metrics

**Example (continued from earlier)**:

```yaml
# Aggregate naming - all policies merge here
HTTPConnectionManager:
  stat_prefix: kri_msvc_default_zone-1_demo-app_backend-api_httpport
  # Stats: http.kri_msvc_default_zone-1_demo-app_backend-api_httpport.*

# Individual route attribution - each route has policy KRI
Routes:
  - name: kri_mhttpr_default_zone-1_demo-app_backend-routing_rule_0
  - name: kri_mhttpr_default__kuma-system_client-to-backend_rule_0
  - name: kri_mhttpr_default_zone-1_demo-app_backend-with-timeout_rule_0
```

**Pros:**
- **Stable names** - don't change with policy additions/removals
- **Low cardinality** - aggregate metrics remain aggregated
- **Practical traceability** - use access logs/tracing for detailed attribution
- **Consistent with Envoy patterns** - matches how Envoy itself handles aggregation
- **Existing solution** - individual routes already use policy KRIs

**Cons:**
- Requires using multiple observability tools (metrics + logs + tracing)
- Not all policy effects visible in high-level metrics
- Requires documentation of which metrics are aggregate vs. attributed

**Verdict**: Most practical approach balancing traceability, cardinality, and operational reality.

### Option 5: Hybrid Approach with Contextual Markers

**Description**: Use contextual naming for truly shared components and policy KRI for policy-specific components.

**Example**:
- Shared RBAC filter: `self_rbac_inbound_dp_httpport` (contextual)
- Policy-specific route: `kri_mhttpr_..._policy_name_rule_0` (policy KRI)

This is essentially what we already do for routes. Could be extended to other scenarios.

**Pros:**
- Clear separation of shared vs. policy-specific
- Aligns with existing contextual format

**Cons:**
- Still doesn't solve aggregate stats problem
- Similar to Option 3 in practice

### Option 6: Maintain Status Quo (Keep Current Naming)

**Description**: Accept that multi-policy resources use aggregate naming based on the primary resource (service KRI or contextual format).

**Pros:**
- No implementation effort required
- No migration needed
- Low metric cardinality
- Stable names across policy changes

**Cons:**
- No policy attribution in metrics
- Inconsistent with single-policy traceability goals
- Difficult troubleshooting when multiple policies interact
- Cannot determine which policy caused observed behavior

**Verdict**: This preserves the current gaps that motivated the naming MADRs. Not recommended.

## Decision

**Adopt Option 3: Use Aggregate Naming + Policy Labels/Tags**

### Naming Strategy

For multi-policy resources and components:

1. **Container resources** (listeners, clusters, HTTPConnectionManager, RouteConfiguration):
   - Use **service KRI** (for outbounds) or **contextual format** (for inbounds/local resources)
   - These provide aggregate metrics across all contributing policies
   - Names are stable and low-cardinality

2. **Individual policy-contributed components** (route entries, per-route config):
   - Use **policy KRI** with component identifier (e.g., `_rule_0`)
   - These enable policy-level attribution in access logs and tracing
   - Already implemented for routes

3. **Stats interpretation**:
   - **Aggregate stats** (`http.<stat_prefix>.*`, `cluster.<name>.*`): Represent combined effect of all policies
   - **Attributed stats** (access logs, traces, per-route metrics): Identify specific policies
   - Use the right tool for the right question

### Observability Strategy

**For aggregate behavior** (overall success rate, total requests, connection counts):

- Use aggregate metrics from listeners, clusters, HTTPConnectionManager
- Query by service KRI or contextual name

**For policy-specific behavior** (which route matched, which policy caused timeout):

- Use access logs with route names (include policy KRI)
- Use distributed tracing
- Use per-route/per-filter metrics where available

**For policy debugging** (which MeshTimeout policy is active):

- Use Kuma Inspect API
- Query policy resources directly
- Check access logs for route-level attribution

### Documentation Requirements

Must document for each policy type:
- Whether it creates aggregate or attributed resources
- Which metrics are aggregate vs. policy-specific
- How to troubleshoot policy-specific issues
- Examples of observability queries

### Examples

**Good question for aggregate metrics:**
- "What's the overall success rate for traffic to backend-api?"
- "How many active connections does backend-api have?"
→ Use `http.kri_msvc_default_zone-1_demo-app_backend-api_httpport.*`

**Good question for attributed metrics:**
- "Which MeshHTTPRoute policy is causing 404s?"
- "What percentage of requests match the /api route?"
→ Use access logs, filter by route name `kri_mhttpr_..._policy_name_rule_0`

**Good question for Inspect API:**
- "Which MeshTimeout policies apply to this proxy?"
- "What are the effective timeout values?"
→ Use `/meshes/{mesh}/dataplanes/{name}/policies`

## Consequences

### Positive

- **Stable names**: Don't change with policy modifications
- **Low cardinality**: No metric explosion
- **Practical traceability**: Use existing observability tools appropriately
- **Consistent**: Aligns with Envoy's native behavior
- **Extensible**: Works for future policy types

### Negative

- **Multi-tool requirement**: Need metrics + logs + tracing for complete picture
- **Learning curve**: Users must understand aggregate vs. attributed metrics
- **Documentation burden**: Must clearly explain which metrics mean what

### Neutral

- **No naming changes needed**: Current behavior is already consistent with this approach
- **Routes already work this way**: Extends existing pattern

## Implementation Guidance

### For Policy Developers

When implementing new policies that may merge:

1. **Determine resource type**:
   - Creates new named component? → Use policy KRI
   - Merges into existing component? → Use service KRI or contextual

2. **Emit appropriate metrics**:
   - Per-policy metrics where possible (like per-route stats)
   - Document which stats are aggregate

3. **Enable observability**:
   - Include policy KRI in access logs
   - Add policy metadata to tracing spans
   - Consider per-policy metric dimensions where practical

### For Users

When troubleshooting multi-policy scenarios:

1. **Start with aggregate metrics**: Identify that there's a problem
2. **Use access logs**: Identify which routes/policies are involved
3. **Use Inspect API**: Verify policy configuration
4. **Use tracing**: Understand end-to-end policy effects

## Migration Considerations

**No breaking changes required** - this formalizes existing behavior.

**Documentation updates needed**:
- Policy reference docs: Explain aggregate vs. attributed metrics
- Observability guide: Show how to use logs/tracing for policy attribution
- Metrics reference: Document which metrics are aggregate

**Tooling updates needed**:
- Kuma GUI: Show policy attribution in access logs view
- Dashboards: Provide templates for both aggregate and attributed queries
- Alerts: Include examples using both metric types
