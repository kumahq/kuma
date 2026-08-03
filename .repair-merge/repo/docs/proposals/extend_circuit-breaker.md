# Support Envoy’s circuit breaker thresholds

## Context

Kuma already has a policy CircuitBreaker, but it relies only on Envoy outlier detector mechanism under the hood. The current proposal aims to provide a way to extend CircuitBreaker.  Extended CircuitBreaker policy should support Envoy’s circuit breaker thresholds alongside the outlier detector.

## Requirements

Extended CircuitBreaker policy should give a way to configure the following Envoy parameters:

* [1] [max_connections](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/cluster/v3/circuit_breaker.proto#envoy-v3-api-field-config-cluster-v3-circuitbreakers-thresholds-max-connections)  - the maximum number of connections that Envoy will make to the upstream cluster. If not specified, the default is 1024.

* [2] [max_pending_requests](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/cluster/v3/circuit_breaker.proto#envoy-v3-api-field-config-cluster-v3-circuitbreakers-thresholds-max-pending-requests) - the maximum number of pending requests that Envoy will allow to the upstream cluster. If not specified, the default is 1024.

* [3] [max_retries](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/cluster/v3/circuit_breaker.proto#envoy-v3-api-field-config-cluster-v3-circuitbreakers-thresholds-max-retries) - the maximum number of parallel retries that Envoy will allow to the upstream cluster. If not specified, the default is 3.

* [4] [max_requests](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/cluster/v3/circuit_breaker.proto#envoy-v3-api-field-config-cluster-v3-circuitbreakers-thresholds-max-requests) - the maximum number of parallel requests that Envoy will make to the upstream cluster. If not specified, the default is 1024.

## Solution

Add a new section `thresholds` alongside `detectors`:

```yaml
type: CircuitBreaker
mesh: default
name: circuit-breaker-example
sources:
- match:
    kuma.io/service: web
destinations:
- match:
    kuma.io/service: backend
conf:
  detectors:
    totalErrors: {}
    standardDeviation: {}
  thresholds:
    maxConnections: 2
    maxPendingRequests: 2
    maxRetries: 2
    maxRequests: 2
```
