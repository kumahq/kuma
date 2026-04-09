# Policy Matching on MeshScoped Zone Proxy

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/8417

## Context and Problem Statement

With mesh-scoped zone proxies ([MADR-095](095-mesh-scoped-zone-ingress-egress.md)),
Zone Ingress and Zone Egress are now regular `Dataplane` resources with a `listeners` array.
The previous decision on policy placement ([MADR-062](062-meshexternalservice-and-zoneegress.md))
deferred applying inbound policies to zone egress (MeshTrafficPermission, MeshRateLimit, MeshFaultInjection)
to a follow-up issue, and left observability and HTTP-level manipulation on zone egress unaddressed.

This MADR establishes the **general policy model** for zone proxies — how inbound and outbound
policies are structured, how zone proxy Dataplanes are targeted, and how policies can select a
specific MeshExternalService destination on zone egress.

**User Stories**

* As a mesh operator I want to give access to service owners to consume external resources (e.g. AWS Aurora)
  so that the system follows the least privilege principle.
* As a mesh operator I want to have observability available on zone proxies so that I can troubleshoot
  issues and monitor performance.
* As a mesh operator I want to be able to override Envoy limits (connectTimeout, maxConnections, maxRetries, etc.)
  so that I can always provide values suited for my traffic.
* As a mesh operator I want to rate limit requests to an external service so that clients don't go over
  service limits and exhaust the budget.
* As a mesh operator I want to inject HTTP headers with a token on the egress for all outgoing requests
  to an external service so that all clients in the mesh can use the same token without granting access
  to the token to individual clients.
* As a mesh operator I want to give access to service owners to a single HTTP endpoint of an external
  resource so that the system follows the least privilege principle.

## Design

### General Policy Model for Zone Proxy

Zone proxy is a regular `Dataplane`. No new resource kind is introduced.
The same policy system used for sidecars applies — the only new capability needed is a way to
select a specific MeshExternalService destination on zone egress inbound rules.

**Inbound policies** (traffic arriving at zone proxy from sidecars or remote zones):
- Use the `rules` structure ([MADR-081](081-inbound-policies-matches.md), [MADR-069](069-inbound-policies.md))
- `MeshTrafficPermission` uses `spec.default.allow/deny/allowWithShadowDeny` with SpiffeID matches
- All other inbound policies (`MeshRateLimit`, `MeshFaultInjection`, `MeshAccessLog`, etc.) use
  `spec.rules[].matches[].default`
- The legacy `from` section is not used for zone proxy policies

**Outbound policies** (traffic leaving zone proxy toward MeshExternalService or local services):
- Use the standard `to` structure with a `targetRef` to `MeshExternalService` or `MeshService`
- This is unchanged from MADR-062 for existing policies

### Targeting Zone Proxy Dataplanes

#### Option A: `sectionName` on top-level targetRef

```yaml
spec:
  targetRef:
    kind: Dataplane
    sectionName: ze-port
```

Targets all Dataplanes that have a zone egress listener named `ze-port`.

* Good, because it reuses the existing `sectionName` mechanism from MADR-095.
* Bad, because the listener name is user-defined and may differ between deployments,
  making it hard to write portable mesh-wide operator policies.

#### Option B: Standardised labels on zone proxy Dataplanes

`pod_controller` automatically propagates `k8s.kuma.io/zone-proxy-type: egress | ingress`
from the Service to the generated Dataplane resource.
This allows label-based targeting:

```yaml
spec:
  targetRef:
    kind: Dataplane
    labels:
      k8s.kuma.io/zone-proxy-type: egress
```

* Good, because policies don't depend on user-chosen listener names.
* Good, because the same label approach works on Universal (operators set the label manually).

#### Decision

Both mechanisms are valid and complementary:

- Use **Option B** (`labels`) as the primary mechanism for mesh-wide policies targeting all
  zone egress or zone ingress proxies.
- Use **Option A** (`sectionName`) when a Dataplane mixes zone proxy listeners with regular
  inbound listeners and the policy must apply only to the zone proxy listener.

On Kubernetes, `pod_controller` sets `k8s.kuma.io/zone-proxy-type: egress` automatically.
Universal users must set the label in their Dataplane template.

### Destination Selector in Inbound Rules

This is the core new concept introduced by this MADR.

