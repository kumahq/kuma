# Policy Matching on MeshScoped Zone Proxy

* Status: accepted

Technical Story: https://github.com/Kong/kong-mesh/issues/9029

## Context and Problem Statement

Zone Ingress and Zone Egress were originally dedicated global-scoped resource types.
Being global-scoped meant they lived outside any mesh context, which created two fundamental blockers:

1. **No mesh identity** — `ZoneEgress` could not participate in mTLS with a
   proper SPIFFE identity. The trust domain is mesh-scoped, so a global resource has no mesh
   to derive an identity from. As a result, source-identity-based access control
   (MeshTrafficPermission) could not be enforced at zone proxy boundaries.

2. **No policy support** — the policy system operates on `Dataplane` resources within a mesh.
   Because `ZoneIngress`/`ZoneEgress` were not `Dataplane` resources, none of the mesh policies:
   observability (MeshAccessLog, MeshMetric, MeshTrace), security (MeshTrafficPermission), or
   traffic management (MeshRateLimit, MeshFaultInjection) — could target them.

[MADR-095](095-mesh-scoped-zone-ingress-egress.md) resolved this by modelling zone proxies as
mesh-scoped `Dataplane` resources with a `listeners` array. With that foundation in place, this
MADR establishes the **unified policy model** for zone proxies - how inbound and outbound policies
are structured, how zone proxy Dataplanes are targeted, and how policies can select a specific
MeshExternalService destination on zone egress. It resolves all policy placement items deferred
by [MADR-062](062-meshexternalservice-and-zoneegress.md).

## User Stories

* As a mesh operator I want to give access to service owners to a specific external resource
  (e.g. AWS Aurora) or to a single HTTP endpoint of that resource so that the system follows
  the least privilege principle.
* As a mesh operator I want to know which workload accessed which external resource so that I
  can audit access post-incident.
* As a mesh operator I want to have observability available on zone proxies so that I can
  troubleshoot issues and monitor performance.
* As a mesh operator I want to be able to override Envoy connection and traffic parameters
  (connectTimeout, maxConnections, etc.) so that I can always provide values suited for my traffic.
* As a mesh operator I want to rate limit requests to an external service so that clients don't
  go over service limits and exhaust the budget.
* As a mesh operator I want to inject HTTP headers with a token on the egress for all outgoing
  requests to an external service so that all clients in the mesh can use the same token without
  granting access to the token to individual clients.

## Design

### Scope of Zone Egress

Zone egress is **not** a general-purpose L7 gateway.
It is a transit proxy for outbound MeshExternalService traffic with policy enforcement.

