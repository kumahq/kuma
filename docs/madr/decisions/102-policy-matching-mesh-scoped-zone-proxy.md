# Policy Matching on MeshScoped Zone Proxy

* Status: accepted

Technical Story: https://github.com/Kong/kong-mesh/issues/9029

## Summary

Builds on [MADR-095](095-mesh-scoped-zone-ingress-egress.md), which turned zone proxies into
mesh-scoped `Dataplane` resources, and defines how policies target them. Zone proxy Dataplanes
are selected with the standard `targetRef.kind: Dataplane` (`name`, `labels`, `sectionName`) —
no special targeting. Each policy group uses a different destination selector based on what its
data model can express; the four groups are summarized below with a small example for each.

Ships in 2.14 (no upgrade actions required). 3.0 removes the legacy global-scoped zone proxies
and the `defaultForbidMeshExternalServiceAccess` flag — see UPGRADE.md.

### `MeshTrafficPermission`

Zone egress is **deny-all by default**. Operators grant access with `allow` entries that
constrain source (`spiffeID`) and destination (`sni`) together. A new `sni` field is added to
the existing `Match` struct and maps directly to Envoy's `server_names` filter chain match —
no CP resolution step. Non-matching SNIs fail closed (denied).

```yaml
type: MeshTrafficPermission
spec:
  targetRef:
    kind: Dataplane
    sectionName: ze-port
  rules:
    - default:
        allow:
          - spiffeID:
              type: Exact
              value: spiffe://default/ns/backend-ns/sa/backend
            sni:
              type: Exact
              value: sni.extsvc.default.zone-1.aws-aurora.8443
```

### `MeshCircuitBreaker`, `MeshHealthCheck`

These policies have no `rules[].matches[]` field, so destination selection uses `spec.to[].targetRef`.
Top-level `targetRef` selects the zone proxy (or a listener via `sectionName`); each
`spec.to[]` entry selects the destination `filterChain` whose SNI matches the referenced
`MeshExternalService` / `MeshService`.

```yaml
type: MeshCircuitBreaker
spec:
  targetRef:
    kind: Dataplane
    sectionName: ze-port
  to:
    - targetRef:
        kind: MeshExternalService
        labels:
          kuma.io/display-name: httpbin
      default:
        # ...
```

This replaces the legacy "target the client, configure egress indirectly" pattern, which
violated the rule that top-level `targetRef` selects the proxy being modified. Old global-scoped
egress continues to use the legacy path, so the change is non-breaking.

### `MeshAccessLog`, `MeshTimeout`, `MeshRateLimit`, `MeshFaultInjection`

These policies have `rules[].matches[]`, which can constrain both source (`spiffeID`) and
destination (`sni`) in the same entry. That expressiveness matters here — e.g. "log requests
from spiffeID X to MES Y" — and is something `spec.to[]` cannot express. The shared `SNIMatch`
type from `MeshTrafficPermission` is reused. The incomplete `rules[].matches[]` implementation
tracked in [#16460](https://github.com/kumahq/kuma/issues/16460) must be finished as part of
this work.

```yaml
type: MeshRateLimit
spec:
  targetRef:
    kind: Dataplane
    sectionName: ze-port
  rules:
    - matches:
        - sni:
            type: Exact
            value: sni.extsvc.default.zone-1.aws-aurora.8443
          spiffeID:
            type: Exact
            value: spiffe://trust.domain/ns/ns-1/sa/sa-1
      default:
        local:
          http:
            requests: 100
            interval: 1s
```

### Proxy-wide configuration — `MeshMetric`, `MeshTrace`, `MeshProxyPatch`

No per-destination granularity — these configure the entire proxy. Use top-level `targetRef`
(optionally with `sectionName`) and nothing else.

```yaml
type: MeshMetric
spec:
  targetRef:
    kind: Dataplane
    labels:
      kuma.io/listener-zoneegress: enabled
  default:
    backends:
      - type: Prometheus
        prometheus:
          port: 5670
          path: /metrics
```

### Out of scope

- `MeshRetry` on zone egress — squared retry amplification.
- `MeshOPA` — was never supported on zone egress; needs its own decision.
- `MeshGlobalRateLimit` — no action required.

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

Zone egress:
- Forwards mTLS-terminated traffic to MeshExternalService endpoints
- Enforces access control (MeshTrafficPermission) and rate limits per source/destination
- Applies policies (timeouts, circuit breakers, health checks) to external endpoints
- Applies observability policies (MeshAccessLog, MeshTrace, MeshMetric) to cross-zone traffic
- **Doesn't** perform the traffic routing (that is the sidecar's job)
- **Doesn't** act as a shared API gateway with complex routing logic

