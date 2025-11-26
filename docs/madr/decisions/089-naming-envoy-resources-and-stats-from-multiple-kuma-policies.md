# Naming Envoy Resources and Stats from Multiple Kuma Policies

- Status: accepted

Technical Story: <https://github.com/kumahq/kuma/issues/13886>

## Context and Problem Statement

### Background: Unified Naming Strategy

Kuma has established a unified naming strategy for Envoy resources and stats through four foundational MADRs:

- [MADR-070: Resource Identifier](070-resource-identifier.md) - KRI format for user resources
- [MADR-077: Standardized Naming for System xDS Resources](077-naming-system-envoy-resources.md) - `system_*` format
- [MADR-078: Consistent Naming for Non-System Resources](078-migrating-to-consistent-and-well-defined-naming-for-non-system-envoy-resources-and-stats.md) - `self_*` contextual format
- [MADR-084: Finalized Schemes](084-finalized-schemes-for-unified-envoy-resources-and-stats-naming-formats.md) - Consolidated formats with scope tokens

The three complementary naming formats are:

- **KRI** (`kri_<type>_<mesh>_<zone>_<namespace>_<name>_<section>`) - resources mapping to Kuma resources
- **Contextual** (`self_<category>_<scope>_<rest>`) - proxy-local resources (scope: `dp`, `zi`, `ze`)
- **System** (`system_*`) - internal infrastructure resources

### The Problem: Multi-Policy Resources

This strategy works well for one-to-one relationships between Kuma and Envoy resources. However, **some Envoy resources result from multiple Kuma policies merging together**:

- Routes from multiple `MeshHTTPRoute` policies targeting the same service
- RBAC filters combining rules from multiple `MeshTrafficPermission` policies
- Configuration from multiple `MeshTimeout`, `MeshRetry`, or `MeshRateLimit` policies

These multi-policy scenarios were explicitly marked as out of scope in previous MADRs. **This MADR addresses that gap.**

### Why This Matters

**Observability impact**: Without addressing multi-policy naming:

- Cannot identify which policy caused observed behavior from aggregate stats
- Troubleshooting requires correlating multiple data sources
- Dashboard and alert creation becomes complex

**Design constraints**: Solutions must balance traceability vs. cardinality, and stability vs. completeness.

## Scope

**In scope:**

- Envoy resources generating stats where multiple policies merge
- Defining naming approach for multi-policy scenarios
- Observability strategy documentation

**Out of scope:**

- Single-policy naming (already solved)
- Resources without stats/metrics
- Non-xDS resources
- Default routes and MeshPassthrough (separate MADRs)

## Current Implementation Status

### What Already Works

Route naming with unified naming enabled already provides policy attribution:

- **Route entries**: Use policy KRI (`kri_mhttpr_default_zone-1_demo-app_backend-routing_rule_0`)
- **Origin tracking**: `OriginByMatches` map records which policy contributed each route match
- **Access logs**: Include route name for request-level attribution

### What This MADR Addresses

- Formalizes the distinction between aggregate and attributed metrics
- Defines observability strategy combining metrics, logs, and tracing
- Documents which stats aggregate across policies vs. provide attribution

## Envoy Resource Architecture

### Stats-Generating Resources

| Resource              | Stats Pattern                  | Naming Source | Multi-Policy Impact           |
|:----------------------|:-------------------------------|:--------------|:------------------------------|
| Listener              | `listener.<name>.*`            | Service KRI   | Aggregate across all policies |
| Cluster               | `cluster.<name>.*`             | Service KRI   | Aggregate across all policies |
| HTTPConnectionManager | `http.<stat_prefix>.*`         | Service KRI   | Aggregate across all policies |
| RBAC Filter           | `<stat_prefix>.allowed/denied` | Contextual    | Aggregate across all MTP      |
| Route Entry           | Access logs, tracing           | Policy KRI    | ✅ Per-policy attribution     |

### Resource Hierarchy with Policy Merge Points

```text
Listener (service KRI)
└── FilterChain
    ├── HTTPConnectionManager (stat_prefix: service KRI)
    │   ├── RouteConfig
    │   │   └── Routes[] ← MERGE: Multiple MeshHTTPRoute (each route: policy KRI)
    │   │       └── route.timeout ← MERGE: MeshTimeout policies
    │   │       └── route.retry_policy ← MERGE: MeshRetry policies
    │   └── HTTPFilters
    │       └── RBAC ← MERGE: Multiple MeshTrafficPermission (stat_prefix: contextual)
    └── TCPProxy (stat_prefix: service KRI)

Cluster (service KRI)
├── health_checks ← MERGE: MeshHealthCheck policies
└── circuit_breakers ← MERGE: MeshCircuitBreaker policies
```

## Multi-Policy Scenarios

| Scenario                    | Merge Location           | Container Naming | Component Naming     | Stats Attribution               |
|:----------------------------|:-------------------------|:-----------------|:---------------------|:--------------------------------|
| Multi-MeshHTTPRoute         | RouteConfig.routes[]     | Service KRI      | Policy KRI per route | Routes: ✅ HCM stats: aggregate |
| Multi-MeshTrafficPermission | RBAC filter rules        | Contextual       | N/A                  | Aggregate                       |
| Multi-MeshTimeout           | route.timeout            | Service KRI      | Inherits route       | Aggregate at HCM level          |
| Multi-MeshRetry             | route.retry_policy       | Service KRI      | Inherits route       | Aggregate at HCM level          |
| Multi-MeshHealthCheck       | cluster.health_checks[]  | Service KRI      | N/A                  | Aggregate                       |
| Multi-MeshCircuitBreaker    | cluster.circuit_breakers | Service KRI      | N/A                  | Aggregate                       |

