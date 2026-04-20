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
MeshExternalService destination on zone egress.

## User Stories

1. As a mesh operator I want to give access to service owners to a specific external resource
   (e.g. AWS Aurora) or to a single HTTP endpoint of that resource so that the system follows
   the least privilege principle.
2. As a mesh operator I want to know which workload accessed which external resource so that I
   can audit access post-incident.
3. As a mesh operator I want to have observability available on zone proxies so that I can
   troubleshoot issues and monitor performance.
4. As a mesh operator I want to be able to override Envoy connection and traffic parameters
   (connectTimeout, maxConnections, etc.) so that I can always provide values suited for my traffic.
5. As a mesh operator I want to rate limit requests to an external service so that clients don't
   go over service limits and exhaust the budget.

### Out-of-scope
1. As a mesh operator I want to inject HTTP headers with a token on the egress for all
   outgoing requests to an external service so that all clients in the mesh can use the same
   token without granting access to the token to individual clients.

## Design

### Scope of Zone Proxies

**Zone egress** is **not** a general-purpose L7 gateway.
It is a transit proxy for outbound MeshExternalService traffic with policy enforcement.

Zone egress is responsible for:
- Forwarding mTLS-terminated traffic to MeshExternalService endpoints
- Enforcing inbound access control (MeshTrafficPermission) and rate limits per source/destination
- Applying outbound policies (timeouts, circuit breakers, health checks) to external endpoints (see policy matrix for prioritization — some are deferred)