On zone egress, every inbound connection from a sidecar carries a **destination**:
the SNI presented in the mTLS handshake identifies the MeshExternalService the sidecar
wants to reach. When zone egress builds per-MeshExternalService filter chains
(required for HTTP-level policies, see outbound section), this destination is available
at the filter chain selection level.

Inbound `rules` can therefore select a specific MeshExternalService by adding a destination
selector to the match. This enables per-destination rate limits, per-destination access control,
and per-destination fault injection — all expressed as inbound policy rules on zone egress.

#### Option A: SNI string match in `Match`

Extend the `Match` struct with an `sni` field:

```go
type Match struct {
    SpiffeID *SpiffeIDMatch `json:"spiffeID,omitempty"`
    SNI      *SNIMatch      `json:"sni,omitempty"`
}
```

Usage:

```yaml
rules:
  - matches:
      - sni:
          type: Prefix
          value: "aws-aurora"
    default:
      local:
        http:
          requestRate:
            num: 100
            interval: 1m
```

* Good, because it maps directly to Envoy's filter chain match on `server_names`.
* Bad, because users must know the generated hostname for each MeshExternalService;
  this hostname is derived from `HostnameGenerator` and can change.
* Bad, because string matching leaks infrastructure details into the policy spec.

#### Option B: MeshExternalService `targetRef` in `Match`

Extend the `Match` struct with a `targetRef` field:

```go
type Match struct {
    SpiffeID  *SpiffeIDMatch   `json:"spiffeID,omitempty"`
    TargetRef *common_api.TargetRef `json:"targetRef,omitempty"`
}
```

The CP resolves the `targetRef` to the MeshExternalService's generated SNI(s) at policy
compilation time, building the correct Envoy filter chain match.

Usage in `MeshTrafficPermission`:

```yaml
type: MeshTrafficPermission
mesh: default
name: backend-to-aurora-only
spec:
  targetRef:
    kind: Dataplane
    labels:
      k8s.kuma.io/zone-proxy-type: egress
  default:
    allow:
      - spiffeId:
          type: Exact
          value: "spiffe://default/ns/backend-ns/sa/backend"
        targetRef:
          kind: MeshExternalService
          name: aws-aurora
```

Usage in other inbound policies (`MeshRateLimit`):

```yaml
type: MeshRateLimit
mesh: default
name: aurora-inbound-rate-limit
spec:
  targetRef:
    kind: Dataplane
    labels:
      k8s.kuma.io/zone-proxy-type: egress
  rules:
    - matches:
        - targetRef:
            kind: MeshExternalService
            name: aws-aurora
      default:
        local:
          http:
            requestRate:
              num: 100
              interval: 1m
```

* Good, because users reference the resource by name; hostname changes are transparent.
* Good, because it is consistent with how other Kuma policies reference resources.
* Good, because the same `Match` extension supports future resource kinds
  (e.g. `MeshService` for zone ingress inbound rules).
* Bad, because CP must resolve `targetRef` to SNI(s) during policy compilation.

#### Decision

Use **Option B** (`targetRef` in `Match`).

The `targetRef` field in `Match` is optional. When absent, the match applies to all
destinations (same as today for regular sidecars). When present, the match applies only
to the filter chain for the referenced resource.

For zone egress, only `kind: MeshExternalService` is valid in `Match.targetRef`.
For zone ingress, only `kind: MeshService` is valid (future).
The CP validates this during policy admission.

When a `Match` contains both `spiffeId` and `targetRef`, both conditions must hold (AND).
When a `Match` contains only `targetRef`, it matches any source connecting to that destination.

### Inbound Policy Examples

#### Access control: allow specific service to reach specific external resource

```yaml
type: MeshTrafficPermission
mesh: default
name: backend-to-aurora
spec:
  targetRef:
    kind: Dataplane
    labels:
      k8s.kuma.io/zone-proxy-type: egress
  default:
    allow:
      - spiffeId:
          type: Exact
          value: "spiffe://default/ns/backend-ns/sa/backend"
        targetRef:
          kind: MeshExternalService
          name: aws-aurora
```

No other service can reach `aws-aurora` through zone egress because no other `allow` entry matches.

#### Access control: global deny-all with per-service exceptions

Enable `mesh.spec.routing.defaultForbidMeshExternalServiceAccess: true`, then:

```yaml
type: MeshTrafficPermission
mesh: default
name: allow-observability-to-all
spec:
  targetRef:
    kind: Dataplane
    labels:
      k8s.kuma.io/zone-proxy-type: egress
  default:
    allow:
      - spiffeId:
          type: Prefix
          value: "spiffe://default/ns/observability/"
```

```yaml
type: MeshTrafficPermission
mesh: default
name: block-compromised-worker
spec:
  targetRef:
    kind: Dataplane
    labels:
      k8s.kuma.io/zone-proxy-type: egress
  default:
    deny:
      - spiffeId:
          type: Exact
          value: "spiffe://default/ns/backend-ns/sa/compromised-worker"
```

Multiple MeshTrafficPermission policies targeting zone egress have their `rules` concatenated
and evaluated together: `deny` takes priority over `allow` across all policies (MADR-081 algorithm).

#### Rate limit per destination

```yaml
type: MeshRateLimit
mesh: default
name: aurora-rate-limit
spec:
  targetRef:
    kind: Dataplane
    labels:
      k8s.kuma.io/zone-proxy-type: egress
  rules:
    - matches:
        - targetRef:
            kind: MeshExternalService
            name: aws-aurora
      default:
        local:
          http:
            requestRate:
              num: 200
              interval: 1m
```

Without `targetRef` in the match, the rate limit applies to all traffic arriving at zone egress,
regardless of destination.

#### Rate limit per source and destination

Combines SpiffeID and destination matching (AND semantics):

```yaml
type: MeshRateLimit
mesh: default
name: backend-aurora-rate-limit
spec:
  targetRef:
    kind: Dataplane
    labels:
      k8s.kuma.io/zone-proxy-type: egress
  rules:
    - matches:
        - spiffeId:
            type: Exact
            value: "spiffe://default/ns/backend-ns/sa/backend"
          targetRef:
            kind: MeshExternalService
            name: aws-aurora
      default:
        local:
          http:
            requestRate:
              num: 50
              interval: 1m
```

#### Fault injection on zone egress inbound

```yaml
type: MeshFaultInjection
mesh: default
name: egress-fault
spec:
  targetRef:
    kind: Dataplane
    labels:
      k8s.kuma.io/zone-proxy-type: egress
  rules:
    - matches:
        - targetRef:
            kind: MeshExternalService
            name: aws-aurora
      default:
        http:
          abort:
            httpStatus: 503
            percentage: 10
```

#### Observability on zone egress inbound

```yaml
type: MeshAccessLog
mesh: default
name: egress-inbound-access-log
spec:
  targetRef:
    kind: Dataplane
    labels:
      k8s.kuma.io/zone-proxy-type: egress
  rules:
    - default:
        backends:
          - type: OpenTelemetry
            openTelemetry:
              endpoint: otel-collector.observability:4317
```

MeshMetric and MeshTrace follow the same pattern. No API changes to these policies are needed.

### Outbound Policy Examples

Outbound policies (`to`) on zone egress target the MeshExternalService directly.
These extend the set already defined in MADR-062 (MeshCircuitBreaker, MeshHealthCheck,
MeshLoadBalancingStrategy).

#### HTTP filter chains on zone egress

For HTTP-level outbound policies (MeshHTTPRoute, outbound MeshRateLimit) to work on zone egress,
zone egress must build an **HTTP Connection Manager (HCM) filter chain** for the MeshExternalService.
Zone egress only creates an HCM chain when at least one HTTP-level policy targets that
MeshExternalService; all other destinations keep the lightweight TCP proxy filter chain.

The SNI from the inbound mTLS connection selects the filter chain, and the HCM in that chain
handles HTTP-level processing before forwarding to the external endpoint.

#### Header injection

```yaml
type: MeshHTTPRoute
mesh: default
name: inject-auth-header
spec:
  targetRef:
    kind: Dataplane
    labels:
      k8s.kuma.io/zone-proxy-type: egress
  to:
    - targetRef:
        kind: MeshExternalService
        name: aws-aurora
      default:
        rules:
          - matches:
              - path:
                  type: PathPrefix
                  value: /
            default:
              filters:
                - type: RequestHeaderModifier
                  requestHeaderModifier:
                    add:
                      - name: Authorization
                        value: Bearer $(token)
```

The token value is delivered via Envoy SDS (`generic_secret`) — see Security section.

#### Path-level access restriction

