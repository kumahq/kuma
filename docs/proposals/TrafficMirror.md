# Traffic Mirroring

## Context

Traffic Mirroring is a useful feature which sends a copy of live traffic to the mirrored service which will help in testing a service with real production traffic.The difference is that the reply from the mirrored service is dropped by the envoy proxy and not returned to the caller service.

        a                      b                     c
    |-----------|             |-----------|    mirror   |------------|
    | service A | ----------> | Service B | ----------> |  Service C |
    |---------- |             |-----------|             |------------|
## Use cases

1. **As a** service owner

   **I want to** illuminate the service under test with production traffic

   **so that** it help reveal bugs that went untested during functional tests and experiment with new features.

2. **As a** service owner

   **I want to** test the service under test with production traffic

   **so that** I can avoid canary deployment.
## Envoy support

Envoy allows to configure `route.request_mirror_policies` on each Route:
```json
{
  "request_mirror_policies": []
}
```
Where `request_mirror_policies` is an array of elements:
```json
{
  "cluster": "...",
  "runtime_fraction": "{...}",
  "trace_sampled": "{...}"
}
```
So traffic mirror in envoy accepts a list of cluster name and the amount of traffic to mirror.[Traffic Mirror](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto.html#envoy-v3-api-msg-config-route-v3-routeaction-requestmirrorpolicy)
## Proposed configuration model

```yaml
apiVersion: kuma.io/v1alpha1
kind: TrafficRoute
mesh: default
metadata:
  name: route-default
spec:
  sources:
    - match:
        kuma.io/service: sleep_mirror-demo_svc_80
        app: sleep
  destinations:
    - match:
        kuma.io/service: httpbin_mirror-demo_svc_8080
  conf:
    http:
      - match:
          path:
            prefix: "/"
        split:
          - weight: 100
            destination:
              kuma.io/service: httpbin_mirror-demo_svc_8080
              version: v1
          - weight: 0
            destination:
              kuma.io/service: httpbin_mirror-demo_svc_8080
              version: v2
        mirror:
          destination:
            kuma.io/service: httpbin_mirror-demo_svc_8080
            version: v2
          percentage: 90
    destination:
      kuma.io/service: httpbin_mirror-demo_svc_8080
      version: v1
```

## What needs to be done

- Add the configuration to the TrafficRoute resource
- Generate seperate clusters from the above configuration with properties name, service, tags.
- Add the traffic mirror envoy policy in the route from the clusters identified as mirror cluster.

## Notes

- Envoy traffic mirror issue with `runtime_fraction` (https://github.com/envoyproxy/envoy/issues/10712)
- POC: https://github.com/tharun208/kuma/commit/ab0020bc01ea1c440fc11c4e724ecb03b32519c2
- Original issue: https://github.com/kumahq/kuma/issues/680
- Original issue: https://github.com/kumahq/kuma/issues/724
