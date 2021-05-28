# L7 Traffic Routing

## Context

Kuma provides traffic management with `TrafficRoute` policy based on Kuma tags.
We can split the traffic to different destinations, we can redirect the traffic etc.

## Requirements

Provide a way to manage the traffic based on HTTP headers, path and method.

## Configuration

```yaml
type: TrafficRoute
name: route-all-default
mesh: default
sources:
- match:
    kuma.io/service: web
destinations:
- match:
    kuma.io/service: backend
conf:
  http: # a new http section which is a list
  - match: # match is required and need to have at least one sub-element
      method: GET
      path: # one of either prefix, exact or regex will be allowed
        prefix: /users
        exact: /users/user-1
        regex: 
      headers:
        some-header: # one of either prefix, exact or regex will be allowed
          exact: some-value
          prefix: some-
          regex
    modify: # optional section
      path: # either rewritePrefix or regex
        rewritePrefix: /not-users # validation that it can be used only if there is prefix in match
        regex: # (example to change the path from "/service/foo/v1/api" to "/v1/api/instance/foo")
          pattern: "^/service/([^/]+)(/.*)$"
          substitution: "\2/instance/\1"
      host: # either value or fromPath
        value: "XYZ"
        fromPath: # (example to extract "envoyproxy.io" host header from "/envoyproxy.io/some/path" path)
          pattern: "^/(.+)/.+$"
          substitution: "\1"
      requestHeaders:
        add:
          - name: x-custom-header
            value: xyz
            append: true # if true then if there is x-custom-header already, it will append xyz 
        remove:
          - name: x-something
      responseHeaders:
        add:
          - name: x-custom-header
            value: xyz
            append: true
        remove:
          - name: x-something
    destination: # required either split or a destination. Destination is a syntax sugar over split to one destination with weight 100 
      kuma.io/service: usr_svc_6379
      version: '1.0'
    split:
      - weight: 90
        destination:
          kuma.io/service: usr_svc_6379
          version: '1.0'
      - weight: 10
        destination:
          kuma.io/service: usr_svc_6379
          version: '2.0'
  destination: # if we have a http section and we don't match anything, then this rule is applied. It's again destination or split
    kuma.io/service: usr_svc_6379
    version: '1.0'
  split:
    - weight: 100
      destination:
        kuma.io/service: usr_svc_6379
        version: '1.0'
```

### More examples

1) Canary deployment

**Use case:** when we release something on prod, but want to check functionality before serving it to all the users

When any service is trying to consume `backend` with header `canary: "true"` the traffic is redirected to 1.1 backend, otherwise the traffic is passed to a 1.0 backend.

```yaml
type: TrafficRoute
name: canary
mesh: default
sources:
- match:
    kuma.io/service: '*'
destinations:
- match:
    kuma.io/service: backend
conf:
  http:
  - match:
      headers:
        canary:
          exact: "true"
    destination:
      kuma.io/service: backend
      version: '1.1'
  destination:
    kuma.io/service: backend
    version: '1.0'
```

2) Redirecting traffic to a different service

**Use case:** Extracting a new `offers` microservice from `backend`, we can gradually shift the traffic with the `TrafficRoute`

```yaml
type: TrafficRoute
name: reroute
mesh: default
sources:
- match:
    kuma.io/service: '*'
destinations:
- match:
    kuma.io/service: backend
conf:
  http:
  - match:
      path:
        prefix: "/offers"
    destination:
      kuma.io/service: offers
  destination:
    kuma.io/service: backend
```

3) Generic canary

With the "traditional" `split` there is support for `*` in the tags. We should also support this use case in `http`.
So this definition is similar to 1) but more generic.

```yaml
type: TrafficRoute
name: canary
mesh: default
sources:
- match:
    kuma.io/service: '*'
destinations:
- match:
    kuma.io/service: '*'
conf:
  http:
  - match:
      headers:
        canary:
          exact: "true"
    split:
      - destination:
          kuma.io/service: '*'
          version: canary
  destination:
    kuma.io/service: '*'
```

## Other notes

* Split will require weight

## Matching

`TrafficRoute` will preserve same behavior of matching, meaning that we will only take into account `sources` and `destinations`
for matching and then take `http` section from it.

## Protocol

If we have `http` section, it will be only applied on HTTP traffic.
If service is marked as TCP, then we won't match and the default `destination` is applied.