Combined with MeshTrafficPermission granting access to zone egress, a MeshHTTPRoute that only
matches a specific path restricts the service to that single endpoint of the external resource:

```yaml
type: MeshHTTPRoute
mesh: default
name: restrict-to-single-endpoint
spec:
  targetRef:
    kind: Dataplane
    labels:
      k8s.kuma.io/zone-proxy-type: egress
  to:
    - targetRef:
        kind: MeshExternalService
        name: aws-aurora
      default:
        rules:
          - matches:
              - path:
                  type: Exact
                  value: /api/v1/data
            default:
              backendRefs:
                - kind: MeshExternalService
                  name: aws-aurora
```

Requests to any other path are dropped because no route matches.

#### Timeout override

```yaml
type: MeshTimeout
mesh: default
name: aurora-timeout
spec:
  targetRef:
    kind: Dataplane
    labels:
      k8s.kuma.io/zone-proxy-type: egress
  to:
    - targetRef:
        kind: MeshExternalService
        name: aws-aurora
      default:
        connectionTimeout: 30s
        http:
          requestTimeout: 60s
          idleTimeout: 600s
```

#### Outbound rate limit (quota protection)

```yaml
type: MeshRateLimit
mesh: default
name: aurora-outbound-quota
spec:
  targetRef:
    kind: Dataplane
    labels:
      k8s.kuma.io/zone-proxy-type: egress
  to:
    - targetRef:
        kind: MeshExternalService
        name: aws-aurora
      default:
        local:
          http:
            requestRate:
              num: 200
              interval: 1m
```

#### Outbound observability

```yaml
type: MeshAccessLog
mesh: default
name: egress-outbound-access-log
spec:
  targetRef:
    kind: Dataplane
    labels:
      k8s.kuma.io/zone-proxy-type: egress
  to:
    - targetRef:
        kind: MeshExternalService
        name: aws-aurora
      default:
        backends:
          - type: OpenTelemetry
            openTelemetry:
              endpoint: otel-collector.observability:4317
```

### Updated Policy Matrix

This table extends the matrix from MADR-062.
"Inbound `rules`" refers to the `spec.rules[]` structure (or `spec.default` for MeshTrafficPermission).
Rows marked **new** were previously "Out of scope (Egress targeting)".

| Policy                    | Zone Egress inbound (`rules`)              | Zone Egress outbound (`to`)         | Notes |
|---------------------------|--------------------------------------------|-------------------------------------|-------|
| MeshAccessLog             | Yes **new**                                | Yes **new**                         | |
| MeshCircuitBreaker        | No                                         | Yes (existing, MADR-062)            | |
| MeshFaultInjection        | Yes **new** (supports `targetRef` match)   | No                                  | |
| MeshHealthCheck           | No                                         | Yes (existing, MADR-062)            | |
| MeshHTTPRoute             | No                                         | Yes **new** (requires HCM chain)    | Header injection, path restriction |
| MeshLoadBalancingStrategy | No                                         | Yes (existing, MADR-062)            | |
| MeshMetric                | Yes **new**                                | Yes **new**                         | |
| MeshProxyPatch            | No                                         | Yes **new**                         | Last-resort Envoy config override |
| MeshRateLimit             | Yes **new** (supports `targetRef` match)   | Yes **new**                         | Inbound: per-source/destination; Outbound: global quota |
| MeshRetry                 | No                                         | Sidecar only (MADR-062 decision)    | Squared retries on both |
| MeshTCPRoute              | No                                         | No                                  | Use MeshHTTPRoute for HTTP destinations |
| MeshTimeout               | No                                         | Yes **new**                         | |
| MeshTrace                 | Yes **new**                                | Yes **new**                         | |
| MeshTrafficPermission     | Yes **new** (SpiffeID + `targetRef` match) | N/A                                 | `spec.default` API, not `from` |

### Zone Ingress Policy Matrix

Zone ingress carries cross-zone traffic. The `targetRef` match in `rules` would reference
`MeshService` (not `MeshExternalService`) to select the destination local service.
This is left for a follow-up.

