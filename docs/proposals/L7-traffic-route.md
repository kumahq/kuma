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
  # split: section is no longer allowed when http: section is present
  http:
  - match:
      method: GET
      path: # one of either prefix or exact will be allowed
        prefix: /users
        exact: /users/user-1
      headers:
        some-header: # one of either prefix or exact will be allowed
          exact: some-value
          prefix: some-
    split:
      - weight: 90
        destination:
          kuma.io/service: usr_svc_6379
          version: '1.0'
      - weight: 10
        destination:
          kuma.io/service: usr_svc_6379
          version: '2.0'
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
    split:
      - destination:
          kuma.io/service: backend
          version: '1.1'
  - split: # notice that we don't have match = which means it always matches
      - destination:
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
    split:
      - destination:
          kuma.io/service: offers
  - split:
      - destination:
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
  - split:
      - destination:
          kuma.io/service: '*'
```

## Matching

`TrafficRoute` will preserve same behavior of matching, meaning that we will only take into account `sources` and `destinations`
for matching and then take `http` section from it.
