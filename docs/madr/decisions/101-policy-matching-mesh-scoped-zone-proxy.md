# Policy Matching on MeshScoped Zone Proxy

* Status: accepted

Technical Story: https://github.com/Kong/kong-mesh/issues/9029

## Context and Problem Statement

Zone Ingress and Zone Egress were originally dedicated global-scoped resource types.
Being global-scoped meant they lived outside any mesh context, which created two fundamental blockers:

1. **No mesh identity** — `ZoneIngress`/`ZoneEgress` could not participate in mTLS with a
   proper SPIFFE identity. The trust domain is mesh-scoped, so a global resource has no mesh
   to derive an identity from. As a result, source-identity-based access control
   (MeshTrafficPermission) could not be enforced at zone proxy boundaries.

2. **No policy support** — the policy system operates on `Dataplane` resources within a mesh.
   Because `ZoneIngress`/`ZoneEgress` were not `Dataplane` resources, none of the mesh policies:
   observability (MeshAccessLog, MeshMetric, MeshTrace), security (MeshTrafficPermission), or
   traffic management (MeshRateLimit, MeshFaultInjection) — could target them.

[MADR-095](095-mesh-scoped-zone-ingress-egress.md) resolved this by modelling zone proxies as
mesh-scoped `Dataplane` resources with a `listeners` array. With that foundation in place, this
MADR establishes the **unified policy model** for zone proxies — how inbound and outbound policies
are structured, how zone proxy Dataplanes are targeted, and how policies can select a specific
MeshExternalService destination on zone egress. It resolves all policy placement items deferred
by [MADR-062](062-meshexternalservice-and-zoneegress.md).

## User Stories

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
  (connectTimeout, maxConnections, etc.) so that I can always provide values suited for my traffic.
* As a mesh operator I want to rate limit requests to an external service so that clients don't
  go over service limits and exhaust the budget.
* As a mesh operator I want to inject HTTP headers with a token on the egress for all outgoing
  requests to an external service so that all clients in the mesh can use the same token without
  granting access to the token to individual clients.
