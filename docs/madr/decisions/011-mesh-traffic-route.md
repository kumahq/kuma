# `MeshTrafficRoute`

- Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/4743
Other issues covered:

- https://github.com/kumahq/kuma/issues/4691
- https://github.com/kumahq/kuma/issues/4690

## Context and Problem Statement

Now that we have the [new policy matching resources](https://github.com/kumahq/kuma/blob/master/docs/madr/decisions/005-policy-matching.md), it's time to create one for traffic routung.
We currently have both `TrafficRoute` and `MeshGatewayRoute`.

This MADR does not address Gateway API integration per sebut does consider Gateway
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

## Decision Outcome

The new resource routes and alters requests depending on where it's coming from and where it's going to. Routes end up in the client-side Envoy config so we choose a `spec.to` resource.

```yaml
spec:
 targetRef: ... # from where requests are coming
 to:
  - targetRef: ... # to where requests are going
 default: ...
```

Each protocol gets its own resource:

- `MeshHTTPRoute`
- `MeshTCPRoute`

which can be extended with others like GRPC.

The route that's used for a given request is the most specific one that Kuma can
apply given the `kuma.io/protocol` tag.

### Positive Consequences

- Route configuration can be very precisely targeted
- Users need know only one set of resources for routing
- A final spec that has the best of `MeshGatewayRoute` and `TrafficRoute`

### New `MeshHTTPRoute` spec

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
 - methods:
    - GET
    - POST
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

A given entry in `match` succeeds if _all_ of the conditions are satisfied and the rule matches when _any one_ of the entries under `match` suceeds. In other words, `match` is an OR'd list of AND'd conditions.

#### Request modification

With `filters` we can change:

- request headers
- response headers
- request mirroring (response is ignored)
- HTTP redirect
- host
- request path

```yaml
filters:
 - type: RequestHeaderModifier|ResponseHeaderModifier|RequestMirror|RequestRedirect|URLRewrite
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
     name: svc_name
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
    name: svc_name
    tags:
      version: v2
  - weight: 10
    kind: MeshServiceSubset
    name: svc_name_other
    tags:
      version: v2
  ...
```

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

##### Surprising behavior of top level `default`

Let's look at two basic theoretical route policies where we include `default` at
the top level of a `to` rule:

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
         path:
          prefix: /v1
        backendRefs:
         - weight: 100
           kind: MeshServiceSubset
           name: backend
           tags:
            version: v1
      - matches:
         path:
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
         path:
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
        path:
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

This MADR proposes to distribute `default` down into individual `rules` and merge
`rules` based on _structural equality_ of the `matches` value.
So we would instead have:

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
         path:
          prefix: /v1
        backendRefs:
         - weight: 100
           kind: MeshServiceSubset
           name: backend
           tags:
            version: v1
      - matches:
         path:
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
         path:
          prefix: /v2
        filters:
         - requestHeaderModifier:
            add:
             - name: env
               value: dev
        backendRefs:
         - weight: 100
           kind: MeshService
           name: backend
           tags:
            version: v2
```

which gives us the rules:

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
         path:
          prefix: /v1
        backendRefs:
         - weight: 100
           kind: MeshServiceSubset
           name: backend
           tags:
            version: v1
     - matches:
        path:
         prefix: /v2
       filters:
        - requestHeaderModifier:
           add:
            - name: env
              value: dev
       backendRefs:
        - weight: 100
          kind: MeshService
          name: backend
          tags:
           version: v2
```

Here, the `/v1` rule from `owner` is unchanged and the `/v2` rule from
`consumer` overrides that of `owner`.

##### Additional use cases

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
         path:
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
         path:
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
  name: svc_b
 to:
  - targetRef:
     kind: MeshService
     name: svc_a
    default:
     rules:
      - matches:
         path:
          prefix: /
        filters:
         - requestHeader:
            add:
             - name: X-Client
               value:
        backendRefs:
         - weight: 100
           kind: MeshServiceSubset
           name: svc_a
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

#### Gateway API

There is no other extension point for `HTTPRoute` than `filters` so any additional configuration
must appear under `filters`.

With [`ExtensionRef`](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1beta1.HTTPRouteFilterType)
a filter is powerful, flexible and extensible enough to support any arbitrary routing configuration.
It doesn't really assume anything about the referenced resource.

For example, to use round robin for all requests:

```yaml
spec:
  ...
  rules:
    - matches: []
      filters:
        - type: ExtensionRef
          extensionRef:
            group: kuma.io/v1alpha1
            kind: RouteConfig
            name: load-balancer-common
            # potential inlined load-balancer-common:
            # spec:
            #   loadBalancer: RoundRobin
      backendRefs: []
```

The Gateway API controller would handle converting these resources to
`MeshHTTPRoute`.

#### Final spec

TODO
