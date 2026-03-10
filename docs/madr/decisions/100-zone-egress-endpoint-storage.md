# Zone Egress Address Storage for Endpoint Generation

* Status: proposed

## Context and Problem Statement

With mesh-scoped zone proxies (MADR-095), a zone egress is represented as a `Dataplane` resource
with a `ZoneEgress` listener instead of the legacy `ZoneEgress` resource. The listener carries
the address and port at which the egress sidecar accepts inbound mTLS traffic from other sidecars
in the same zone.

When the control plane builds endpoints for `MeshExternalService` traffic it must behave
differently depending on who is asking:

- **Normal sidecar**: traffic must be sent to the local zone egress, which then proxies to the
  real external endpoint.
- **Zone egress sidecar itself**: must connect directly to the real external endpoint.

The gap is that the existing endpoint-building logic reads the egress address from the legacy
`ZoneEgress` resource. There is no equivalent source for a `Dataplane`-based zone egress, so the
routing path for normal sidecars is currently broken for mesh-scoped proxies.

The question is: **where should the zone egress address be stored so the CP can build correct
endpoints for both cases?**

## Decision Drivers

- Normal sidecars must route `MeshExternalService` traffic through the zone egress.
- The zone egress sidecar must route directly to the external endpoint (no self-loop).
- Zone egress is zone-local — each zone handles its own external service traffic independently.
- The egress address must be available to the CP at context-build time without additional lag.

## Considered Options

### Option 1: Derive address directly from the live Dataplane at context-build time

When the CP builds the mesh context it collects egress addresses from both the legacy
`ZoneEgress` resources and any `Dataplane` resources that carry a `ZoneEgress` listener, merging
them into a unified list. The endpoint-building logic consumes this list instead of querying
only the legacy resource type.

#### Pros and Cons

* Good, because no new resource type, status field, or controller is needed
* Good, because the address is always current — read directly from the Dataplane at the moment
  the CP builds config, with no reconciliation lag

### Option 2: Store egress address in `MeshExternalService` status

A controller (MeshExternalService status updater) watches `Dataplane` resources with `ZoneEgress` listeners and writes the current
egress address into a new `status.egressEndpoints` field on each `MeshExternalService`. The
CP endpoint-building logic then reads from this status field rather than scanning Dataplanes.

The zone egress sidecar continues to use the external endpoint from the service spec directly,
ignoring the status field.

Example status shape:
```yaml
status:
  addresses:
    - hostname: example.extsvc.mesh.local
  egressEndpoints:
    - address: 10.42.0.11
      port: 10002
```

#### Pros and Cons

* Good, because the egress address is co-located with the service that needs it — a single,
  well-known field rather than a scan across all Dataplanes
* Bad, because a new reconciliation loop is required — a controller must watch Dataplanes and
  keep the status field up to date
* Bad, because the status field is inherently stale between a Dataplane change (pod restart,
  IP reassignment) and the next reconciliation cycle; during that window normal sidecars may
  send traffic to a dead endpoint
* Bad, because every egress address change (e.g. a rolling update of the zone egress deployment)
  triggers a fan-out reconciliation across all `MeshExternalService` resources in the zone —
  write amplification that grows linearly with the number of external services

## Decision

Chosen option: **Option 1** (derive address directly from the live Dataplane).

Zone egress is inherently zone-local: each zone handles its own external service traffic and
there is no cross-zone egress routing. This means the address only needs to be available within
the zone where the egress runs, and pod IPs are fully routable within a zone on any Kubernetes
cluster with a standard CNI. The two-source merge adds a small amount of code at context-build
time but avoids an entire new reconciliation loop and its associated staleness window.

Option 2's benefit of co-locating the address on the service resource is real, but the cost
(extra controller, potential for stale routing, coupling of concerns) outweighs it when the
simpler approach already satisfies all requirements.

## Migration

During the transition from cluster-scoped `ZoneEgress` resources to mesh-scoped `Dataplane`-based
zone egress proxies, both resource types may coexist in the same zone. The CP must handle all
combinations without operator intervention:

| Cluster-scoped `ZoneEgress` present | Mesh-scoped `Dataplane` zone egress present | Behaviour |
|:---:|:---:|:---|
| yes | no | Use the cluster-scoped `ZoneEgress` address — legacy path, unchanged |
| no | yes | Use the `Dataplane` listener address — new path |
| yes | yes | Prefer the mesh-scoped `Dataplane` zone egress; ignore the cluster-scoped resource |
| no | no | No egress routing |

The preference for mesh-scoped proxies when both are present allows operators to migrate
incrementally: deploy the new `Dataplane`-based zone egress, verify it works, then decommission
the legacy `ZoneEgress` resource — without any traffic interruption or coordination window.

## Implications for Kong Mesh

None.
