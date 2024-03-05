# `MeshTrafficRoute`

- Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/4743
Other issues covered:

- https://github.com/kumahq/kuma/issues/4691
- https://github.com/kumahq/kuma/issues/4690

## Context and Problem Statement

Now that we have the [new policy matching resources](https://github.com/kumahq/kuma/blob/master/docs/madr/decisions/005-policy-matching.md), it's time to create one for traffic routing.
We currently have both `TrafficRoute` and `MeshGatewayRoute`.

This MADR does not address Gateway API integration per se but does consider Gateway
API's `HTTPRoute`, both as inspiration as well as to make sure we don't block
integration.

## Decision Drivers

- `MeshGatewayRoute` and `TrafficRoute` have a large overlap in functionality
  and the use cases for routing from a gateway and between services are similar
- Gateway API has different resources per protocol

## Considered Options

- Two resources: `MeshGatewayRoute` and `MeshTrafficRoute`
- Unified resource for services and gateway: `MeshTrafficRoute`
- Unified resources but one per protocol, e.g. `MeshHTTPRoute`, `MeshTCPRoute`
- Two resources: `MeshGatewayRoute` and `MeshTrafficRoute`, which both point to
  a `HTTPRouteRules` resource

## Decision Outcome

We decided to create two resources `MeshGatewayHTTPRoute` and `MeshHTTPRoute`.
They have the exact same structure for routing rules but the way they refer to
targets is different.

Each future protocol will get its own resource:

- `MeshHTTPRoute`
- `MeshTCPRoute`

which can be extended with others like GRPC.

### `MeshHTTPRoute`

The new resource routes and alters requests from one service to another
depending on where the request is coming from and where it's going to.
Routes end up in the client-side Envoy config so the source of the affected requests will
be under `spec.targetRef`. The destination we can select in `spec.to`.

```yaml
spec:
 targetRef: ... # from where requests are coming
 to:
  - targetRef: ... # to where requests are going
    rules: ...
```

The route that's used for a given request is the most specific one that Kuma can
apply given the `kuma.io/protocol` tag.

### `MeshGatewayHTTPRoute`

This new resource attaches directly to a `MeshGateway` via `spec.targetRef`:

```yaml
spec:
 targetRef:
  kind: MeshGateway # leave open potential other selector kinds
  name: edge-gateway
 hostnames:
  - example.com
  - *.example.net
 rules: ...
```

Here as well the `spec.targetRef` tells us which proxy is being affected by
this configuration.

#### Motivation

If we imagine using the `spec.targetRef`/`to` structure for `MeshGateway`, we
have to explain to users that `to` is nonsensical in the gateway case.

In order to prevent confusion around how to target routing
policies, we create a separate policy with a different targeting structure, as
opposed to finding a more expressive, potentially confusing structure that can
be used for both gateways and services.

#### Hostnames

`MeshGatewayHTTPRoute` also includes a `hostnames` field for further targeting of specific hostnames
that the referenced `MeshGateway` serves.

Matching of hostnames follows the rules outlined in the gateway API
[`HTTPRoute.spec.hostnames`](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io%2fv1beta1.HTTPRoute)
field.

### Positive Consequences

- Route configuration can be very precisely targeted
- Users need to know only one schema for specifying rules
- A final spec that has the best of `MeshGatewayRoute` and `TrafficRoute`

### New rule spec

Routing configuration consists of a list of rules where each rule _matches_ some
requests or traffic, modifies it in some way with _filters_ and then sends it to a _backend_.

#### Matching

The new route matches requests based on:

- headers
- path
- method
- query parameters

##### Headers

Headers can be tested for:

- presence
- absence
- exact value
- prefix value
- regex value

##### Spec

A rule contains a list of match objects, each of which specifies a match for one
or more of the above types:

```yaml
matches:
 - method: POST
   path:
    type: RegularExpression|Exact|Prefix
    value: ...
   headers:
    - type: RegularExpression|Exact|Prefix|Absent
      name: "..."
      value: ...
   queryParams:
    - type: RegularExpression|Exact
      name: "..."
      value: ...
```

A given entry in `match` succeeds if _all_ of the conditions are satisfied and the rule matches when _any one_ of the entries under `match` succeeds. In other words, `match` is an OR'd list of AND'd conditions.

#### Request modification

With `filters` we can change:

- request headers
- response headers
- request mirroring (response is ignored)
- HTTP redirect
- host
- request path
- the response by directly returning one

```yaml
filters:
 - type: RequestHeaderModifier|ResponseHeaderModifier|RequestMirror|RequestRedirect|URLRewrite|DirectResponse
   requestHeaderModifier:
    set:
     - name:
       value:
    add:
     - name:
       value:
    remove:
     - "name"
   responseHeaderModifier:
    set:
     - name:
       value:
    add:
     - name:
       value:
    remove:
     - name
   requestMirror:
    backendRef:
     kind: MeshService
     name: backend-mirror
   requestRedirect:
    scheme:
    hostname:
    path:
    port:
    statusCode:
   urlRewrite:
    hostname:
    path:
     replaceFullPath:
     replacePrefixMatch:
   directResponse:
    status:
    body:
     inlineBytes: <base64>
     inlineString:
     filename:
     environmentVariable:
```

#### Routing

If `matches` succeeds, the request is routed to the specified weighted destinations.

The destinations appear as references in `backendRefs`.
In the case that a request should be routed to the service it would otherwise
have been routed to, `backendRefs` can be omitted.

```yaml
- backendRefs:
  - weight: 90
    kind: MeshServiceSubset
    name: backend
    tags:
      version: v2
  - weight: 10
    kind: MeshServiceSubset
    name: backend
    tags:
      version: v2
  ...
```

Note that it's not a valid configuration to have a `requestRedirect` or `directResponse`
and `backendRefs` specified or both `requestRedirect` and `urlRewrite`.

Note that in Gateway API, filters are also available on a _per-backend_ basis (`HTTPBackendRef.filters`).
We leave this option open for our policy in the future.

#### Load balancing

At the moment `TrafficRoute` can be used to configure load balancing behavior at
the top, rule-independent level. This MADR proposes **removing** load balancing from
the route policies and instead later creating a separate policy to handle load
balancing configuration.

The main motivator for this is that load balancing is done on an Envoy cluster.
Including load balancing options on a match would force creation of a new cluster for that
set of `backendRefs` which has negative consequences including but not limited to
metrics handling and a potentially large number of clusters.

#### Merging behavior

With `targetRef` resources, the ultimate configuration of a data plane proxy
is always a result of all policies targeting that proxy depending on their
specificity.

> :question: Is the current merge behavior what we need for routing policy? What surprises are there? See the next sections

This proposal addresses one issue with a "naive" application of `targetRef`
structure to a routing policy.

##### Proposal

This MADR proposes to distribute `default` down into individual `rules` and merge
`rules` based on _structural equality_ of the `matches` value.
So we would have:

```yaml
metadata:
 name: owner
spec:
 targetRef:
  kind: Mesh
 to:
  - targetRef:
     kind: MeshService
     name: backend
    rules:
     - matches:
        - path:
           prefix: /v1
       default:
        backendRefs:
         - weight: 100
           kind: MeshServiceSubset
           name: backend
           tags:
            version: v1
     - matches:
        path:
         prefix: /v2
       default:
        backendRefs:
         - weight: 100
           kind: MeshServiceSubset
           name: backend
           tags:
            version: v2
---
metadata:
 name: consumer
spec:
 targetRef:
  kind: MeshService
  name: frontend
 to:
  - targetRef:
     kind: MeshService
     name: backend
    rules:
     - matches:
        - path:
           prefix: /v2
       default:
        filters:
         - responseHeaderModifier:
            add:
             - name: env
               value: dev
```

which gives us the final list of rules:

```yaml
spec:
 targetRef:
  kind: MeshService
  name: frontend
 to:
  - targetRef:
     kind: MeshService
     name: backend
    rules:
      - matches:
         - path:
            prefix: /v1
        backendRefs:
         - weight: 100
           kind: MeshServiceSubset
           name: backend
           tags:
            version: v1
     - matches:
        - path:
           prefix: /v2
       filters:
        - requestHeaderModifier:
           add:
            - name: env
              value: dev
       backendRefs:
        - weight: 100
          kind: MeshServiceSubset
          name: backend
          tags:
           version: v2
```

Here, the `/v1` rule from `owner` is unchanged and the `/v2` rule from
`consumer` merges with that of `owner`.

What this structure doesn't allow is blanket overwriting of all less-specific rules
by a more specifically targeted resource. Instead, each rule needs to be
explicitly overriden:

```yaml
metadata:
 name: owner
spec:
 targetRef:
  kind: Mesh
 to:
  - targetRef:
     kind: MeshService
     name: backend
    rules:
     - matches:
        - path:
           prefix: /v1
       default:
        backendRefs:
         - weight: 100
           kind: MeshServiceSubset
           name: backend
           tags:
            version: v1
     - matches:
        - path:
           prefix: /v2
       default:
        backendRefs:
         - weight: 100
           kind: MeshServiceSubset
           name: backend
           tags:
            version: v2
---
metadata:
 name: consumer
spec:
 targetRef:
  kind: MeshService
  name: frontend
 to:
  - targetRef:
     kind: MeshService
     name: backend
    rules:
     - matches:
        - path:
           prefix: /v1
       default:
        filters: []
        backendRefs: []
     - matches:
        - path:
           prefix: /v2
       default:
        filters: []
        backendRefs: []
```

The next section shows the motivation for this schema.

##### Surprising behavior of top level `default`

Let's look at two basic theoretical route policies where we include `default` at
the top level of a `to` rule, i.e. like a "regular" `targetRef` policy:

```yaml
metadata:
 name: owner
spec:
 targetRef:
  kind: Mesh
 to:
  - targetRef:
     kind: MeshService
     name: backend
    default:
     rules:
      - matches:
         - path:
            prefix: /v1
        backendRefs:
         - weight: 100
           kind: MeshServiceSubset
           name: backend
           tags:
            version: v1
      - matches:
         - path:
            prefix: /v2
        backendRefs:
         - weight: 100
           kind: MeshServiceSubset
           name: backend
           tags:
            version: v2
---
metadata:
 name: consumer
spec:
 targetRef:
  kind: MeshService
  name: frontend
 to:
  - targetRef:
     kind: MeshService
     name: backend
    default:
     rules:
      - matches:
         - path:
            prefix: /v2
        filters:
         - requestHeaderModifier:
            add:
             - name: env
               value: dev
        backendRefs:
         - weight: 100
           kind: MeshServiceSubset
           name: backend
           tags:
            version: v2
```

The rules for merging a field in a `targetRef` policy is that more specific
routes replace the value from less specific routes. So the final value of `rules`
is the value from the `consumer` route:

```yaml
spec:
 targetRef:
  kind: MeshService
  name: frontend
 to:
  - targetRef:
     kind: MeshService
     name: backend
    rules:
     - matches:
        - path:
           prefix: /v2
       filters:
        - requestHeaderModifier:
           add:
            - name: env
              value: dev
       backendRefs:
        - weight: 100
          kind: MeshServiceSubset
          name: backend
          tags:
           version: v2
```

This is likely surprising, especially because `consumer` didn't specify anything
at all for `/v1`.

#### Gateway API

A note about Gateway API and our routes. The only extension point for `HTTPRoute` is `filters`, so any additional configuration
must appear there.

With [`ExtensionRef`](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1beta1.HTTPRouteFilterType)
a filter is powerful, flexible and extensible enough to support any arbitrary routing configuration.
It doesn't really assume anything about the referenced resource.

The Gateway API controller would handle converting an `HTTPRoute` with an `ExtensionRef` filter to `MeshHTTPRoute`.

#### Final top level spec

```yaml
spec:
 targetRef:
  kind: MeshService
  name: frontend
 to:
  - targetRef:
     kind: MeshService
     name: backend
    rules:
     - matches:
        - ...
        - ...
       default:
        filters: [...]
        backendRefs: [..]
```

### Additional use cases

The following uses cases and examples are kept for the record.
They discuss potential issues of the policies in this proposal.

Consider:

- the owner of a service `backend`
  may want to configure a routing policy
  for their service that applies to requests from all services.
- another service owner may want
  to apply routing to requests from their service `frontend` to `backend`.

How should these routes be merged? One invariant might be that the consumer of a
service shouldn't be able to override the entire route of the service owner.

Given a policy from the owner that configures canarying and a filter:

```yaml
metadata:
 name: owner
spec:
 targetRef:
  kind: Mesh
 to:
  - targetRef:
     kind: MeshService
     name: backend
    default:
     rules:
      - matches:
         - path:
            prefix: /
        filters:
         - responseHeader:
            add:
             - name: X-Some-Header
               value: something
        backendRefs:
         - weight: 90
           kind: MeshServiceSubset
           name: backend
           tags:
            version: v1
         - weight: 10
           kind: MeshServiceSubset
           name: backend
           tags:
            version: v2
```

and a route by the `frontend` owner that configures an additional filter:

```yaml
metadata:
 name: consumer
spec:
 targetRef:
  kind: MeshService
  name: frontend
 to:
  - targetRef:
     kind: MeshService
     name: backend
    default:
     rules:
      - matches:
         - path:
            prefix: /
        filters:
         - requestHeader:
            add:
             - name: X-Client
               value:
        backendRefs:
         - weight: 100
           kind: MeshServiceSubset
           name: backend
           tags:
            version: v1
```

when these resources are merged the latter will have higher priority and
override the less specific resource:

```yaml
metadata:
 name: consumer
spec:
 targetRef:
  kind: MeshService
  name: frontend
 to:
  - targetRef:
     kind: MeshService
     name: backend
    default:
     rules:
      - matches:
         - path:
            prefix: /
        filters:
         - requestHeader:
            add:
             - name: X-Client
               value:
        backendRefs:
         - weight: 100
           kind: MeshServiceSubset
           name: backend
           tags:
            version: v1
```

Two issues:

1. the `backendRefs` of `owner` are gone. But maybe the owner wants final say
   over `backendRefs`.
1. the `filters` of `owner` are gone. But maybe the owner of `backend` can tolerate additional filters?

For 1., we need a way for the `owner` policy to set the final value.
This is a good use case for [`override`](https://gateway-api.sigs.k8s.io/v1alpha2/references/policy-attachment/#hierarchy)?

For 2., we would need a way for the `owner` policy to say it's OK to add to this
list of filters. Could we add a `merge` next to `default`?

## Implementation

Almost all other plugins will depend on the route configuration generated
by `MeshHTTPRoute` resources.
Though this policy is a new, `targetRef`-based policy, this means it will not be a
`PolicyPlugin` because the current plugin interface of modifying existing Envoy
resources is insufficient for writing such a fundamental piece of config
generation.
Instead it will likely have to be integrated directly in generation similar to
the current `TrafficRoute` resources.

## Considered options

This section goes over alternatives we considered for this MADR.

### Two routing policies or one routing policies

This section considers different ways of handling the overlap between
`MeshGateway` routing and service to service routing.

#### `MeshGateway`

This new resource is intended to replace the `MeshGatewayRoute` resource for
configuring routes for a `MeshGateway`.

The core difference is that routes for a `MeshGateway` are fundamentally different from in-mesh
routes because there's a proxy "in between" requests. That is, when I send a
request from service to service, the Envoy configuration and routing happens on the source
proxy side. When I send a request to a `MeshGateway`, the request reaches the
`MeshGateway` proxy and _then_ gets routed, i.e. they're on the destination
side.

There are two different ways we can express a `MeshGateway` attachment given a
unified `MeshHTTPRoute` resource. Either as
`spec.targetRef` & `to.targetRef` or `from.targetRef` & `spec.targetRef`.

An additional option is to instead have two differnet "top-level" policies that
both point to a new resource `HTTPRouteRules`, which holds the routes as
previously described.

##### Only `spec.targetRef`

We support attaching routes to `MeshGateway` with `spec.targetRef` and simply
not support any targeting via `from` or `to` in this case. This would only prevent
targeting routes for cross-mesh gateways based on the source mesh.

###### Edge and cross-mesh gateways

```yaml
spec:
 targetRef:
  kind: MeshGateway
  name: edge-or-cross-mesh-gateway
 default:
  rules: ...
```

where top level `default` is allowed only for `spec.targetRef.kind: MeshGateway`.

##### No more `spec.targetRef`

We don't need to have a strict `spec.targetRef` policy. Routes are fundamentally
different from other policies. We can use only `from.targetRef` and `to.targetRef`.
Both, one or none of these could be lists as well.

###### In-mesh

```yaml
from:
 targetRef:
  kind: MeshService
  name: frontend
to:
 targetRef:
  kind: MeshService
  name: backend
```

###### Edge gateway

In order to configure a non-cross-mesh `MeshGateway` resource,
the user points `to.targetRef` to a `MeshGateway`.

```yaml
to:
 targetRef:
  kind: MeshGateway
  name: edge-gateway
```

###### Cross-mesh gateway

For cross-mesh `MeshGateways`:

```yaml
to:
 targetRef:
  kind: MeshGateway
  name: cross-mesh-gateway
```

would target _any requests_ from _any `Mesh`_ made to this `MeshGateway` whereas
`kind: Mesh` must be used to target a specific `Mesh` as source:

```yaml
from:
 targetRef:
  kind: Mesh
  name: other-mesh
to:
 targetRef:
  kind: MeshGateway
  name: cross-mesh-gateway
```

##### `MeshGateway` as `to.targetRef`

This is a somewhat different way of thinking about `targetRef` because here the
`to.targetRef` points to the Envoy whose configuration is changing, as opposed
to `spec.targetRef`. However, a `MeshGateway` proxy is fundamentally different from a
sidecar proxy because it does not attach to a service, so perhaps this is
justified?

###### Edge gateway

In order to configure a non-cross-mesh `MeshGateway` resource,
one option is that the user points `to.targetRef` to a `MeshGateway`.

```yaml
targetRef:
 kind: Any
to:
 - targetRef:
    kind: MeshGateway
    name: edge-gateway
```

the `spec.targetRef` of `kind: Any` can be read as meaning "any
connections made to the `MeshGateway`". Only `kind: Any` is permitted with a
`to.targetRef` of `kind: MeshGateway`.

###### Cross-mesh gateway

For cross-mesh `MeshGateways`:

```yaml
targetRef:
 kind: Any
to:
 - targetRef:
    kind: MeshGateway
    name: mesh-gateway
```

would target _any requests_ from _any `Mesh`_ made to this `MeshGateway` whereas
`kind: Mesh` must be used to target a specific `Mesh` as source:

```yaml
targetRef:
 kind: Mesh
 name: other-mesh
to:
 - targetRef:
    kind: MeshGateway
    name: mesh-gateway
```

##### `MeshGateway` in `spec.targetRef`

If we put `MeshGateway` in `spec.targetRef`, we could instead use `from`.

###### Edge gateway

In order to configure a non-cross-mesh `MeshGateway` resource:

```yaml
targetRef:
 kind: MeshGateway
 name: edge-gateway
from:
 - targetRef:
    kind: Any
```

the `from.targetRef` of `kind: Any` can be read as "any
connections made to the `MeshGateway`". Only `kind: Any` is permitted as a
`from.targetRef` with `spec.targetRef` of `MeshGateway`.

###### Cross-mesh gateway

For cross-mesh `MeshGateways`:

```yaml
targetRef:
 kind: MeshGateway
 name: mesh-gateway
from:
 - targetRef:
    kind: Any
```

would target _any requests_ from _any `Mesh`_ made to this `MeshGateway` whereas
`kind: Mesh` must be used to target a specific `Mesh` as source:

```yaml
targetRef:
 kind: MeshGateway
 name: mesh-gateway
from:
 - targetRef:
    kind: Mesh
    name: other-mesh
```

##### Different top-level policies

A different option would be to have two top-level policies that both point to a
new `HTTPRouteRules` resource. This resource would hold the rules themselves, whose schema
is described already.

One advantage of this is that it becomes possible to combine rules without
relying on the behavior of `targetRef` specificity:

```yaml
---
kind: HTTPRouteRules
metadata:
  name: owner-rules
spec:
  rules:
  - matches:
    ...
  - matches:
    ...
---
kind: MeshTrafficRoute
spec:
 targetRef:
  kind: Mesh
 to:
  - targetRef:
     kind: MeshService
     name: backend
    default:
      rulesRefs:
      - owner-rules
---
kind: MeshGatewayRoute
spec:
 targetRef:
  kind: MeshGateway
  name: cross-mesh
 from:
  - targetRef:
     kind: Mesh
     name: other-mesh
    default:
     rulesRefs:
     - owner-rules
```

```yaml
---
kind: HTTPRouteRules
metadata:
  name: consumer-rules
spec:
  rules:
  - matches:
    ...
  - matches:
    ...
---
kind: MeshTrafficRoute
spec:
 targetRef:
  kind: MeshService
  name: frontend
 to:
  - targetRef:
     kind: MeshService
     name: backend
    default:
     rulesRefs:
     - owner-rules
     - consumer-rules
```