Zone ingress:
- Receives mTLS connections from remote zone sidecars and forwarding them to local service endpoints
- Applies observability policies (MeshAccessLog, MeshMetric) to cross-zone traffic
- **Doesn't** terminate the mTLS
- **Doesn't** enforce access control (connections are already mTLS-authenticated at the sidecar level)

### Targeting

Zone proxy `Dataplane` resources are targeted the same way as any other `Dataplane` — using
`targetRef.kind: Dataplane` with `name`, `labels`, or `sectionName`. No special targeting
mechanism is introduced.

- Use `name/namespace` to target a specific zone proxy instance. Because `Dataplane` resources always live on
  Zone CPs, name-based policies must be applied on the Zone CP, not the Global CP.
- Use `labels` to target a group of zone proxies (e.g. all proxies in a namespace).
  When the policy is applied on the Global CP, label-based matching is the only
  approach since `name/namespace` matching doesn't work due to KDS hashing.
- Use `sectionName` to target a specific listener when a Dataplane mixes zone proxy and regular
  inbound listeners.

### Destination Selector

Applying policies to zone proxies on a per-destination basis adds granularity.
For example, instead of enabling access logging for all `MeshExternalService`s on egress,
we can enable it for just one.

Every mTLS connection carries a **destination** encoded in the SNI,
see the [MADR-101](101-sni-format-improvements.md) for more details.
Zone proxies can leverage SNI-based matching to apply functionality selectively to a subset of destinations.

Policies can be divided into groups based on their support of per-destination granularity:

* (1) `MeshTrafficPermission` supports per-destination granularity.
  The policy already has support Envoy Matching API, adding `sni` matcher is a straightforward change.

* (2) `MeshCircuitBreaker`, `MeshHealthCheck` support per-destination granularity but only on zone proxies.
  This is possible because zone proxies generate separate Envoy `filterChains` for different destinations.

