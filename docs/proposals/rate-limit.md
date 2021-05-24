# Rate Limit

## Context

Today the only way to support Rate Limiting using Kuma – create a ProxyTemplate. Current proposal introduces a new 
policy Rate Limit.

## Requirements

Provide a way to configure local rate limit both for [TCP](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/network/local_ratelimit/v3/local_rate_limit.proto#envoy-v3-api-msg-extensions-filters-network-local-ratelimit-v3-localratelimit) and [HTTP](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/local_ratelimit/v3/local_rate_limit.proto#envoy-v3-api-msg-extensions-filters-http-local-ratelimit-v3-localratelimit).

## Configuration

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
    max: 1000 # if 'max' is not specified we can take it 10 times more than 'connections'
  http:
	  requests: 100 # how many connections are allowed per 'period'
	  period: 1s	
    max: 1000 # if 'max' is not specified we can take it 10 times more than 'requests'
    status: 401
    headers:
     - key: x-local-rate-limit
       value: 'true'
       append: false
```

## Implementation details

Since it is a local rate limiting we have to set it on the destination side. It could be challenging to support 
`source` selector, but still quite possible for HTTP. We already set header `x-kuma-tags` to identify the source for
FaultInjections, I believe the same approach will work for RateLimit. For TCP most likely we’ll have a limitation 
on `source` selector unless it’s possible to somehow use SANs but still this is mTLS only. 
