# `MeshHTTPRoute` for `MeshGateway`

- Status: accepted

Technical Story: #8142

## Context and Problem Statement

We have two very similar resources, `MeshHTTPRoute` and `MeshGatewayRoute` that
both configure route matching, filtering and redirecting semantics.
This MADR proposes adapting `MeshHTTPRoute` for use with
`MeshGateway.`

## Decision Drivers <!-- optional -->

* Fewer CRDs
* Parity with `MeshGatewayRoute`

## Considered Options

- Keep two resources
- Use `MeshHTTPRoute`

## Decision Outcome

Chosen option:

- Adapt `MeshHTTPRoute`
- Add `spec.to[].hostnames`
- Allow `spec.to[].targetRef` to be empty
- Require `backendRefs` to be set
- Don't support policy matching on `MeshHTTPRoute` if it attaches to a gateway

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

### `to.targetRef`

When using `MeshHTTPRoute` for service to service communication, it's very
useful to limit the effects based on request source and destination. We do
this by setting `spec.targetRef`, the source, and `spec.to[].targetRef`, the destination.

With `MeshGateway`, once a request hits the gateway, the routing resources
_define_ the destination. We have to put the rules somewhere, so this MADR
proposes `spec.to` but there's no semantic value to a `spec.to[].targetRef`
or for that matter, a `spec.from[].targetRef`.

This MADR proposes allowing `spec.to[].targetRef` to be omitted. Another
possibility would be requiring `spec.to[].targetRef.kind: Mesh` but it's
not clear that this has any advantages.

### `MeshHTTPRoute` matching

This MADR doesn't address supporting policy matching of `MeshHTTPRoutes`
that attach to `MeshGateways` because we can't policy match on `MeshGatewayRoute` at
the moment anyway. The feasability/details should be addressed in a separate MADR.

### Example

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
  - hostnames:
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

### Positive Consequences <!-- optional -->

- Fewer CRDs
- One fewer "old style matching" resource

### Negative Consequences <!-- optional -->

None