Zone egress is responsible for:
- Forwarding mTLS-terminated traffic to MeshExternalService endpoints
- Enforcing inbound access control (MeshTrafficPermission) and rate limits per source/destination
- Applying outbound policies (timeouts, circuit breakers, health checks) to external endpoints (see policy matrix for prioritization — some are deferred)
Zone egress is NOT responsible for:
- Intra-mesh traffic routing (that is the sidecar's job)
- Acting as a shared API gateway with complex routing logic

### Targeting

Zone proxy dataplanes are targeted using the existing mechanisms from [MADR-095](095-mesh-scoped-zone-ingress-egress.md):

- `sectionName` selects a specific zone proxy listener on a Dataplane.
  Use when a Dataplane mixes zone proxy listeners with regular inbound listeners.
- Labels (`k8s.kuma.io/zone-proxy-type: egress | ingress`) select all zone proxies of a type.
  On Kubernetes, `pod_controller` sets this automatically. Universal users set it in their Dataplane template.

The policy plugin determines whether a policy applies to a matched DPP and ignores it otherwise
(same as existing policies — see MADR-095 "Policy Targeting" section).

### Destination Selector in Inbound Rules

On zone egress, every inbound connection from a sidecar carries a **destination**:
the SNI presented in the mTLS handshake identifies the MeshExternalService the sidecar
wants to reach. When zone egress builds per-MeshExternalService filter chains,
this destination is available at the filter chain selection level.

Inbound `rules` can select a specific MeshExternalService destination,
enabling per-destination rate limits, access control, and fault injection on zone egress.

#### Option A: SNI string match in `Match`

Extend the `Match` struct with an `sni` field:

```go
type Match struct {
    SpiffeID *SpiffeIDMatch `json:"spiffeID,omitempty"`
    SNI      *SNIMatch      `json:"sni,omitempty"`
}
```

With the KRI-based SNI format ([MADR-101](101-sni-format-improvements.md)), SNIs follow the
pattern `sni.<type>.<mesh>.<zone>.<namespace>.<name>[.<sectionName>]` and are human-readable,
predictable, and bidirectionally convertible to KRI. `sectionName` is omitted when not applicable.

```yaml
type: MeshTrafficPermission
mesh: default
name: backend-to-aurora
spec:
  targetRef:
    kind: Dataplane
    sectionName: ze-port
  default:
    allow:
      - spiffeID:
          type: Exact
          value: "spiffe://default/ns/backend-ns/sa/backend"
        sni:
          value: "sni.meshexternalservice.default.zone-1.backend-ns.aws-aurora"
```

* Good, because it maps directly to Envoy's filter chain match on `server_names`.
* Good, because it is transparent — what you write is what Envoy matches on.
* Good, because the KRI-based SNI format (MADR-101) makes SNIs human-readable and
  predictable — users can construct them from resource attributes without querying the CP.
* Bad, because SNIs are an internal representation and coupling policy to them creates
  a dependency on the SNI format stability.
* Bad, because users are already accustomed to the selector-based `targetRef` model and
  would need to learn a new, non-intuitive matching mechanism.
* Bad, because resources created on the Global CP do not include a zone component in
  their SNI, which is confusing — users must understand when zone is present vs omitted
  in the SNI string.

#### Option B: `targetRef` in `Match`

Extend the `Match` struct with a `targetRef` field:

```go
type Match struct {
    SpiffeID  *SpiffeIDMatch        `json:"spiffeID,omitempty"`
    TargetRef *common_api.TargetRef `json:"targetRef,omitempty"`
}
```

The CP resolves the `targetRef` to the MeshExternalService's generated SNI(s) at policy
compilation time, building the correct Envoy filter chain match.

```yaml
type: MeshTrafficPermission
mesh: default
name: backend-to-aurora
spec:
  targetRef:
    kind: Dataplane
    sectionName: ze-port
  default:
    allow:
      - spiffeID:
          type: Exact
          value: "spiffe://default/ns/backend-ns/sa/backend"
        targetRef:
          kind: MeshExternalService
          name: aws-aurora
      - spiffeID:
          type: Exact
          value: "spiffe://default/ns/backend-ns/sa/backend"
        targetRef:
          kind: MeshExternalService
          labels:
            k8s.kuma.io/namespace: backend-ns
```

* Good, because users reference the resource by name or labels.
* Good, because it is consistent with how other Kuma policies reference resources.
* Good, because label-based matching provides broad scoping (by namespace, zone, team, etc.) without requiring users to know internal SNI formats.
* Good, because users are already familiar with `targetRef` selectors from existing policies and do not need to learn a new matching mechanism.
* Good, because policies are insensitive to SNI format changes, if the SNI format evolves, existing policies require no updates; the CP re-resolves them.
* Bad, because CP must resolve `targetRef` to SNI(s) during policy compilation.
* Bad, because `targetRef` in `Match` is only meaningful on zone proxy Dataplanes where per-destination filter chains exist. On regular sidecars the field is invalid and must be rejected at admission, meaning the shared `Match` struct carries a field that most consumers can never use.

#### Option C: `rules` with SpiffeID only + `to` with MeshExternalService

Instead of extending `Match` with a destination field, keep `rules` for source matching only
and use the existing `to` structure for destination selection.

```yaml
type: MeshTrafficPermission
mesh: default
name: backend-to-ns
spec:
  targetRef:
    kind: Dataplane
    sectionName: ze-port
  rules:
    - matches:
        - spiffeID:
            type: Exact
            value: "spiffe://default/ns/backend-ns/sa/backend"
  to:
    - targetRef:
        kind: MeshExternalService
        labels:
          k8s.kuma.io/namespace: backend-ns
```

* Good, because it reuses the existing `to` structure already familiar to users.
* Good, because `rules` stays clean — source matching only, no destination mixed in.
* Good, because label-based `to` provides broad scoping without SNI knowledge.
* Bad, because `MeshTrafficPermission` currently does not have a `to` section;
  adding one is a non-trivial API change and may conflict with MADR-081 semantics.
* Bad, because the semantics of `rules` (inbound match) and `to` (outbound target) are
  orthogonal — combining them in one policy for an inbound-only proxy (zone egress inbound)
  is conceptually confusing.

#### Decision

Implement `targetRef` only in the first iteration:

```go
type Match struct {
    SpiffeID  *SpiffeIDMatch        `json:"spiffeID,omitempty"`
    TargetRef *common_api.TargetRef `json:"targetRef,omitempty"`
}
```

- `targetRef` is the primary, user-facing mechanism. The CP resolves it to SNI(s) at
  compile time. For zone egress, only `kind: MeshExternalService` is valid.
  For zone ingress, only `kind: MeshService` and `kind: MeshMultiZoneService`.
- `targetRef` is optional. When absent, the match applies to all destinations (same as today).
  When present, the match applies only to the filter chain for the referenced resource.
- When a `Match` contains both `spiffeID` and `targetRef`, both conditions must hold.
- `targetRef` in `Match` is valid only when the policy targets a zone proxy Dataplane.

`sni` is deferred to a follow-up. The `Match` struct can be extended with an `sni` field
without breaking existing policies.


### Inbound Policy Examples

#### Access control: allow specific service to reach specific external resource

```yaml
type: MeshTrafficPermission
mesh: default
name: backend-to-aurora
spec:
  targetRef:
    kind: Dataplane
    sectionName: ze-port
  default:
    allow:
      - spiffeID:
          type: Exact
          value: "spiffe://default/ns/backend-ns/sa/backend"
        targetRef:
          kind: MeshExternalService
          name: aws-aurora
```

No other service can reach `aws-aurora` through zone egress because no other `allow` entry matches.

#### Access control: deny specific source from all external resources

Combines SpiffeID and destination matching (AND semantics). Only `spiffeID` + `targetRef` in the same entry are AND-ed; separate entries are OR-ed.

```yaml
type: MeshTrafficPermission
mesh: default
name: deny-compromised-worker-to-aurora
spec:
  targetRef:
    kind: Dataplane
    sectionName: ze-port
  default:
    deny:
      - spiffeID:
          type: Exact
          value: "spiffe://default/ns/backend-ns/sa/compromised-worker"
        targetRef:
          kind: MeshExternalService
          name: aws-aurora
```

### Outbound Policy Examples

Outbound policies (`to`) on zone egress target the MeshExternalService directly.

#### Timeout override

```yaml
type: MeshTimeout
mesh: default
name: aurora-timeout
spec:
  targetRef:
    kind: Dataplane
    sectionName: ze-port
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

#### Inbound match combined with outbound label selector

`rules` and `to` can be combined in the same policy. The `rules` inbound match selects
the filter chain by destination; `to` targets the upstream cluster by `targetRef` (name or labels).

```yaml
type: MeshTimeout
mesh: default
name: aurora-combined
spec:
  targetRef:
    kind: Dataplane
    sectionName: ze-port
  rules:
    - matches:
        - targetRef:
            kind: MeshExternalService
            name: aws-aurora
      default:
        connectionTimeout: 10s
  to:
    - targetRef:
        kind: MeshExternalService
        labels:
          kuma.io/display-name: foo
          k8s.kuma.io/namespace: bar
      default:
        connectionTimeout: 30s
        http:
          requestTimeout: 60s
```

### Updated Policy Matrix

Priority column indicates planned release milestone.

| Policy                    | Zone Egress               | Zone Ingress          | Priority      | Notes |
|---------------------------|---------------------------|-----------------------|---------------|-------|
| MeshTrafficPermission     | Yes (SpiffeID + targetRef) | No                   | **2.14**      | Required for MeshExternalService access control to be complete |
| MeshAccessLog             | Yes                       | Yes                   | **2.14**      | |
| MeshMetric                | Yes                       | Yes                   | **2.14**      | |
| MeshTrace                 | Yes                       | No                    | **2.14**      | Ingress is TCP-only — tracing not meaningful |
| MeshProxyPatch            | Yes                       | Yes                   | 3.0           | Independent of SNI/MES designs; customer demand in ask-mesh |
| MeshTLS                   | Yes                       | Yes                   | 3.0           | |
| MeshRateLimit             | Maybe                     | No                    | Maybe         | |
| MeshTimeout               | Maybe                     | No                    | Maybe         | |
| MeshFaultInjection        | Maybe                     | No                    | Maybe         | |
| MeshCircuitBreaker        | Maybe                     | No                    | Maybe         | |
| MeshHealthCheck           | Maybe                     | No                    | Maybe         | |
| MeshLoadBalancingStrategy | No                        | No                    | Probably not  | |
| MeshRetry                 | No                        | No                    | No            | Squared retries on both sides |
| MeshHTTPRoute             | No                        | No                    | No            | |
| MeshTCPRoute              | No                        | No                    | No            | |

## Security implications and review

### MeshTrafficPermission Default Behaviour

Zone egress is a security boundary — it is the sole exit path for MeshExternalService traffic.
The default behaviour when no MeshTrafficPermission targets zone egress must be **deny-all**.
This MADR proposes changing `mesh.spec.routing.defaultForbidMeshExternalServiceAccess` from its
current default of `false` (permissive, set in MADR-062 when egress targeting was not yet available)
to `true` (deny-all). Now that MeshTrafficPermission can target zone egress with per-destination
granularity, the fail-closed default is both safe and practical. This requires an explicit
implementation change to the API default value.

Operators must create explicit `allow` entries in MeshTrafficPermission to grant access.

### Audit Trail

MeshAccessLog on zone egress inbound captures the source SPIFFE identity and the original
destination SNI, providing an audit trail of which workload accessed which external resource.

## Implications for Kong Mesh

Kong Mesh ships its own rate-limiting and header-injection policies. When zone egress gains
inbound `rules` support and HTTP outbound policies, Kong Mesh policy equivalents should be
validated against the new filter chain architecture.

## Decision

1. **Policy structure**:
   - `MeshTrafficPermission` uses `spec.default.allow/deny/allowWithShadowDeny` with SpiffeID matches.
   - Outbound policies use the standard `to` structure.

2. **Destination selector in inbound rules**: Extend the `Match` struct with an optional
   `targetRef` field (first iteration). `targetRef` is the primary, user-facing mechanism —
   the CP resolves it to SNI(s) at compile time. `sni` direct matching is deferred; the
   struct can be extended in a follow-up without breaking existing policies.

3. **Default RBAC behaviour**: `defaultForbidMeshExternalServiceAccess` defaults to `true`
   (deny-all). This supersedes MADR-062's permissive default, which was set before egress
   targeting was available.

4. **Resolution-failure semantics**: Unresolvable `targetRef` in security policies fails closed
   (match-none). In traffic policies, unresolvable matches are silently dropped.

5. **SNI tombstone on delete**: Deleted MeshExternalService SNIs must not be reused until all
   zone proxies have converged on the new snapshot.

## Notes

* `targetRef` in `Match` is valid only when the policy targets a zone proxy Dataplane.
* `sni` in `Match` is deferred to a follow-up; the struct can be extended without breaking
  existing policies.
* Zone ingress inbound `rules` with `targetRef: kind: MeshService` is only required for `MeshAccessLog`, because it is the only zone ingress policy that needs per-destination log filtering; other zone ingress policies (MeshMetric, MeshTrace) apply uniformly to all inbound traffic without destination discrimination.
* MeshRetry on zone egress outbound remains out of scope — squared retry amplification.
* `MeshTrafficPermission` places `targetRef` in allow/deny entries (alongside `spiffeID`)
  rather than in `rules[].matches[]` — this is a consequence of the MADR-081 API design and
  must be documented prominently for contributors.


## Resources

[MADR-095](095-mesh-scoped-zone-ingress-egress.md) resolved this by modelling zone proxies as
mesh-scoped `Dataplane` resources with a `listeners` array.
[MADR-062](062-meshexternalservice-and-zoneegress.md) established which `to` policies apply
on zone egress for MeshExternalService traffic but deferred egress targeting and inbound policies.
[MADR-101](101-sni-format-improvements.md) introduced a new KRI-based SNI format that makes
SNIs human-readable and bidirectionally convertible to KRI.