| Policy                    | Zone Ingress inbound (`rules`)         | Zone Ingress outbound (`to`) | Notes |
|---------------------------|----------------------------------------|------------------------------|-------|
| MeshAccessLog             | Yes                                    | Yes                          | |
| MeshFaultInjection        | Yes                                    | No                           | |
| MeshMetric                | Yes                                    | Yes                          | |
| MeshProxyPatch            | No                                     | Yes                          | |
| MeshRateLimit             | Yes                                    | No                           | |
| MeshTimeout               | No                                     | Yes                          | |
| MeshTrace                 | Yes                                    | Yes                          | |
| MeshTrafficPermission     | Yes (SpiffeID + `targetRef` match)     | N/A                          | |

## Security Implications and Review

### Token / Credential Injection

When zone egress injects credentials (Authorization header, API keys) on behalf of services:

1. The token must not be stored in the Dataplane spec or in a policy resource readable by service owners.
   Use Envoy SDS (`generic_secret`) to deliver secrets out-of-band.
2. In Kubernetes, the secret is stored in a `Secret` resource with RBAC allowing only
   the zone egress pod's service account to read it.
3. MeshHTTPRoute `RequestHeaderModifier` references the secret by name via the `secretRef` field
   (requires Envoy HCM with SDS enabled).

### MeshTrafficPermission Default Behaviour

When mTLS is enabled and no MeshTrafficPermission targets zone egress, behaviour is controlled
by `mesh.spec.routing.defaultForbidMeshExternalServiceAccess`:

- `false` (default): all services may connect (permissive).
- `true`: no service may connect without an explicit `allow` rule.

Least-privilege deployments should set `defaultForbidMeshExternalServiceAccess: true` and
create explicit allow policies per service (optionally scoped to specific MeshExternalServices
via `targetRef` in the match).

### Audit Trail

MeshAccessLog on zone egress inbound captures the source SPIFFE identity and the original
destination SNI, providing an audit trail of which workload accessed which external resource.

## Reliability Implications

### HTTP Filter Chain per MeshExternalService

When zone egress builds an HCM chain per destination, the total number of filter chains grows
with the number of MeshExternalService resources that have HTTP-level outbound policies.
For large deployments, this increases zone egress memory and CPU usage.

Mitigation: zone egress only creates an HCM chain when at least one HTTP-level policy
(MeshHTTPRoute, outbound MeshRateLimit `to`) targets that MeshExternalService.
Destinations without such policies keep the lightweight TCP proxy filter chain.

### Rate Limit Interaction

Inbound rate limit (per-source or per-destination) and outbound rate limit (per-destination
quota) operate independently at different points in zone egress. They can be combined:
inbound per-source limits govern individual client throughput, and outbound limits protect
the external service's global quota.

## Implications for Kong Mesh

Kong Mesh ships its own rate-limiting and header-injection policies. When zone egress gains
inbound `rules` support and HTTP outbound policies, Kong Mesh policy equivalents should be
validated against the new filter chain architecture.

## Decision

1. **Targeting**: Use `k8s.kuma.io/zone-proxy-type: egress | ingress` label on Dataplane
   resources as the primary targeting mechanism for zone proxy policies.
   Use `sectionName` only for mixed Dataplanes.

2. **Policy structure**:
   - Inbound policies on zone proxy use the `rules` structure (MADR-081/069), not `from`.
   - `MeshTrafficPermission` uses `spec.default.allow/deny/allowWithShadowDeny` with SpiffeID matches.
   - Outbound policies use the standard `to` structure.

3. **Destination selector in inbound `rules`**: Extend the `Match` struct with an optional
   `targetRef` field. When set to `kind: MeshExternalService`, the rule applies only to
   connections destined for that resource (resolved to SNI by the CP). When absent, the
   rule applies to all destinations.

4. **HTTP filter chains**: Zone egress creates an HCM filter chain per MeshExternalService
   only when at least one HTTP-level outbound policy targets it.

5. **Token injection security**: Credentials injected via MeshHTTPRoute `RequestHeaderModifier`
   must be delivered through Envoy SDS (`generic_secret`).

6. **Default RBAC behaviour unchanged**: Follows `mesh.spec.routing.defaultForbidMeshExternalServiceAccess`
   from MADR-062.

## Notes

* `targetRef` in `Match` is valid only when the policy targets a zone proxy Dataplane.
  On regular sidecars, the field is rejected at admission time.
* Zone ingress inbound `rules` with `targetRef: kind: MeshService` (to select a specific
  local destination) is deferred to a follow-up.
* MeshRetry on zone egress outbound remains out of scope.
  Applying retries on both sidecar and zone egress creates squared retry amplification.
