# `MeshHTTPRoute`/`MeshTCPRoute` for `MeshGateway`

- Status: accepted

Technical Story: #8142

## Context and Problem Statement

We have some very similar resources, `MeshHTTPRoute`/`MeshTCPRoute` and `MeshGatewayRoute` that
both configure route matching, filtering and redirecting semantics.
This MADR proposes adapting `MeshHTTPRoute` and `MeshTCPRoute` for use with
`MeshGateway.`

## Decision Drivers <!-- optional -->

- Fewer CRDs
- Parity with `MeshGatewayRoute`

## Considered Options

- Keep two resources
- Use `MeshHTTPRoute`/`MeshTCPRoute`

## Decision Outcome

Chosen option:

- Adapt `MeshHTTPRoute`/`MeshTCPRoute`
- Add `spec.to[].hostnames` to `MeshHTTPRoute`
- Require `spec.to[].targetRef.kind: Mesh`
- Require `backendRefs` to be set
- Don't support policy matching on `MeshHTTPRoute`/`MeshTCPRoute` if it attaches to a gateway

### Hostname

With `MeshGatewayRoute` and `HTTPRoute`, users can set hostnames to only apply
the route rules to requests going to the given hosts.
We don't have this possibility with `MeshHTTPRoute`, so we can either:

- add `hostnames` to the `MeshHTTPRoute`, likely directly under `spec.to`
  entries
- have users on each rule do hostname matching using the header match on `host`

This MADR proposes adding `spec.to[].hostnames`. Hostname matching is in some sense the
"highest level" matching possible with HTTP and it's likely that all rules of a
given resource are intended to apply to a given hostname set, whether that's a
specific hostname, a wildcard pattern or to all hostnames.

The second option might be feasible but it only makes sense if the cost of having
host matching separate is very high. It also differs from Gateway API, which is
a disadvantage.

#### Precedence

Note that the order of matching requests to hostnames is always
dependent on the specified hostnames themselves, not the order of the routes
they are specified in.

Precedence is given to rules attached to hostnames in order:

* Characters in a matching non-wildcard hostname
* Characters in a matching hostname

Basically, rules are first applied without wildcards in the hostname are applied before
rules with a wildcard hostname.

Only if rules can't be sorted according to this order does the order of policies
come into play.

### `to.targetRef`

When using `MeshHTTPRoute`/`MeshTCPRoute` for service to service communication, it's very
useful to limit the effects based on request source and destination. We do
this by setting `spec.targetRef`, the source, and `spec.to[].targetRef`, the destination.

With `MeshGateway`, once a request hits the gateway, the routing resources
_define_ the destination. We have to put the rules somewhere, so this MADR
proposes `spec.to`. There's no real semantic value to a distinct
`spec.to[].targetRef`, or for that matter, a `spec.from[].targetRef`.

This MADR proposes requiring `spec.to[].targetRef.kind: Mesh`.

