# Traffic Mirroring

## Context

Traffic Mirroring is a useful feature which sends a copy of live traffic to the mirrored service which will help in testing a service with real production traffic.The difference is that the reply from the mirrored service is dropped by the envoy proxy and not returned to the caller service.

	    a                      b                     c
	|----------|             |-----------|    mirror   |------------|
	| service A| ----------> | Service B | ----------> |  Service C |
	|----------|             |-----------|             |------------|

Envoy provides full support of traffic mirroring [Envoy's Traffic Mirror](https://www.envoyproxy.io/docs/envoy/latest/api-v2/api/v2/route/route_components.proto#envoy-api-msg-route-routeaction-requestmirrorpolicy),
so to support the same functionality in kuma.


## Proposed configuration model

We need to extend `TrafficRoute` policy as below

```yaml
type: TrafficRoute
mesh: default
metadata:
  name: route-example
spec:
  sources:
    - match:
        kuma.io/service: backend_default_svc_80
  destinations:
    - match:
        kuma.io/service: redis_default_svc_6379
  conf:
    split:
      - weight: 100
        destination:
          kuma.io/service: redis_default_svc_6379
          version: '1.0'
        mirror:
        - percentage: 50 # percentage of traffic to be sent to the mirrored service.
          destination:
            kuma.io/service: redis_default_svc_6379
            version: '2.0'
```

## Notes
- Original issue: https://github.com/kumahq/kuma/issues/680