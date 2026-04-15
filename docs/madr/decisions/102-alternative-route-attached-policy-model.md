# Zone Proxy Policy Model — Alternative: Route-Attached Policy Model

* Status: proposed

Technical Story: https://github.com/kumahq/kuma/issues/8417

## Context and Problem Statement

This document presents an alternative approach to [MADR-102](102.md) for zone egress policy matching.

Zone egress is a **two-dimensional enforcement point**: every request has a *source*
(SPIFFE identity of the calling workload) and a *destination* (the MeshExternalService
being accessed). Kuma's existing policy model handles one dimension at a time:
- Inbound policies: "who is calling me" (source dimension)
- Outbound policies: "where am I going" (destination dimension)

Zone egress needs BOTH dimensions simultaneously. MADR-102 solves this by adding
`targetRef` and `sni` fields into the inbound `Match` struct — embedding destination
semantics into a source-identity matcher.

This alternative proposes keeping the dimensions separate by reusing existing Kuma
patterns (`from`/`to`/`targetRef`) without introducing new API types or fields on `Match`.

### User Stories

Same as MADR-102:

* As a mesh operator I want to deny access to external resources by default and grant it only to
  specific services so that compromised workloads cannot pivot to external systems.
* As a mesh operator I want to give access to service owners to a specific external resource
  (e.g. AWS Aurora) or to a single HTTP endpoint of that resource so that the system follows
  the least privilege principle.
* As a mesh operator I want to know which workload accessed which external resource so that I
  can audit access post-incident.
* As a mesh operator I want to have observability available on zone proxies so that I can
  troubleshoot issues and monitor performance.
* As a mesh operator I want to be able to override Envoy connection and traffic parameters
  (connectTimeout, maxConnections, etc.) so that I can provide values suited for my traffic.
* As a mesh operator I want to rate limit requests to an external service so that clients don't
  go over service limits and exhaust the budget.
* As a service owner I want to grant access to specific external services for owned workloads.
* As a service owner I want to restrict specific paths on an external service via MeshHTTPRoute
  on zone egress.

## Design

### Core Idea

Model zone egress like Kuma already models MeshGateway: a dedicated proxy with routes
(MeshExternalService) attached to it. Policies attach to the proxy or its routes using
existing `targetRef` + `to` + `from` patterns.

**No new resource types. No new fields in `Match`. Reuse existing Kuma patterns consistently.**

### Three-Layer Targeting

```
Layer 1: targetRef       → selects the proxy (zone egress Dataplane via labels or sectionName)
Layer 2: to[].targetRef  → selects the destination (MeshExternalService)
Layer 3: from[].targetRef or allow[].spiffeId → selects the source (workload identity)
```

This is the same structure MeshFaultInjection already uses (it implements both
`PolicyWithFromList` and `PolicyWithToList`). The egress matching code in
`matchers/egress.go` already converts `to` policies into artificial `from` policies
for zone egress processing — this proposal makes that pattern explicit and primary.

### Comparison with MADR-102

| Aspect | MADR-102 | This Proposal |
|--------|----------|---------------|
| Destination in inbound | New `targetRef`/`sni` fields in `Match` | Standard `to.targetRef` (existing pattern) |
| MTP destination | `targetRef` inside `allow[]/deny[]` entries | `to` section with `allow[]/deny[]` nested per destination |
| SNI escape hatch | `sni` field with Exact/Prefix matching | Not needed — label selectors on MeshExternalService |
| Per-source per-dest | `spiffeId` + `targetRef` AND'd in same `Match` | `from` + `to` structure (existing MeshFaultInjection pattern) |
| Namespace/zone scoping | SNI prefix matching | Label selectors on `to.targetRef` |
| New API surface | 2 new fields on `Match` + `SNIMatch` type | 0 new types — reuse `from`/`to`/`targetRef` |
| Match struct changes | Yes (needs admission rejection on sidecars) | None |

### Authorization: MeshTrafficPermission with `to` on Zone Egress

Instead of adding `targetRef` inside `Match.allow[]/deny[]` entries, give
MeshTrafficPermission a `to` section **when targeting zone egress**. This keeps
destination selection in the same place it lives in every other policy.

```yaml
type: MeshTrafficPermission
mesh: default
name: backend-to-aurora-only
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
        action: Allow
        allow:
          - spiffeId:
              type: Exact
              value: "spiffe://default/ns/backend-ns/sa/backend"
```

The policy reads naturally: "on egress, for traffic TO aws-aurora, ALLOW spiffeId backend."

* Good, because `to.targetRef` for destination selection is the standard Kuma pattern.
* Good, because `allow[].spiffeId` stays purely about source identity — no overloading.
* Good, because no `sni` escape hatch is needed — `to.targetRef` can match by name, labels,
  or namespace using existing selectors.