Allowing `to[].targetRef` to be empty would be a break from all other policies.
Eventually we can support it generally via a default value in the open API schema,
see also [#6070](https://github.com/kumahq/kuma/issues/6070).

### Route matching

This MADR doesn't address supporting policy matching of routes
that attach to `MeshGateways` because we can't policy match on `MeshGatewayRoute` at
the moment anyway. The feasability/details should be addressed in a separate MADR.

### `MeshHTTPRoute` example

```
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: edge-routes
spec:
  targetRef:
    kind: MeshGateway
    name: edge-gateway
  to:
  - targetRef:
      kind: Mesh
    hostnames:
    - example.com
    rules:
    - matches:
      - path:
          type: PathPrefix
          value: /
      default:
        backendRefs:
        - kind: MeshService
          name: backend_demo_svc_8080
```

### `MeshTCPRoute` example

```
apiVersion: kuma.io/v1alpha1
kind: MeshTCPRoute
metadata:
  name: edge-routes
spec:
  targetRef:
    kind: MeshGateway
    name: edge-gateway
  to:
  - targetRef:
      kind: Mesh
    rules:
      - default:
          backendRefs:
            - kind: MeshServiceSubset
              name: backend_kuma-demo_svc_3001
              tags:
                version: "1.0"
              weight: 90
            - kind: MeshServiceSubset
              name: backend_kuma-demo_svc_3001
              tags:
                version: "2.0"
              weight: 10
```

### Merging example

```
---
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: specific-hostname
spec:
  targetRef:
    kind: MeshGateway
    name: edge-gateway
  to:
  - targetRef:
      kind: Mesh
    hostnames:
    - test.example.com
    rules:
    - matches:
      - path:
          type: PathPrefix
          value: /
      default:
        backendRefs:
        - kind: MeshServiceSubset
          name: backend_demo_svc_8080
          tags:
            version: v1
---
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: wildcard-hostname
spec:
  targetRef:
    kind: MeshGateway
    name: edge-gateway
  to:
  - targetRef:
      kind: Mesh
    hostnames:
    - *.example.com
    rules:
    - matches:
      - path:
          type: Exact
          value: /v2
      default:
        backendRefs:
        - kind: MeshServiceSubset
          name: backend_demo_svc_8080
          tags:
            version: v2
```

results in the rules:

```
  to:
  - targetRef:
      kind: Mesh
    hostnames:
    - test.example.com
    rules:
    - matches:
      - path:
          type: PathPrefix
          value: /
      default:
        backendRefs:
        - kind: MeshServiceSubset
          name: backend_demo_svc_8080
          tags:
            version: v1
  - targetRef:
      kind: Mesh
    hostnames:
    - *.example.com
    rules:
    - matches:
      - path:
          type: Exact
          value: /v2
      default:
        backendRefs:
        - kind: MeshServiceSubset
          name: backend_demo_svc_8080
          tags:
            version: v2
```

so that requests to `test.example.com/v2` are sent to `version: v1`!.

On the other hand:

```
---
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: specific-hostname
spec:
  targetRef:
    kind: MeshGateway
    name: edge-gateway
  to:
  - targetRef:
      kind: Mesh
    hostnames:
    - test.example.com
    rules:
    - matches:
      - path:
          type: PathPrefix
          value: /
      default:
        backendRefs:
        - kind: MeshServiceSubset
          name: backend_demo_svc_8080
          tags:
            version: v1
---
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: wildcard-hostname
spec:
  targetRef:
    kind: MeshGateway
    name: edge-gateway
  to:
  - targetRef:
      kind: Mesh
    hostnames:
    - test.example.com
    - *.example.com
    rules:
    - matches:
      - path:
          type: Exact
          value: /v2
      default:
        backendRefs:
        - kind: MeshServiceSubset
          name: backend_demo_svc_8080
          tags:
            version: v1
```

gives

```
  to:
  - targetRef:
      kind: Mesh
    hostnames:
    - test.example.com
    rules:
    - matches:
      - path:
          type: PathPrefix
          value: /
      default:
        backendRefs:
        - kind: MeshServiceSubset
          name: backend_demo_svc_8080
          tags:
            version: v1
    - matches:
      - path:
          type: Exact
          value: /v2
      default:
        backendRefs:
        - kind: MeshServiceSubset
          name: backend_demo_svc_8080
          tags:
            version: v2
  - targetRef:
      kind: Mesh
    hostnames:
    - *.example.com
    rules:
    - matches:
      - path:
          type: Exact
          value: /v2
      default:
        backendRefs:
        - kind: MeshServiceSubset
          name: backend_demo_svc_8080
          tags:
            version: v2
```

so that requests to `test.example.com/v2` are sent to `version: v2`!. In other
words, merging between hostnames is not semantic and doesn't take into account wildcards.
It's based on an exact string match..

### Positive Consequences <!-- optional -->

- Fewer CRDs
- One fewer "old style matching" resource

### Negative Consequences <!-- optional -->

None
