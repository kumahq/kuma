# `MeshTCPRoute`

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/6343

## Context and Problem Statement

Now that we have the [new policy matching resources](./005-policy-matching.md),
[new traffic route designed](https://github.com/kumahq/kuma/issues/4743), and
[basic api of `MeshHTTPRoute`](https://github.com/kumahq/kuma/issues/5470)
implemented, it's time to introduce `MeshTCPRoute`.

`MeshTCPRoute` should allow to
* reroute TCP traffic to different versions of a service or even completely
  different service;
* split TCP traffic between services with different tags implementing 
  A/B testing or canary deployments;

## Considered Options

* Create a `MeshTCPRoute` policy

## Decision Outcome

Chosen option: Create a `MeshTCPRoute` policy

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
    default:
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
    default:
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
