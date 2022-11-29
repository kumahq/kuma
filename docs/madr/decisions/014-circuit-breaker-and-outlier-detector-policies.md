# Circuit Breaker and Outlier Detector policies compliant with 2.0 model

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
capabilities behind the scene. Consequences of this approach may introduce the situation when during reading envoy's
stats for debugging purposes, you may look for stats prefixed with `cluster.<name>.circuit_breakers.` with occurred
fault in mind, and according to the behaviour of our policy, if the data plane was ejected because of fault it's taking
into account, stats which contain interesting us data will be prefixed with `cluster.<name>.outlier_detection.` prefix.

## Considered Options

* Create two separate policies: `MeshCircuitBreaker` and `MeshOutlierDetector`
* Create single `MeshCircuitBreaker` policy which would map 1:1 current policy
* Create single `MeshFaultDetector` policy which would also map 1:1 to the current `CircuitBreaker`, but with non
  misleading name

## Decision Outcome

Chosen option: create two separate -`MeshCircuitBreaker` and `MeshOutlierDetector` policies.

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

Top-level targetRef can have all available kinds (with restriction of `MeshHTTPRoute`
not being implemented as of time of writing this MADR):

```yaml
targetRef:
  kind: Mesh|MeshSubset|MeshService|MeshServiceSubset|MeshGatewayRoute|MeshHTTPRoute
  name: ...
```

#### From level

`MeshCircuitBreaker` and `MeshOutlierDetector` are outbound only policies, so only `to` should be configured.

#### To level

```yaml
to:
- targetRef:
    kind: Mesh|MeshSubset|MeshService|MeshServiceSubset|MeshGatewayRoute|MeshHTTPRoute
    name: ...
```

#### Minimal configuration

##### `MeshCircuitBreaker`

As all the circuit breaking related properties have default values, there is no need to specify any of these to apply
the policy

```yaml
default: { }
```

##### `MeshOutlierDetector`

```yaml
default:
  detectors:
    totalErrors:
      consecutive: ...
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
      maxConnections: 1024
      maxPendingRequests: 1024
      maxRetries: 3
      maxRequests: 1024
```

##### `MeshOutlierDetector`

There won't be a default `MeshOutlierDetestor`  (following the example of our existing `CircuitBreaker` policy)

#### Extensive configuration

##### `MeshCircuitBreaker`

As the default policy is extensive enough, there is no need to duplicate the examples.

##### `MeshOutlierDetector`

```yaml
type: MeshOutlierDetector
mesh: default
name: extensive-outlier-detector
spec:
  targetRef:
    kind: Mesh
    name: default
  to:
  - targetRef:
      kind: Mesh
      name: default
    default:
      interval: 5s
      baseEjectionTime: 30s
      maxEjectionPercent: 20
      splitExternalAndLocalErrors: true
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