* Good, because multiple `to` entries enable different rules per destination in one policy.
* Bad, because MeshTrafficPermission gains a `to` section it did not have before
  (only valid for zone proxy targeting — admission webhook rejects it on sidecars).

#### Deny-all with per-service exceptions

Enable `mesh.spec.routing.defaultForbidMeshExternalServiceAccess: true`, then:

```yaml
type: MeshTrafficPermission
mesh: default
name: allow-observability-to-aurora
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
        allow:
          - spiffeId:
              type: Prefix
              value: "spiffe://default/ns/observability/"
```

Block a compromised workload from a specific external service:

```yaml
type: MeshTrafficPermission
mesh: default
name: block-compromised-worker-from-aurora
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
        deny:
          - spiffeId:
              type: Exact
              value: "spiffe://default/ns/backend-ns/sa/compromised-worker"
```

Block a compromised workload from ALL external services (blanket deny — no `to` section):

```yaml
type: MeshTrafficPermission
mesh: default
name: block-compromised-worker-from-all
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

MTP without `to` applies to ALL filter chains (blanket deny/allow for all destinations).
MTP with `to` applies only to the referenced MeshExternalService filter chains.
MTP with `to` applies only to the referenced MeshExternalService filter chains.
Multiple MTPs targeting the same destination: concatenate rules, deny takes priority
(unchanged from MADR-081 algorithm).

### Rate Limiting: Per-source Per-destination

Use `from` + `to` together (like MeshFaultInjection already does):

```yaml
type: MeshRateLimit
mesh: default
name: backend-aurora-rate-limit
spec:
  targetRef:
    kind: Dataplane
    labels:
      k8s.kuma.io/zone-proxy-type: egress
  from:
    - targetRef:
        kind: MeshSubset
        tags:
          kuma.io/service: backend
      to:
        - targetRef:
            kind: MeshExternalService
            name: aws-aurora
          default:
            local:
              http:
                requestRate:
                  num: 50
                  interval: 1m
```

Or without source filtering (all sources, specific destination):

```yaml
type: MeshRateLimit
mesh: default
name: aurora-global-quota
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

### Namespace/Zone-wide Scoping (Replaces SNI Prefix Matching)

Instead of SNI prefix matching, use label selectors on MeshExternalService:

```yaml
type: MeshRateLimit
mesh: default
name: all-external-in-namespace
spec:
  targetRef:
    kind: Dataplane
    labels:
      k8s.kuma.io/zone-proxy-type: egress
  to:
    - targetRef:
        kind: MeshExternalService
        labels:
          k8s.kuma.io/namespace: backend-ns
      default:
        local:
          http:
            requestRate:
              num: 500
              interval: 1m
```

* Good, because no coupling to internal SNI format.
* Good, because works with any label, not just the SNI hierarchy.
* Good, because consistent with how `targetRef` labels work everywhere else in Kuma.
* Good, because future-proof — adding new labels requires no API changes.

### Fault Injection on Zone Egress

```yaml
type: MeshFaultInjection
mesh: default
name: egress-fault
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
        http:
          abort:
            httpStatus: 503
            percentage: 10
```

### Observability on Zone Egress

Unchanged. Attach to zone proxy Dataplane, no destination filtering needed:

```yaml
type: MeshAccessLog
mesh: default
name: egress-access-log
spec:
  targetRef:
    kind: Dataplane
    labels:
      k8s.kuma.io/zone-proxy-type: egress
  default:
    backends:
      - type: OpenTelemetry
        openTelemetry:
          endpoint: otel-collector.observability:4317
```

MeshMetric and MeshTrace follow the same pattern. No API changes needed.

### Outbound Policy Examples

Outbound policies (`to`) on zone egress target MeshExternalService directly.
These are identical to MADR-062/102:

#### Path-level access restriction

