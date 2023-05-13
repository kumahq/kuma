# `MeshTCPRoute`

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/6343

## Context and Problem Statement

Now that we have the [new policy matching resources](./005-policy-matching.md),
[new traffic route designed](https://github.com/kumahq/kuma/issues/4743), and
[basic api of `MeshHTTPRoute`](https://github.com/kumahq/kuma/issues/5470)
implemented, it's time to introduce `MeshTCPRoute`.

### Goals

Policy should allow to
* reroute TCP traffic to different versions of a service or even completely
  different service;
* split TCP traffic between services with different tags implementing 
  A/B testing or canary deployments.

### Non-goals

* TCP traffic mirroring

  It would be nice for the policy to allow to mirror TCP traffic. Unfortunately
  it seems to be a non-trivial task as according to short research, Envoy
  doesn't support this case out of the box. If there will be need for
  this feature, it should be implemented in the future iterations.

  Ref:
  * https://github.com/envoyproxy/envoy/issues/18172

* `MeshGatewayTCPRoute`

  According to [Traffic Route MADR](./011-mesh-traffic-route.md), every 
  protocol-related route policy, should be split into two versions - one for
  gateway (i.e. `MeshGatewayTCPRoute`) and one for other resources (i.e.
  `MeshTCPRoute`).

  In case of `MeshTCPRoute` it doesn't make sense to introduce two types of
  this policy as the only difference would be the root `targetRef` which in
  gateway's case would point to `MeshGateway`. This MADR addresses only
  `MeshTCPRoute` and the first implementation should accept non-MeshGateway 
  root targetRefs only. If there will be consensus to unify `Mesh(.*)Route`
  with `MeshGateway(.*)Route`, we'll be able to just implement support for
  this targetRef. If not, we can easily introduce `MeshGatewayTCPRoute`.

## Considered Options

* Create a `MeshTCPRoute` policy

## Decision Outcome

Chosen option: Create a `MeshTCPRoute` policy

### `spec.to[].rules`

At this point there is no plan to introduce address matching capabilities for
`MeshTCPRoute` in foreseeable future. We try to be as close with structures of
our policies to the Gateway API as possible. It means, that even if Gateway API
currently doesn't have plans to support this kind of matching as well (ref.
[Kubernetes Gateway API GEP-735: TCP and UDP addresses matching](https://gateway-api.sigs.k8s.io/geps/gep-735/)),
its structures are ready to potentially support it.

As a result every element of the route destination section of the policy
configuration (`spec.to[]`) contains a `rules` property. This property is a list
of elements, which potentially will allow to specify `match` configuration.

Implementation of the `MeshTCPRoute` which should be a result of this document,
should validate that this list will contain only one element. This is due
to the fact, that without specifying `match`es, it would be nonsensical to
accept more `rules.`

### `MeshTCPRoute` and highest common protocol precedence in situation of clashes

In situations of clashes between different protocol routes, policy with protocol
higher on the OSI model matching both sides of the route takes precedence. One
side of the route (a source) is target specified in `spec.targetRef`. Other side
of the route (a destination) is a target specified in `spec.to[].targetRef`.

It means, in a hypothetical situation when sources and destinations targeted by
`MeshTCPRoute` and `MeshHTTPRoute` policies overlap `MeshHTTPRoute` policy
will take precedence over `MeshTCPRoute` only when protocols of both - sources
and destinations will be equal `http` or `http2`.

| source | destination | precedence    |
|--------|-------------|---------------|
| tcp    | tcp         | MeshTCPRoute  |
| tcp    | http        | MeshTCPRoute  |
| http   | tcp         | MeshTCPRoute  |
| http   | http        | MeshHTTPRoute |

### Traffic Rerouting

If `matches` succeeds, the request is routed to the specified destinations
(similar like in traffic split, but in this case `weight` property can be
omitted)

The destinations appear as references in `backendRefs`.

```yaml
spec:
  targetRef:
    kind: Mesh|MeshSubset|MeshService|MeshServiceSubset
  to:
  - targetRef:
      kind: MeshService
      name: tcp-backend
    rules:
    - default:
        backendRefs:
        - kind: MeshService
          name: tcp-other-backend
```

### Traffic Split

If `matches` succeeds, the request is routed to the specified weighted
destinations.

The destinations appear as references in `backendRefs`.

```yaml
spec:
  targetRef:
    kind: Mesh|MeshSubset|MeshService|MeshServiceSubset
  to:
  - targetRef:
      kind: MeshService
      name: tcp-backend
    rules:
    - default:
        backendRefs:
        - weight: 90
          kind: MeshServiceSubset
          name: tcp-backend
          tags:
            version: v2
        - weight: 10
          kind: MeshServiceSubset
          name: tcp-backend
          tags:
            version: v2
```
