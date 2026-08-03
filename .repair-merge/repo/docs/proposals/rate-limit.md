# Rate Limit

## Context

Today the only way to support Rate Limiting using Kuma – create a ProxyTemplate. Current proposal introduces a new 
policy Rate Limit.

## Requirements

Provide a way to configure local rate limit both for [TCP](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/network/local_ratelimit/v3/local_rate_limit.proto#envoy-v3-api-msg-extensions-filters-network-local-ratelimit-v3-localratelimit) and [HTTP](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/local_ratelimit/v3/local_rate_limit.proto#envoy-v3-api-msg-extensions-filters-http-local-ratelimit-v3-localratelimit).

## Configuration

### Design 1

```yaml
type: RateLimit
name: rt-1
mesh: default
sources:
- match:
    kuma.io/service: web
destinations:
- match:
    kuma.io/service: backend
conf:
  tcp:
    connections: 100 # how many connections are allowed per 'period'
    period: 1s
  http:
    requests: 100 # how many requests are allowed per 'period'
    period: 1s
    status: 401
    headers:
     - key: x-local-rate-limit
       value: 'true'
       append: false
```

Configuration of the rate-limiter happens on the `source` side. Such kind of the rate-limiting allows user to limit the number
of the outgoing requests. If you put 100 RPS between `web` and `backend` that does **NOT** mean `backend` will be limited 
by 100 request per second. It means every instance of `web` won't be making more than 100 request per second (i.e. if you have 
10 instances of `web` then `backend` will be handling 1000 request per second).

### Design 2

```yaml
type: RateLimit
name: rt-1
mesh: default
sources:
- match:
    kuma.io/service: web
destinations:
- match:
    kuma.io/service: backend
conf:
  tcp:
    connections: 100 # how many connections are allowed per 'period'
    period: 1s
  http:
    requests: 100 # how many requests are allowed per 'period'
    period: 1s
    status: 401
    headers:
     - key: x-local-rate-limit
       value: 'true'
       append: false
```

Same as [Design 1](#Design 1), but rate-limiter is configured on the `destination` side. This allows user to protect 
`backend` from being DDoS'ed. If you put 100 RPS between `web` and `backend` in this case it's guaranteed that every
instance of the `backend` service won't receive more than 100 requests per second.

Downsides: should work only with mTLS for 2 reasons:
1. In order to support `sources` selector for TCP we have to encode information about the service into SNI (we already do it for Zone Ingress)
2. Security. If you set quotas 10 RPS for `web` and 1000 RPS for `offers` then `web` service should not be able to impersonate `offers` and bypass quotas.

### Design 3

```yaml
type: RateLimit
name: rt-1
mesh: default
selectors:
  - match:
      kuma.io/service: backend # apply on backend for any traffic
conf:
  tcp:
    connections: 100 # how many connections are allowed per 'period'
    period: 1s
```

This design allows user to protect `backend` from being DDoS'ed, but doesn't have an option to select by `source` service.
In other words, no way to configure quotas per `service` like in [Design 2](#Design 2)
