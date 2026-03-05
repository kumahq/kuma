# Zone Proxy Deployment Topology

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/15672

## Context and Problem Statement

Zone proxies serve two distinct roles in multizone deployments:

- **Zone Ingress** — accepts cross-zone traffic from other zones destined for services in this zone.
- **Zone Egress** — forwards outgoing cross-zone traffic from services in this zone to services in other zones.

Historically, ZoneIngress and ZoneEgress were separate global-scoped resources, each backed by their own Deployment.
With the transition to mesh-scoped Dataplane resources ([MADR 095](095-mesh-scoped-zone-ingress-egress.md)), a single Dataplane can carry both listener types simultaneously.
[MADR 094](094-zone-proxy-deployment-model.md) introduced the `meshes` Helm list and defined `ingress`, `egress`, and `combinedProxies` as topology options, deferring the recommendation to this MADR.

The questions to answer:

1. Should zone ingress and zone egress share a single Deployment or remain separate?
2. When should each topology be used?
3. How can users deploy ingress-only or egress-only?

## Design

### Option 1: Separate Deployments

Each role is deployed independently. Users enable ingress-only, egress-only, or both via the `enabled` flag:

```yaml
# Both (default recommended)
meshes:
  - name: default
    ingress:
      enabled: true
    egress:
      enabled: true

# Ingress-only
meshes:
  - name: default
    ingress:
      enabled: true

# Egress-only
meshes:
  - name: default
    egress:
      enabled: true
```

Each enabled role gets its own Deployment and HPA (Horizontal Pod Autoscaler). Zone ingress gets a Service (LoadBalancer/NodePort) so other zones can reach it; zone egress does not need one since it is only accessed from within the same zone. Both connect to the control plane independently, producing separate Dataplane resources:

```yaml
# Zone ingress Dataplane
type: Dataplane
mesh: default
name: zone-proxy-default-ingress-abc12
spec:
  networking:
    address: 10.0.0.1
    listeners:
      - type: ZoneIngress
        address: 10.0.0.1
        port: 10001
        name: zi-port
        state: Ready
---
# Zone egress Dataplane
type: Dataplane
mesh: default
name: zone-proxy-default-egress-xyz34
spec:
  networking:
    address: 10.0.0.2
    listeners:
      - type: ZoneEgress
        address: 10.0.0.2
        port: 10002
        name: ze-port
        state: Ready
```

Pros:
- Good, because it preserves the existing operational model operators are familiar with.
- Good, because each role scales independently — a zone that primarily serves traffic can scale ingress without scaling egress.
- Good, because ingress and egress are in separate failure domains; a crash in one does not affect the other (e.g.: incorrect xDS configuration, identity provisioning error).
- Good, because different resource limits and Kubernetes security contexts can be applied per role.

Cons:
- Bad, because it requires managing two Deployments per mesh instead of one, which also increases cost.

### Option 2: Combined Deployment (`combinedProxies`)

A single Deployment runs both roles. The pod exposes two ports and registers one Dataplane resource with both listener types:

```yaml
meshes:
  - name: default
    combinedProxies:
      enabled: true
```

```yaml
type: Dataplane
mesh: default
name: zone-proxy-default-combined-abc12
spec:
  networking:
    address: 10.0.0.1
    listeners:
      - type: ZoneIngress
        address: 10.0.0.1
        port: 10001
        name: zi-port
        state: Ready
      - type: ZoneEgress
        address: 10.0.0.1
        port: 10002
        name: ze-port
        state: Ready
```

`combinedProxies` is mutually exclusive with `ingress`/`egress` — Helm validates and fails if both are set.

Even in combined mode, a Service is created only for the zone ingress port so its address can be advertised to other zones for cross-zone routing.

When using an HPA, scale-up is triggered by either ingress or egress load — a spike on one side scales the other. This is acceptable when traffic patterns are symmetric but wasteful for asymmetric zones.

Pros:
- Good, because it reduces operational surface to one Deployment per mesh.
- Good, because it halves the minimum pod count, lowering resource cost.

Cons:
- Bad, because ingress and egress share a failure domain — a pod crash takes down both roles simultaneously.
- Bad, because HPA scales both roles together, which is wasteful for zones with asymmetric ingress/egress traffic.
- Bad, because separate resource limits and security contexts per role are not possible.

### Comparison

| Aspect | Separate | Combined |
|:-------|:---------|:---------|
| Pods per mesh | 2 minimum | 1 minimum |
| HPAs | 2 (independent) | 1 (shared) |
| Independent scaling | Yes | No |
| Failure isolation | Independent | Single failure domain |
| Deployments to manage | 2 | 1 |

## Implications for Kong Mesh

None.

## Decision

**Separate Deployments are the recommended default for zone proxy deployment.**

`meshes[].ingress` and `meshes[].egress` deploy zone ingress and zone egress as independent Deployments, each with their own HPA and lifecycle. This preserves the existing operational model and provides independent scaling and failure isolation.

`meshes[].combinedProxies` is opt-in and suited for:
- Development and test environments
- Edge zones or other resource-constrained environments
- Zones with symmetric ingress/egress traffic

| Question | Decision |
|:---------|:---------|
| Recommended default | Separate (`ingress` + `egress`) |
| Opt-in alternative | Combined (`combinedProxies`) |
| Ingress-only | Set only `ingress.enabled: true` |
| Egress-only | Set only `egress.enabled: true` |
| Mutual exclusivity | `combinedProxies` and `ingress`/`egress` cannot coexist in one mesh entry |

## Notes

- MADR 094 introduced the `meshes` Helm schema and defers the detailed topology recommendation to this document (footnote `[^2]` will be backfilled there).
- MADR 095 confirmed a single Dataplane can carry both zone proxy listener types, which is what makes `combinedProxies` technically feasible.