* (3) `MeshAccessLog`, `MeshTimeout`, `MeshRateLimit`, `MeshFaultInjection` support per destination granularity.
  These policies have support for `rules[].matches[]`
  (although the implementation is currently incomplete, see [#16460](https://github.com/kumahq/kuma/issues/16460))
  and at the same time `spec.to[]` (similar to group (2)). 
  We have a choice on how we'd like to implement this group.

* (4) `MeshMetric`, `MeshTrace`, `MeshProxyPatch` don't support per-destination granularity.
  These policies are applied to the entire proxy at once.


#### (1) `MeshTrafficPermission`

`MeshTrafficPermission` already supports matches.
We just need to agree on the matcher format that will be implemented under the hood using SNI matching.

##### Option A: SNI string match in `Match`

Extend the `Match` struct with an `sni` field:

```go
type Match struct {
    SpiffeID *SpiffeIDMatch `json:"spiffeID,omitempty"`
    SNI      *SNIMatch      `json:"sni,omitempty"`
}
```

With the KRI-based SNI format [MADR-101](101-sni-format-improvements.md), SNIs follow the
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
  rules:
    - default:
        allow:
          - spiffeID:
              type: Exact
              value: "spiffe://default/ns/backend-ns/sa/backend"
            sni:
              type: Exact
              value: "sni.extsvc.default.zone-1.aws-aurora.8443"
```

* Good, because it maps directly to Envoy's filter chain match on `server_names`.
* Good, because it is transparent — what you write is what Envoy matches on.
* Good, because SNI is observable at the wire - policy application can be verified by inspecting traffic directly.
* Good, because the KRI-based SNI format (MADR-101) makes SNIs human-readable and
  predictable — users can construct them from resource attributes without querying the CP.
* Bad, because coupling policy to SNI strings creates a dependency on SNI format stability,
  even though [MADR-101](101-sni-format-improvements.md) makes SNIs human-readable and predictable.
* Bad, because users are already accustomed to the selector-based `targetRef` model and
  would need to learn a new, non-intuitive matching mechanism.
* Bad, because resources created on the Global CP do not include a zone component in
  their SNI, which is confusing — users must understand when zone is present vs omitted
  in the SNI string.

##### Option B: `targetRef` in `Match`

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
  rules:
    - default:
        allow:
          - spiffeID:
              type: Exact
              value: "spiffe://default/ns/backend-ns/sa/backend"
            targetRef:
              kind: MeshExternalService
              labels:
                k8s.kuma.io/namespace: backend-ns
```

* Good, because it is consistent with how other Kuma policies reference resources.
* Good, because label-based matching provides broad scoping (by namespace, zone, team, etc.) without requiring users to know SNI formats.
* Good, because policies are insensitive to SNI format changes, if the SNI format evolves, existing policies require no updates; the CP re-resolves them.
* Bad, because CP must resolve `targetRef` to SNI(s) during policy processing (when building Envoy configuration).
* Bad, because `labels` matching can unintentionally select multiple MeshExternalServices, which may not be the user's intent.
* Bad, because `targetRef` in `Match` is only meaningful on zone proxy Dataplanes where per-destination filter chains exist. On regular sidecars the field has no matching filter chain and is silently ignored — same behavior as MeshHTTPRoute on DPPs without the referenced route.

##### Decision

Implement `sni` in the first iteration:

```go
type Match struct {
    SpiffeID *SpiffeIDMatch `json:"spiffeID,omitempty"`
    SNI      *SNIMatch      `json:"sni,omitempty"`
}

// +kubebuilder:validation:Enum=Exact
type SNIMatchType string

const SNIExactMatchType SNIMatchType = "Exact"

type SNIMatch struct {
    Type  SNIMatchType `json:"type"`
    Value string       `json:"value"`
}
```

- `sni` maps directly to Envoy's `server_names` filter chain match — no CP resolution step.
- `sni` is optional. When absent, the match applies to all destinations (same as today).
  When present, the match applies only to the filter chain whose SNI matches.
- When a `Match` contains both `spiffeID` and `sni`, both conditions must hold.
- With the KRI-based SNI format [MADR-101](101-sni-format-improvements.md), SNIs are human-readable and predictable —
  users construct them from resource attributes without querying the CP.
- `sni` in `Match` is not limited to zone proxy Dataplanes. It applies wherever the policy
  implementation has SNI available for matching traffic; zone proxy inbound listeners are the
  first supported consumer.

`targetRef` in `Match` may be added as a follow-up once `sni` is proven in production.
The `Match` struct can be extended without breaking existing policies.

##### Policy Examples

###### Access control: allow specific service to reach specific external resource

```yaml
type: MeshTrafficPermission
mesh: default
name: backend-to-aurora
spec:
  targetRef:
    kind: Dataplane
    sectionName: ze-port
  rules:
    - default:
        allow:
          - spiffeID:
              type: Exact
              value: spiffe://default/ns/backend-ns/sa/backend
            sni:
              type: Exact
              value: sni.extsvc.default.zone-1.aws-aurora.8443
```

No other service can reach `aws-aurora` through zone egress because no other `allow` entry matches.

###### Access control: deny specific source from a specific external resource

Combines SpiffeID and destination matching (AND semantics). Only `spiffeID` + `sni` in the same entry are AND-ed; separate entries are OR-ed.

```yaml
type: MeshTrafficPermission
mesh: default
name: deny-compromised-worker-to-aurora
spec:
  targetRef:
    kind: Dataplane
    sectionName: ze-port
  rules:
    - default:
        deny:
          - spiffeID:
              type: Exact
              value: "spiffe://default/ns/backend-ns/sa/compromised-worker"
            sni:
              type: Exact
              value: sni.extsvc.default.zone-1.aws-aurora.8443
```

###### Access control: deny all access to a specific external resource ("deny always wins")

A `deny` entry with only `sni` and no `spiffeID` matches every source. Because deny always wins
over allow, this policy blocks all access to `aws-aurora` regardless of any other MTP with `allow`.

```yaml
type: MeshTrafficPermission
mesh: default
name: deny-aurora
spec:
  targetRef:
    kind: Dataplane
    sectionName: ze-port
  rules:
    - default:
        deny:
          - sni:
              type: Exact
              value: sni.extsvc.default.zone-1.aws-aurora.8443
```

#### (2) `MeshCircuitBreaker`, `MeshHealthCheck`

These policies support per-destination granularity on zone proxies, but they don't support `rules[].matches[]`.
Today, it's possible to use these policies to indirectly configure legacy global scoped zone egress.

For example,

```yaml
type: MeshCircuitBreaker
spec:
  targetRef:
    kind: Dataplane
    labels:
      app: client
  to:
    - targetRef:
        kind: MeshExternalService
        labels:
          kuma.io/display-name: httpbin
      default: $conf1
```

The configuration `$conf1` is going to be applied to Zone Egress, even though `spec.targetRef` targets the `app: client`.

**The existence of these policies breaks the established rule "top-level targetRef selects what proxy to be modified",
so we should not apply this approach to new mesh-scoped zone proxies.**

Instead, the policy should configure Zone Egress only when the top-level `targetRef` selects the zone proxy.
In that case, `spec.to[].targetRef` selects the `filterChain` on zone proxy.

Top-level `targetRef.sectionName` still should work and select the `listener` by its name.

For example,

```yaml
type: MeshCircuitBreaker
spec:
  targetRef:
    kind: Dataplane
    sectionName: ze-port
  to:
    - targetRef:
        kind: MeshExternalService
        labels:
          kuma.io/display-name: httpbin
      default: $conf1
```

The policy will be applied on listeners named `ze-port` on the `filterChain` that sends the traffic to `httpbin`.

This transition can be made non-breaking by keying on how the MeshExternalService cluster is
generated: when the cluster routes through the old global-scoped Zone Egress listener, the
`EgressMatchedPolicies` path applies these policies to the egress as before. When the cluster
routes through a mesh-scoped zone proxy Dataplane, the policies are applied on the sidecar's
outbound cluster instead. Both paths can coexist during the migration window, so no `UPGRADE.md`
entry is required for 2.14.

#### (3) `MeshAccessLog`, `MeshTimeout`, `MeshRateLimit`, `MeshFaultInjection` 

These policies support `rules[].matches[]`,
although the implementation is incomplete and going to include only `spiffeID` matcher,
see [#16460](https://github.com/kumahq/kuma/issues/16460)

There are 2 options how could we implement these policies on zone proxies.

##### Option A: support `rules[].matches[]` on zone proxies

When applied on zone proxy `rules[].matches[].sni` selects `filterChain` and applies the policy to it.

```yaml
type: MeshRateLimit
spec:
  targetRef:
    kind: Dataplane
    sectionName: ze-port
  rules:
    - matches:
        - sni:
            type: Exact
            value: sni.extsvc.default.zone-1.aws-aurora.8443
      default:
        local:
          http:
            requests: 100
            interval: 1s
```

* Good, because it reuses the `SNIMatch` type introduced for group (1) — the matching API
  implementation is shared across policies.
* Bad, because users must know the SNI format and cannot reference a `MeshExternalService`
  by name or labels.
* Bad, because group (2) policies (`MeshCircuitBreaker`, `MeshHealthCheck`) have no `rules[]`
  field and must use `spec.to[]` for destination targeting. Picking `rules[].matches[]` here
  means users see two different shapes for the same conceptual operation (selecting a
  destination filter chain on a zone proxy) depending on which policy they pick up.
* Bad, because the existing `rules[].matches[]` implementation is incomplete on these policies
  (see [#16460](https://github.com/kumahq/kuma/issues/16460)) and must be finished before this
  is usable end-to-end.

##### Option B: support `spec.to[]` on zone proxies similar to group (2)

Top-level `targetRef` selects the zone proxy (or its listener via `sectionName`),
and `spec.to[].targetRef` selects the `filterChain` on that zone proxy.

```yaml
type: MeshAccessLog
spec:
  targetRef:
    kind: Dataplane
    sectionName: ze-port
  to:
    - targetRef:
        kind: MeshExternalService
        labels:
          kuma.io/display-name: aws-aurora
      default: $conf1
```

* Good, because it is consistent with group (2) (`MeshCircuitBreaker`, `MeshHealthCheck`) on zone proxies.
* Good, because users can target a `MeshExternalService` by labels without knowing the SNI format.
* Bad, because it cannot express source-conditioned per-destination policies like "access log
  requests from spiffeID X to MeshExternalService Y" — `spec.to[]` selects only the destination
  filter chain and has no place to constrain the source identity.

##### Decision

Implement Option A: `rules[].matches[].sni` selects the `filterChain` on the zone proxy.

The deciding factor is expressiveness: `rules[].matches[]` can constrain both source
(`spiffeID`) and destination (`sni`) in the same entry, which is required for use cases like
"access log requests from spiffeID X to MeshExternalService Y". Option B's `spec.to[]` cannot
express this — it selects a destination filter chain with no place for source conditioning.

The shape inconsistency with group (2) (`MeshCircuitBreaker`, `MeshHealthCheck`, which use
`spec.to[]`) is acceptable: groups (2) and (3) differ in that group (3) policies legitimately
care about the source, so reusing the matcher API introduced for `MeshTrafficPermission` is the
right fit. The `SNIMatch` type is shared across (1) and (3).

The incomplete `rules[].matches[]` implementation tracked in
[#16460](https://github.com/kumahq/kuma/issues/16460) must be finished as part of this work.

##### `spec.to[]` is also accepted as an alias for `rules[].matches[].sni`

Per Kuma's "apply where possible" style, `spec.to[]` on a zone proxy is not rejected for
group (3) policies — it is accepted as an alias for `rules[].matches[].sni`. The CP
resolves `spec.to[].targetRef → MeshExternalService → SNI`, locates the filter chain
whose `server_names` matches that SNI, and applies the policy's `default` fields wherever
each field lives in Envoy (filter chain, cluster, HCM filter, route). The outcome is
identical to writing `rules[].matches[].sni` with the resolved SNI.

```yaml
# These two are equivalent on zone egress.

# (a) spec.to[] form — label-friendly, no SNI knowledge required
type: MeshTimeout
spec:
  targetRef: { kind: Dataplane, sectionName: ze-port }
  to:
    - targetRef:
        kind: MeshExternalService
        labels: { kuma.io/display-name: aws-aurora }
      default:
        connectionTimeout: 5s
        http: { requestTimeout: 10s }

# (b) rules[].matches[] form — explicit SNI, can also constrain source
type: MeshTimeout
spec:
  targetRef: { kind: Dataplane, sectionName: ze-port }
  rules:
    - matches:
        - sni: { type: Exact, value: sni.extsvc.default.zone-1.aws-aurora.8443 }
      default:
        connectionTimeout: 5s
        http: { requestTimeout: 10s }
```

`rules[].matches[]` remains the only form that can constrain by source — pick it whenever
`spiffeID` matters. Use `spec.to[]` when destination-only is enough and label-based
selection is preferable to writing the SNI string.

##### Matches containing `spiffeID` and field placement

When a `rules[].matches[]` entry contains `spiffeID`, the policy default is asking to be
source-conditioned. Each field of the default lands at a specific position in the Envoy
config; whether source conditioning is honored depends on whether SPIFFE is available at
that position — and adding `sni` to the same match does not change the answer, it only
locates the filter chain/cluster.

- **Applied** — fields evaluated *after* the mTLS handshake, where SPIFFE is known: e.g.
  `MeshAccessLog` formatters/conditional logging, `MeshRateLimit` HTTP-level descriptors
  on HCM filter chains, RBAC-gated HTTP fields.
- **Not applied** — fields on the cluster (e.g. `MeshTimeout.connectionTimeout`) or on
  the filter chain itself (e.g. TCP `idleTimeout` on `tcp_proxy`). The cluster is shared
  across all sources reaching the same destination; the filter chain is selected before
  the handshake completes. In both cases the position cannot be keyed by source even when
  `sni` is also present in the match.

The validator emits a warning when a match with `spiffeID` specifies a field that cannot
be source-keyed, so operators see why their config did not take effect.

#### (4) `MeshMetric`, `MeshTrace`, `MeshProxyPatch`

These policies do not support per-destination granularity — they configure the proxy as a whole.
Top-level `targetRef` (with optional `sectionName`) selects the zone proxy (or one of its
listeners) and the policy is applied to all filter chains on that target. There is no
`spec.to[]` or `sni` matching involved.

```yaml
type: MeshMetric
spec:
  targetRef:
    kind: Dataplane
    labels:
      kuma.io/listener-zoneegress: enabled
  default:
    backends:
      - type: Prometheus
        prometheus:
          port: 5670
          path: /metrics
```

## Security implications and review

### MeshTrafficPermission Default Behavior

Zone egress is a security boundary — it is the sole exit path for MeshExternalService traffic.
The default behavior when no MeshTrafficPermission targets a mesh-scoped zone egress Dataplane
is **deny-all**.

The legacy flag `mesh.spec.routing.defaultForbidMeshExternalServiceAccess` applied only to the
old global-scoped `ZoneEgress` resource. It is irrelevant for mesh-scoped zone proxy Dataplanes
and will be removed in 3.0.

Operators must create explicit `allow` entries in MeshTrafficPermission to grant access.

### Audit Trail

MeshAccessLog on zone egress inbound captures the source SPIFFE identity and the original
destination SNI, providing an audit trail of which workload accessed which external resource.

## Implications for Downstream Projects

- **MeshGlobalRateLimit**: no action required.
- **MeshOPA**: was never supported on zone egress. Explicitly out of scope for this MADR;
  requires a separate decision if needed.

## UPGRADE.md

### 2.14

None

### 3.0

* global scoped zone proxies are removed
* drop MeshLoadBalancingStrategy support for MeshExternalServices
* `defaultForbidMeshExternalServiceAccess` is removed