### Example: Three MeshHTTPRoute Policies Merged

```yaml
Listener:
  name: kri_msvc_default_zone-1_demo-app_backend-api_httpport
  filter_chains:
    - filters:
        - name: envoy.filters.network.http_connection_manager
          typed_config:
            stat_prefix: kri_msvc_default_zone-1_demo-app_backend-api_httpport  # Aggregate
            route_config:
              routes:
                - name: kri_mhttpr_default_zone-1_demo-app_backend-routing_rule_0      # Policy 1
                - name: kri_mhttpr_default__kuma-system_client-to-backend_rule_0       # Policy 2
                - name: kri_mhttpr_default_zone-1_demo-app_backend-with-timeout_rule_0 # Policy 3
```

**Stats interpretation:**

- `http.kri_msvc_..._backend-api_httpport.downstream_rq_2xx` → Aggregate across all three policies
- Access log `route_name: kri_mhttpr_..._backend-routing_rule_0` → Specific policy attribution

## Design Options

| Option                      | Approach                                              | Verdict                                  |
|:----------------------------|:------------------------------------------------------|:-----------------------------------------|
| Encode all policies         | Include all policy KRIs in resource name              | ❌ Cardinality explosion, unstable names |
| Multi-policy marker         | Add `__multipolicy` suffix                            | ❌ No actual attribution benefit         |
| Status quo                  | Keep current behavior                                 | ❌ Maintains traceability gaps           |
| **Aggregate + Attribution** | Service KRI for containers, policy KRI for components | ✅ Recommended                           |

### Recommended: Aggregate Naming + Policy Attribution

**Approach:**

1. **Container resources** (listeners, clusters, HCM): Use service KRI or contextual format
2. **Component resources** (individual routes): Use policy KRI
3. **Observability**: Combine aggregate metrics with attributed data sources

**Why this works:**

- **Stable names**: Don't change with policy additions/removals
- **Low cardinality**: Aggregate metrics remain aggregated
- **Practical traceability**: Use access logs and tracing for detailed attribution
- **Consistent with Envoy patterns**: Matches how Envoy handles aggregation
- **Already implemented**: Routes already use policy KRIs

## Decision

### Naming Strategy

For multi-policy resources:

1. **Container resources**: Use **service KRI** (outbounds) or **contextual format** (inbounds)
2. **Policy-contributed components**: Use **policy KRI** with component identifier (`_rule_0`)
3. **Stats interpretation**:
   - Aggregate stats (`http.<stat_prefix>.*`) → Combined effect of all policies
   - Attributed data (access logs, traces) → Specific policy identification

### Observability Strategy

| Question Type    | Data Source                        | Example                                                |
|:-----------------|:-----------------------------------|:-------------------------------------------------------|
| Aggregate        | Metrics (cluster, listener, HCM)   | `envoy_cluster_upstream_rq_xx{cluster="kri_msvc_..."}` |
| Policy-specific  | Access logs (filter by route_name) | `route_name =~ "kri_mhttpr_.*policy-name.*"`           |
| Policy debugging | Inspect API                        | `GET /meshes/{mesh}/dataplanes/{name}/policies`        |

## Consequences

**Positive:**

- Stable, low-cardinality names
- Practical traceability using existing tools
- Extensible to future policy types
- Aligns with Envoy's native patterns

**Negative:**

- Requires multiple tools (metrics + logs + tracing) for complete picture
- Learning curve for aggregate vs. attributed metrics

**Neutral:**

- No naming changes required (formalizes current behavior)
- Routes already implement this pattern

## Implementation Guidance

### For Policy Developers

When implementing policies that may merge:

1. If creating new named component → Use policy KRI
2. If merging into existing component → Use service KRI or contextual
3. Include policy KRI in access logs and tracing spans

### For Users

**Aggregate metrics** (overall behavior):

```promql
# Error rate to backend-api service
sum(rate(envoy_cluster_upstream_rq_xx{envoy_cluster_name="kri_msvc_default_zone-1_demo-app_backend-api_httpport"}[5m]))
```

**Policy attribution** (which policy caused behavior):

```text
# Access logs: filter by route name
route_name =~ "kri_mhttpr_default_zone-1_demo-app_backend-routing_.*"
```

**Policy verification** (what's configured):

```bash
# Inspect API
curl /meshes/default/dataplanes/backend-pod/policies
```

## Migration Considerations

**No breaking changes** - this formalizes existing behavior.

**Documentation updates needed:**

- Policy reference docs: Document which metrics are aggregate vs. attributed
- Observability guide: Show how to use logs and tracing for policy attribution
- Metrics reference: Clarify aggregate metric semantics

**Tooling updates:**

- Kuma GUI: Highlight policy attribution in access logs view
- Dashboards: Provide templates for both aggregate and attributed queries