Zone egress is NOT responsible for:
- Intra-mesh traffic routing (that is the sidecar's job)
- Acting as a shared API gateway with complex routing logic

**Zone ingress** is the entry point for cross-zone traffic destined to services in the local zone.

Zone ingress is responsible for:
- Receiving mTLS connections from remote zone sidecars and forwarding them to local service endpoints
- Applying observability policies (MeshAccessLog, MeshMetric) to cross-zone inbound traffic

Zone ingress is NOT responsible for:
- Access control (connections are already mTLS-authenticated at the sidecar level)
- Outbound traffic (zone egress handles that)

### Targeting

Zone proxy `Dataplane` resources are targeted the same way as any other `Dataplane` — using
`targetRef.kind: Dataplane` with `name`, `labels`, or `sectionName`. No special targeting
mechanism is introduced.

- Use `name` to target a specific zone proxy instance.
- Use `labels` (e.g. `k8s.kuma.io/zone-proxy-type: egress`) to target all zone proxies of a type.
  On Kubernetes these labels are set automatically, Universal users set them in their Dataplane template.
- Use `sectionName` to target a specific listener when a Dataplane mixes zone proxy and regular
  inbound listeners.

### Destination Selector in Inbound Rules

On zone egress, every inbound connection from a sidecar carries a **destination**:
the SNI presented in the mTLS handshake identifies the MeshExternalService the sidecar
wants to reach. When zone egress builds per-MeshExternalService filter chains,
this destination is available at the filter chain selection level.

Inbound `rules` can select a specific MeshExternalService destination,
enabling per-destination configuration on zone egress.

#### Option A: SNI string match in `Match`

Extend the `Match` struct with an `sni` field:

```go
type Match struct {
    SpiffeID *SpiffeIDMatch `json:"spiffeID,omitempty"`
    SNI      *SNIMatch      `json:"sni,omitempty"`
}
```

With the KRI-based SNI format (MADR-101), SNIs follow the
pattern `sni.<type>.<mesh>.<zone>.<namespace>.<name>.<sectionName>` and are human-readable,
predictable, and bidirectionally convertible to KRI.

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
* Bad, because coupling policy to SNI strings creates a dependency on SNI format stability,
  even though MADR-101 makes SNIs human-readable and predictable.
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

The CP resolves the `targetRef` to the MeshExternalService's generated SNI(s) when generating Envoy configuration.

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
* Bad, because CP must resolve `targetRef` to SNI(s) during policy processing (when building Envoy configuration).
* Bad, because `labels` matching can unintentionally select multiple MeshExternalServices, which may not be the user's intent.
* Bad, because `targetRef` in `Match` is only meaningful on zone proxy Dataplanes where per-destination filter chains exist. On regular sidecars the field has no matching filter chain and is silently ignored — same behavior as MeshHTTPRoute on DPPs without the referenced route.

#### Option C: destination in `rules[].targetRef`

Keep `Match` for source matching only and express the destination as a top-level `targetRef`
on the rule itself, separate from the per-match `spiffeID`.

```yaml
type: MeshTrafficPermission
mesh: default
name: backend-to-aurora
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

* Good, because destination and source are cleanly separated at the rule level.
* Good, because `Match` struct stays unchanged — no new fields needed.
* Good, because `rules[].targetRef` is already a concept in other policies.
* Bad, because `MeshTrafficPermission` uses `spec.default` not `rules` today — adding
  `rules[].targetRef` is a non-trivial API change.
* Bad, because `to[].targetRef` with `rules` is not supported.

#### Decision

Implement `sni` in the first iteration:

```go
type Match struct {
    SpiffeID *SpiffeIDMatch `json:"spiffeID,omitempty"`
    SNI      *SNIMatch      `json:"sni,omitempty"`
}
```

- `sni` maps directly to Envoy's `server_names` filter chain match — no CP resolution step.
- `sni` is optional. When absent, the match applies to all destinations (same as today).
  When present, the match applies only to the filter chain whose SNI matches.
- When a `Match` contains both `spiffeID` and `sni`, both conditions must hold.
- With the KRI-based SNI format (MADR-101), SNIs are human-readable and predictable —
  users construct them from resource attributes without querying the CP.
- `sni` in `Match` is valid only when the policy targets a zone proxy Dataplane.

`targetRef` in `Match` may be added as a follow-up once `sni` is proven in production.
The `Match` struct can be extended without breaking existing policies.

### Policy Examples

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
        sni:
          value: "sni.meshexternalservice.default.zone-1.backend-ns.aws-aurora.8443"
```

No other service can reach `aws-aurora` through zone egress because no other `allow` entry matches.

#### Access control: deny specific source from a specific external resource

Combines SpiffeID and destination matching (AND semantics). Only `spiffeID` + `sni` in the same entry are AND-ed; separate entries are OR-ed.

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
        sni:
          value: "sni.meshexternalservice.default.zone-1.backend-ns.aws-aurora.8443"
```

#### MeshTimeout: match via `rules`

`spec.rules[].matches[].sni` selects a specific filter chain on the zone proxy inbound.

```yaml
type: MeshTimeout
mesh: default
name: aurora-timeout
spec:
  targetRef:
    kind: Dataplane
    sectionName: ze-port
  rules:
    - matches:
        - sni:
            value: "sni.meshexternalservice.default.zone-1.backend-ns.aws-aurora.8443"
      default:
        connectionTimeout: 10s
```

### Updated Policy Matrix

Priority column indicates planned release milestone.

| Policy                    | Zone Egress               | Zone Ingress          | Priority      | Notes |
|---------------------------|---------------------------|-----------------------|---------------|-------|
| MeshTrafficPermission     | Yes                       | No                    | **2.14**      | Required for MeshExternalService access control to be complete |
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
The default behavior when no MeshTrafficPermission targets zone egress must be **deny-all**.
This MADR proposes changing `mesh.spec.routing.defaultForbidMeshExternalServiceAccess` from its
current default of `false` (permissive, set in MADR-062 when egress targeting was not yet available)
to `true` (deny-all). Now that MeshTrafficPermission can target zone egress with per-destination
granularity, the fail-closed default is both safe and practical. This requires an explicit
implementation change to the API default value.

Operators must create explicit `allow` entries in MeshTrafficPermission to grant access.

> **Breaking change**: this default flip MUST be documented in `UPGRADE.md` in bold.
> Users relying on the permissive default will have all MeshExternalService access blocked
> after upgrading without any MeshTrafficPermission policy in place.

### Audit Trail

MeshAccessLog on zone egress inbound captures the source SPIFFE identity and the original
destination SNI, providing an audit trail of which workload accessed which external resource.

## Implications for Downstream Projects

- **MeshGlobalRateLimit**: no action required.
- **MeshOPA**: was never supported on zone egress. Explicitly out of scope for this MADR;
  requires a separate decision if needed.

## Decision

1. **Policy structure**:
   - `MeshTrafficPermission` uses `spec.default.allow/deny/allowWithShadowDeny` with SpiffeID matches.

2. **Destination selector in inbound rules**: Extend the `Match` struct with an `sni` field.
   SNI maps directly to Envoy's `server_names` filter chain match — no CP resolution step.
   With the KRI-based SNI format (MADR-101), SNIs are human-readable and predictable.
   `targetRef` in `Match` is deferred to a follow-up.

3. **Default RBAC behaviour**: `defaultForbidMeshExternalServiceAccess` defaults to `true`
   (deny-all). This supersedes MADR-062's permissive default, which was set before egress
   targeting was available.

4. **Non-matching SNI semantics**: If no `allow` entry matches the SNI, access is denied
   (fail-closed). In traffic policies, non-matching `sni` entries are silently skipped.

## Notes

* `targetRef` in `Match` is deferred to a follow-up; the struct can be extended without
  breaking existing policies once `sni` is proven in production.
* Zone ingress inbound `rules` with `sni` matching is only required for `MeshAccessLog` —
  it is the only zone ingress policy that needs per-destination log filtering; other zone
  ingress policies (MeshMetric, MeshTrace) apply uniformly to all inbound traffic.
* MeshRetry on zone egress outbound remains out of scope — squared retry amplification.
* `MeshTrafficPermission` places `sni` in allow/deny entries (alongside `spiffeID`)
  rather than in `rules[].matches[]` — this is a consequence of the MADR-081 API design and
  must be documented prominently for contributors.

## Resources

[MADR-095](095-mesh-scoped-zone-ingress-egress.md) resolved this by modelling zone proxies as
mesh-scoped `Dataplane` resources with a `listeners` array.
[MADR-062](062-meshexternalservice-and-zoneegress.md) established which `to` policies apply
on zone egress for MeshExternalService traffic but deferred egress targeting and inbound policies.
MADR-101 (pending) introduces a new KRI-based SNI format that makes SNIs human-readable and
bidirectionally convertible to KRI.