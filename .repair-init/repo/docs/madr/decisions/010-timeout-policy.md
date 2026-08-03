# Timeout policy compliant with 2.0 model

* Status: accepted

Technical Story: [#4739](https://github.com/kumahq/kuma/issues/4739)

## Context and Problem Statement

[New policy matching MADR](https://github.com/kumahq/kuma/blob/master/docs/madr/decisions/005-policy-matching.md) introduces 
a new approach how Kuma applies configuration to proxies. Rolling out strategy implies creating a new set of policies that 
satisfy a new policy matching. Current MADR aims to define a timeout policy that will be compliant with 2.0 policies model.

## Considered Options

* Create `MeshUpstreamTimeout` policy
* Create two separate policies: `MeshUpstreamTimeout` and `MeshDownstreamTimeout`
* Create single `MeshTimeout` policy

## Decision Outcome

We will be crating single `MeshTimeout` policy that will cover both downstream and upstream timeouts.

## Solution

### How Timeout policy can be configured right now
This is the default timeout configuration in Kuma right now:

```yaml
 conf:
    connectTimeout: 5s # all protocols
    tcp: 
      idleTimeout: 1h 
    http:
      requestTimeout: 15s 
      idleTimeout: 1h
      streamIdleTimeout: 30m
      maxStreamDuration: 0s
```

Why idle timeout can be configured on both TCP and HTTP connections?
Is there a chance that client would like to configure different idle timeouts for database/Kafka and for HTTP services? - I can't come up on any valid case when this can be used.

### What timeouts can we configure in envoy
1. TCP specific
    1. connect timeout
    2. idle timeout
2. HTTP/GRPC connections specific
    1. connection timeouts
        1. idle timeout
        2. max connection duration
    2. stream timeouts
        1. request timeout
        2. request headers timeout
        3. stream idle timeout
        4. max stream duration
    3. route timeout
        1. idle timeout
        2. timeout
        3. per try timeout
        4. per try idle timeout
        5. max stream duration
           Detailed explanation of those timeouts can be found in [Envoy docs](https://www.envoyproxy.io/docs/envoy/latest/faq/configuration/timeouts).

#### Common timeouts
As we can see, `connection` and `idle timeout` can be configured for every connection type. There are two options here:

1. Keep separate idle timeouts for TCP and HTTP
2. Merge those two timeouts into single configuration

Ad 1. It gives clients more fine-tuning of timeouts, but it can cause chaos
Ad 2. Simpler config and you can still achieve the same result as previously by configuring it for `MeshSubset|MeshService|MeshServiceSubset`

#### Route timeouts
As for the route timeouts, we only care about `idle timeout` and `timeout` since `per try timeout` and `per try idle timeout` 
are retry related timeouts and should be configured in retry policy.

#### Request headers timeout
Last one specific timeout is `request headers timeout` but this timeout makes no sense on outbound connections. 
This only can be used to protect yourself from malicious inbound clients. We can add this timeout to `from` configuration

#### Scaled timeouts
Moreover, there is a mechanism of scaled timeout in envoy, which I am not sure if we want to expose to our users, at least not at first.

### Specification

#### Top level
Top-level targetRef can have all available kinds:
```yaml
targetRef:
  kind: Mesh|MeshSubset|MeshService|MeshServiceSubset|MeshGatewayRoute|MeshHTTPRoute
  name: ...
```
In this policy same limitations applies as in `MeshAccessLog` [policy](https://github.com/kumahq/kuma/blob/master/docs/madr/decisions/008-mesh-logging.md#top-level):
- `MeshGatewayRoute` can only have from (there is no outbound listener).
- `MeshHTTPRoute` here is an inbound route and can only have from (to always goes to the application).


#### From level

```yaml
from:
  - targetRef:
      kind: Mesh
      name: ...
```

Since timeouts are mostly configured on clusters and listeners there and we have single inbound in most cases we can only
configure `Mesh` kind in from section.

#### To level

```yaml
to:
  - targetRef:
      kind: Mesh|MeshService|MeshHTTPRoute
      name: ...
```

For now this policy will be implemented only for `Mesh` and `MeshService` in the future support for `MeshHTTPRoute` can be added.
For `MeshHTTPRoute` there are few limitations. According to docs for route we can only configure: `idleTimeout`,
`requestTimeout` and `maxStreamDuration`. In the future when adding support for `MeshHTTPRoute` we can validate if only allowed timeouts are used.

#### Minimal configuration
Taking into account previous considerations, the minimal configuration for `MeshUpstreamTimeout` policy should look like this:
```yaml
default:
  connectTimeout: ... # all protocols
  idleTimeout: ... # all protocols
  http:
    requestTimeout: ... 
    streamIdleTimeout: ...
    maxStreamDuration: ...
    maxConnectionDuration: ...
```


### Examples
#### Default configuration
```yaml
type: MeshUpstreamTimeout  
mesh: default  
name: default-timeouts  
spec:  
  targetRef:  
    kind: Mesh  
    name: default  
  from:
    - targetRef:
        kind: Mesh
        name: default
      default:
        connectTimeout: 10s
        idleTimeout: 2h
        http:
          requestTimeout: 0s
          streamIdleTimeout: 1h
          maxStreamDuration: 0s
          maxConnectionDuration: 0s
  to:  
    - targetRef:  
        kind: Mesh  
        name: default  
      default:  
        connectTimeout: 5s  
        idleTimeout: 1h  
        http:  
          requestTimeout: 15s  
          streamIdleTimeout: 30m  
          maxStreamDuration: 0s
          maxConnectionDuration: 0s
```

#### Configuration with overridden TCP specific timeouts for database

```yaml
type: MeshTimeout  
mesh: default  
name: default-timeouts  
spec:  
  targetRef:  
    kind: Mesh  
    name: default  
  to:  
    - targetRef:  
        kind: Mesh  
        name: default  
      default:  
        connectTimeout: 5s  
        idleTimeout: 1h  
        http:  
          requestTimeout: 15s  
          streamIdleTimeout: 30m  
          maxStreamDuration: 0s
          maxConnectionDuration: 0s
---  
type: MeshTimeout  
mesh: default  
name: database-tcp-timeouts  
spec:  
  targetRef:  
    kind: Mesh  
    name: default  
  to:  
    - targetRef:  
        kind: MeshService  
        name: postgres  
      default:  
        connectTimeout: 1s  
        idleTimeout: 10m
```

#### Configuration of MeshHTTPRoute for single service

```yaml
type: MeshTimeout  
mesh: default  
name: default-timeouts  
spec:  
  targetRef:  
    kind: MeshService  
    name: backend  
  to:  
    - targetRef:  
        kind: MeshHTTPRoute  
        name: payment-v1-route  
      default:  
        idleTimeout: 1h  
        http:  
          requestTimeout: 2s    
          maxStreamDuration: 0s
```

#### Configuration of MeshGatewayRoute on cross mesh communication

```yaml
type: MeshTimeout  
mesh: expose  
name: default-gateway-timeouts
spec:
  targetRef:
    kind: MeshGatewayRoute
    name: default-gateway-route
  to:
    - targetRef:
        kind: MeshService
        name: backend
      default:
        idleTimeout: 30m
        http:
          requestTimeout: 2s
```