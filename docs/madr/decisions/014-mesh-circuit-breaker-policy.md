# Circuit Breaker policy compliant with 2.0 model

* Status: accepted

Technical Story: [#4736](https://github.com/kumahq/kuma/issues/4736)

## Context and Problem Statement

[New policy matching MADR](https://github.com/kumahq/kuma/blob/master/docs/madr/decisions/005-policy-matching.md)
introduces a new approach how Kuma applies configuration to proxies. Rolling out strategy implies creating a new set of
policies that satisfy a new policy matching. Current MADR aims to define a timeout policy that will be compliant with
2.0 policies model.

Current `CircuitBreaker` policy combines
envoy's [circuit breaking](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/circuit_breaking)
and [outlier detection](https://gstwww.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/outlier)
capabilities behind the scene, but there is no clear differentiation between two in the configuration specification. It
may be misleading that we have a policy named the same as some envoy functionality, but it actually combines more
functionalities. Consequences of this approach may introduce the situation when person more familiar with envoy during
debugging process, may look only at stats prefixed with `cluster.<name>.circuit_breakers.`, but according to the
behaviour of our policy, if the data plane was ejected because of one of the faults defined in `detectors` section,
stats which contain interesting us data will be prefixed with `cluster.<name>.outlier_detection.`.

## Considered Options

* Create two separate policies: `MeshCircuitBreaker` and `MeshOutlierDetector`
* Create single `MeshCircuitBreaker` policy which would map 1:1 current policy
* Create single `MeshCircuitBreaker` with configuration divided
  to `connectionPool` ([envoy's circuit breaker](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/circuit_breaking))
  and `outlierDetection` ([envoy's outlier detection](https://gstwww.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/outlier))

## Decision Outcome

Chosen option: Create single `MeshCircuitBreaker` with configuration divided into two sections.

## Positive Consequences

* Having two separate policies, each with explicit configurations makes it easier to look for appropriate Envoy stats
  and XDP configuration values.
* By combining two abstractions, which exist in Envoy, and calling them both combined with a name of one of these is
  confusing, especially when you are trying to debug some issue without deep understanding of internals of our service
  mesh.
* Default `MeshCircuitBreaker` policy is sufficient to replicate behaviour of current version of this policy - there
  won't be a need to include a default `MeshOutlierDetector` policy

## Negative Consequences

* `MeshCircuitBreaker` will be a policy with very small amount of configuration available, which may introduce some
  clutter (at some point there may be so many policies, that it will be much more difficult to understand all of them)

## Solution

By dividing the configuration section into two separate, logical sections: `connectionPool` and `outlierDetection` it's
easier to differentiate underlying envoy features without the need to split policy into two separate ones.

Additional changes to the original `CircuitBreaker` policy consists of:

- flattening `detectors.*` options with only single (`.consecutive`) property
- changing word `Errors` to `Failures`
- adding `origin` suffix to the local failures detector

It means `totalErrors.consecutive`, `gatewayErrors.consecutive` and `localErrors.consecutive`
becomes: `consecutiveTotalFailures`, `consecutiveGatewayFailures` and `consecutiveLocalOriginFailures` respectively.

### Current configuration

Below is the sample `MeshCircuitBreaker` configuration

```yaml
conf:
  interval: 1s
  baseEjectionTime: 30s
  maxEjectionPercent: 20
  splitExternalAndLocalErrors: false
  thresholds:
    maxConnections: 2
    maxPendingRequests: 2
    maxRequests: 2
    maxRetries: 2
  detectors:
    totalErrors:
      consecutive: 20
    gatewayErrors:
      consecutive: 10
    localErrors:
      consecutive: 7
    standardDeviation:
      requestVolume: 10
      minimumHosts: 5
      factor: 1.9
    failure:
      requestVolume: 10
      minimumHosts: 5
      threshold: 85

```

### Specification

#### Top level

As a part of this MADR root `targetRef`'s kind can be configured as one
of `Mesh|MeshSubset|MeshService|MeshServiceSubset`. When we finally introduce `MeshHTTPRoute` kind and if there will be
a reasonable use case, in the next iteration `MeshGatewayRoute` and `MeshHTTPRoute` could be added as a viable kinds.

```yaml
targetRef:
  kind: Mesh|MeshSubset|MeshService|MeshServiceSubset
  name: ...
```

#### From level

`MeshCircuitBreaker` is outbound only policy, so only `to` should be configured.

#### To level

```yaml
to:
- targetRef:
    kind: Mesh|MeshSubset|MeshService|MeshServiceSubset
    name: ...
```

#### Minimal configuration

As all the connection pool related properties have default values, there is no need to specify any of these to apply the
policy

```yaml
default:
  connectionPool: { }
```

### Examples

#### Default configuration

##### `MeshCircuitBreaker`

```yaml
type: MeshCircuitBreaker
mesh: default
name: default-circuit-breaker
spec:
  targetRef:
    kind: Mesh
    name: default
  to:
  - targetRef:
      kind: Mesh
      name: default
    default:
      connectionPool:
        maxConnections: 1024
        maxPendingRequests: 1024
        maxRetries: 3
        maxRequests: 1024
```

#### Extensive configuration

```yaml
type: MeshCircuitBreaker
mesh: default
name: extensive-circuit-breaker
spec:
  targetRef:
    kind: Mesh
    name: default
  to:
  - targetRef:
      kind: Mesh
      name: default
    default:
      connectionPool:
        maxConnections: 1024
        maxPendingRequests: 1024
        maxRetries: 3
        maxRequests: 1024
      outlierDetection:
        interval: 5s
        baseEjectionTime: 30s
        maxEjectionPercent: 20
        splitExternalAndLocalErrors: true
        detectors:
          consecutiveTotalFailures: 10
          consecutiveGatewayFailures: 10
          consecutiveLocalOriginFailures: 10
          successRate:
            minimumHosts: 5
            requestVolume: 10
            standardDeviationFactor: 1.9
          failurePercentage:
            requestVolume: 10
            minimumHosts: 5
            threshold: 85
```
