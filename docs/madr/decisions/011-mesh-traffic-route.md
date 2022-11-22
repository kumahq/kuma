# `MeshTrafficRoute`

- Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/4743
Other issues covered:
- https://github.com/kumahq/kuma/issues/4691
- https://github.com/kumahq/kuma/issues/4690
- https://github.com/kumahq/kuma/issues/3456
- https://github.com/kumahq/kuma/issues/5309

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
the top, rule-independent level. This MADR proposes load balancing as a _filter_.
This way it can be applied on a _per-route_ basis as well as a _per-backend_ basis.

```yaml
spec:
  default:
    rules:
    - filters:
      - type: LoadBalancing
        loadBalancing: 
          localityAware: true
          type: RoundRobin|LeastRequest|RingHash|Random|Maglev
          leastRequest:
            choiceCount: 8
          ringHash:
            hashFunction: "MURMUR_HASH_2"
            minRingSize: 64
            maxRingSize: 1024
            hashPolicies:
              - # oneof
                header:
                  name:
                  regexRewrite:
                cookie:
                  name:
                  ttl:
                  path:
                connectionProperties:
                  sourceIP: <bool>
                queryParameter:
                  name:
                filterState:
                  key:
                terminal: <bool>
          maglev:
            tableSize: 65537
            hashPolicies: ... # same schema as above
```

##### Gateway API

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

#### Merging behavior

With `targetRef` resources, the ultimate configuration is a result of overriden/merged policies depending on their specificity.

> :question: Is the current merge behavior what we need for routing policy? What surprises are there? See the next sections

##### Service owner routes

There's a discussion in the Gateway API repository about what different types of
routes exist and how they should interact. Consider:

- the owner of a service `svc_a`
may want to configure a routing policy for their service that applies to requests from all services.
- another service owner may want to apply routing to requests from their service
to `svc_a`.

How should these routes be merged? One invariant might be that the consumer of a
service shouldn't be able to override the routes from the service owner.

The route that applies to all requests:

```yaml
metadata:
  name: owner
spec:
 targetRef:
   kind: Mesh
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
        - responseHeader:
            add:
            - name: X-Some-Header
              value: something
        backendRefs:
        - weight: 90
          kind: MeshServiceSubset
          name: svc_a
          tags:
            version: v1
        - weight: 10
          kind: MeshServiceSubset
          name: svc_a
          tags:
            version: v2
```

The route that applies to requests from a specific service:

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

But the service owner probably expects to somehow
be able to "enforce" routes on consumers that can't be overriden.

Perhaps we finally need [`override`](https://gateway-api.sigs.k8s.io/v1alpha2/references/policy-attachment/#hierarchy)?

##### Merging

Is there a use case for being able to write a policy that combines with the above `owner` policy such that the final resource would have both `filters`?

Perhaps given the above `owner` policy along with a different `svc_a` specific
policy:

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
```

a user might expect:

```yaml
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
        - responseHeader:
            add:
            - name: X-Some-Header
              value: something
        backendRefs:
        - weight: 100
          kind: MeshServiceSubset
          name: svc_a
          tags:
            version: v1
```

Could we achieve this by always combining `filters` using `matches` as a merge key
with structural equality and
specifying `default`/`override` and maybe even `merge` for `filters` _next to_
`matches`?
It probably doesn't make sense to merge `backendRefs` so they should be
limited to `default`/`override`.

Given:

```yaml
metadata:
  name: owner
spec:
 targetRef:
   kind: Mesh
 to:
  - targetRef:
      kind: MeshService
      name: svc_a
    rules:
    - matches:
        path:
          prefix: /
      merge:
        filters:
        - responseHeader:
            add:
            - name: X-Some-Header
              value: something
      override:
        backendRefs:
        - weight: 90
          kind: MeshServiceSubset
          name: svc_a
          tags:
            version: v1
        - weight: 10
          kind: MeshServiceSubset
          name: svc_a
          tags:
            version: v2
```

along with:

```yaml
metadata:
  name: owner
spec:
 targetRef:
   kind: MeshService
   name: svc_b
 to:
  - targetRef:
      kind: MeshService
      name: svc_a
    rules:
    - matches:
        path:
          prefix: /
      merge:
        filters:
        - requestHeader:
            add:
            - name: X-Client
              value:
```

would give us the rules:

```yaml
spec:
 targetRef:
   kind: MeshService
   name: svc_b
 to:
  - targetRef:
      kind: MeshService
      name: svc_a
    rules:
    - matches:
        path:
          prefix: /
      filters:
      - responseHeader:
          add:
          - name: X-Some-Header
            value: something
      - requestHeader:
          add:
          - name: X-Client
            value:
      backendRefs:
      - weight: 90
        kind: MeshServiceSubset
        name: svc_a
        tags:
          version: v1
      - weight: 10
        kind: MeshServiceSubset
        name: svc_a
        tags:
          version: v2
```

Both `filters` are applied and the `backendRefs` are taken from the
less-specific route.

#### Final spec

TODO