```yaml
type: MeshHTTPRoute
mesh: default
name: restrict-to-single-endpoint
spec:
  targetRef:
    kind: Dataplane
    sectionName: ze-port
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

### Updated Policy Matrix

| Policy                    | Zone Egress (`to` destination) | Zone Egress (`from` source) | Zone Egress (blanket) | Notes |
|---------------------------|-------------------------------|----------------------------|----------------------|-------|
| MeshAccessLog             | Yes                           | No                         | Yes                  | |
| MeshCircuitBreaker        | Yes (existing, MADR-062)      | No                         | No                   | |
| MeshFaultInjection        | Yes                           | Yes                        | Yes                  | |
| MeshHealthCheck           | Yes (existing, MADR-062)      | No                         | No                   | |
| MeshHTTPRoute             | Yes (requires HCM chain)      | No                         | No                   | Path restriction |
| MeshLoadBalancingStrategy | Yes (existing, MADR-062)      | No                         | No                   | |
| MeshMetric                | Yes                           | No                         | Yes                  | |
| MeshProxyPatch            | Yes                           | No                         | No                   | Last-resort Envoy override |
| MeshRateLimit             | Yes                           | Yes                        | Yes                  | Per-source per-dest via from+to |
| MeshRetry                 | Sidecar only (MADR-062)       | No                         | No                   | Squared retries |
| MeshTCPRoute              | No                            | No                         | No                   | |
| MeshTimeout               | Yes                           | No                         | No                   | |
| MeshTrace                 | Yes                           | No                         | Yes                  | |
| MeshTrafficPermission     | Yes (via `to` section)        | N/A (SpiffeID in allow/deny) | Yes                | `to` only valid on zone proxy |

### Zone Ingress Policy Matrix

Zone ingress `to.targetRef` would reference `MeshService` (not MeshExternalService).
Deferred to a follow-up — same as MADR-102.

## Security Implications and Review

### MeshTrafficPermission Default Behaviour

Same as MADR-102: `defaultForbidMeshExternalServiceAccess` defaults to `true` (deny-all).
Operators must create explicit MTP with `allow` entries.

### Resolution-failure Semantics

When a `to.targetRef` cannot be resolved (MeshExternalService does not exist or not yet synced):

- **Security policies** (MeshTrafficPermission): unresolvable `to` entry is treated as
  **match-none** (fail closed). Traffic to the unresolvable destination is denied.
- **Traffic policies** (MeshRateLimit, MeshFaultInjection, etc.): unresolvable `to` entry
  is silently dropped (no modification applied, but traffic is not blocked).

### SNI Tombstone Not Required

Since `to.targetRef` resolves by resource identity (name/labels) rather than by SNI string,
the CP naturally handles renames and deletes through standard resource lifecycle. When a
MeshExternalService is deleted, any `to.targetRef` referencing it becomes unresolvable and
fails closed. No explicit SNI tombstone mechanism is needed — this eliminates the credential
misdirection vector described in MADR-102 without additional CP logic.

### Audit Trail

MeshAccessLog on zone egress captures source SPIFFE identity and destination — same as MADR-102.

## Reliability Implications

### HTTP Filter Chains

Same as MADR-102: zone egress only creates an HCM chain per MeshExternalService when at least
one HTTP-level policy targets it. L4-to-L7 promotion must be observable via status condition.

### Rate Limit Interaction

Same as MADR-102: inbound (per-source) and outbound (per-destination quota) operate independently.
The `from` + `to` combined structure makes the interaction explicit in the policy definition
rather than implicit through separate policies.

### TargetRef Resolution Consistency

Same eventual-consistency model as MADR-102. Security policies fail closed during convergence.

## Implications for Kong Mesh

Same as MADR-102: Kong Mesh rate-limiting and header-injection policies should be validated
against zone egress filter chain architecture when `to` support is added.

## Decision

1. **Policy structure**:
   - MeshTrafficPermission gains a `to` section for zone egress targeting. `to.targetRef`
     selects the destination MeshExternalService. `allow[]/deny[]` within each `to` entry
     carry SpiffeID matchers for source identity. The `to` section is rejected by admission
     webhook when targeting non-zone-proxy Dataplanes.
   - Other inbound policies use `from` + `to` combined when per-source per-destination
     granularity is needed (existing pattern from MeshFaultInjection).
   - Outbound policies use the standard `to` structure (unchanged).

2. **No changes to `Match` struct**: The `Match` struct retains only `SpiffeID`.
   Destination selection lives in `to.targetRef` — the same place it lives in every
   other Kuma policy.

3. **No SNI in user-facing API**: Label selectors on `to.targetRef` replace SNI prefix
   matching. Labels are more powerful (arbitrary dimensions), more consistent (standard
   Kuma pattern), and decoupled from internal SNI format.

4. **HTTP filter chains**: Same as MADR-102. Zone egress creates HCM chain per
   MeshExternalService only when HTTP-level policy targets it.

5. **Default RBAC behaviour**: `defaultForbidMeshExternalServiceAccess` defaults to `true`.

6. **Resolution-failure semantics**: Same as MADR-102. Unresolvable `to.targetRef` in
   security policies fails closed.

7. **No SNI tombstone needed**: Resource identity resolution eliminates the credential
   misdirection vector without additional CP mechanisms.

## Notes

* MeshTrafficPermission `to` is only valid when targeting zone proxy Dataplanes.
  On regular sidecars, admission webhook rejects it.
* Zone ingress with `to.targetRef: kind: MeshService` is deferred to a follow-up.
* MeshRetry on zone egress remains out of scope — squared retry amplification.
* The `from` + `to` combined pattern already has precedent: MeshFaultInjection implements
  both `PolicyWithFromList` and `PolicyWithToList`. The egress matcher in
  `matchers/egress.go` already converts `to` into artificial `from` rules.
