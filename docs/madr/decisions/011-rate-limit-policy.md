# Rate limit policy

- Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/4740

## Context and Problem Statement

We want to create a [new policy matching compliant](https://github.com/kumahq/kuma/blob/22c157d4adac7f518b1b49939c7e9ea4d2a1876c/docs/madr/decisions/005-policy-matching.md)
resource for managing rate limiting.

Rate Limiting in Kuma is implemented with Envoy's [rate limiting
support](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/local_rate_limit_filter).
There is an issue with configuration of TCP rate limiting, because [current filter](https://www.envoyproxy.io/docs/envoy/latest/configuration/listeners/network_filters/local_rate_limit_filter#config-network-filters-local-rate-limit) works on listener level which can only 
allows to limit connection on specific listener but without checking the origin of the request.

## Decision Drivers

- Replace existing policies with new policy matching compliant resources
- Add support for TCP connection limiting
- Should we add support for global rate limiting?

## Considered Options

### Naming

- MeshRateLimit

## Solution

### Current configuration
Below is sample rate limiting configuration.

```yaml
spec:
  sources:
    - match:
        kuma.io/service: "redis_kuma-demo_svc_6379"    
  destinations:
    - match:
        kuma.io/service: "demo-app_kuma-demo_svc_5000"
  conf:
    http:
      requests: 5
      interval: 10s
      onRateLimit:
        status: 423
        headers:
          - key: "x-kuma-rate-limited"
            value: "true"
            append: true
```
We are using [Envoy's local rate limit](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/local_rate_limit_filter) which support only HTTP configuration.

The configuration translate to the Envoy configuration on route:

```yaml
{
  "match": {
    "prefix": "/",
    "headers": [
    {
      "name": "x-kuma-tags",
      "safe_regex_match": {
      "google_re2": {},
      "regex": ".*&kuma.io/service=[^&]*redis_kuma-demo_svc_6379[,&].*"
      }
    }
    ]
  },
  "route": {
    "cluster": "localhost:5000",
    "timeout": "0s"
  },
  "typed_per_filter_config": {
    "envoy.filters.http.local_ratelimit": {
    "@type": "type.googleapis.com/envoy.extensions.filters.http.local_ratelimit.v3.LocalRateLimit",
    "stat_prefix": "rate_limit",
    "status": {
      "code": "Locked"
    },
    "token_bucket": {
      "max_tokens": 5,
      "tokens_per_fill": 5,
      "fill_interval": "10s"
    },
    "filter_enabled": {
      "default_value": {
      "numerator": 100
      },
      "runtime_key": "local_rate_limit_enabled"
    },
    "filter_enforced": {
      "default_value": {
      "numerator": 100
      },
      "runtime_key": "local_rate_limit_enforced"
    },
    "response_headers_to_add": [
      {
      "header": {
        "key": "x-kuma-rate-limited",
        "value": "true"
      },
      "append": true
      }
    ]
    }
  }
}

```

### Specification

Rate limiting can be configured on both HTTP connection managers and routes but we can simplify it and configure it just on the route. Thanks to header `kuma.io/service` which is propagated and set by Envoy we are able to recognize origin of the request.

#### Top level

Top-level targetRef can have all available kinds:
```yaml
targetRef:
  kind: Mesh|MeshSubset|MeshService|MeshServiceSubset|MeshGatewayRoute|MeshHTTPRoute
  name: ...
```
Rate limiting is an inbound policy so only `from` should be configured.

#### From level

```yaml
from:
  - targetRef:
      kind: Mesh|MeshSubset|MeshService|MeshServiceSubset
      name: ...
```

Matching on MeshGatewayRoute and MeshHTTPRoute does not make sense (there is no route that a request originates from).

#### Configuration

```yaml
  conf:
    http:
      requests: 5
      interval: 10s
      onRateLimit:
        status: 423
        headers:
          - key: "x-kuma-rate-limited"
            value: "true"
            append: true
```
