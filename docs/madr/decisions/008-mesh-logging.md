# MeshLogging

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/4733

# Context and Problem Statement

[New policy matching MADR](./005-policy-matching.md) introduces a new approach how Kuma applies configuration to proxies.
Rolling out strategy implies creating a new set of policies that satisfy a new policy matching.
Current MADR aims to define a `MeshLogging`.

# Considered Options

* Create a MeshLogging policy
  * Keep backends in the `Mesh` resource for now, move them into `MeshLoggingBackend` policy in the future
  * Move backend to the `MeshLoggingBackend` policy now
  * Embed backend in `MeshLogging` policy

# Decision Outcome

To be determined.

## Overview

MeshLogging allows users to log incoming and outgoing traffic for services.
This feature is useful in the following scenarios:
* auditing - services that contain important data might want to log access to its resources for auditing purposes
* debugging - an engineer might want to see traffic going in/out of a service to debug misbehaviour
* insights - logs are useful when trying to understand users and their behaviour

## Specification

### Naming

During the weekly meeting we started talking about the appropriate name for this policy, here are the names mentioned:
- `MeshAccessLog`
- `MeshLog`
- `MeshTrafficLog`

Please vote on your preferred name in the comments. Feel free to edit this document and add your suggestions.

### Matching

There are 3 parts to matching, each meaning a different part of the request flow:
- `spec.targetRef` - where the policy attaches to, `to` and `from` are relative to this element(s)
- `spec.to.targetRef` - outbound(s) of the `spec.targetRef`
- `spec.from.targetRef` - inbound(s) of the `spec.targetRef`

#### Top level

`spec.targetRef` can have the following kinds: `Mesh|MeshSubset|MeshService|MeshServiceSubset|MeshGatewayRoute|MeshHTTPRoute`

Matching on `MeshGatewayRoute` and `MeshHTTPRoute` can be achieved by using a
[HeaderMatcher](https://github.com/envoyproxy/envoy/blob/23a9a686bb4237934cd575d8e62d3e0df98b59ee/api/envoy/config/accesslog/v3/accesslog.proto#L207)
with a value of `:path`.

#### From level

`spec.from.targetRef` can only have: `Mesh|MeshSubset|MeshService|MeshServiceSubset`.
Matching on `MeshGatewayRoute` and `MeshHTTPRoute` does not make sense (there is no `route` that a request originates **from**).

#### To level

`spec.to.targetRef` can have: `Mesh|MeshSubset|MeshService|MeshServiceSubset|MeshHTTPRoute`.
Matching on `MeshGatewayRoute` and `MeshHTTPRoute` will be achieved the same way as on the top level `targetRef`
but it will be attached on the `outbound` instead of the `inbound`.

### Examples

MeshLogging can be an `inbound`, or `outbound` or `inbound/outbound` policy
meaning it can have `from`, `to`, or both `from` and `to` sections.

#### Only inbound

```yaml
type: MeshLogging
mesh: default
spec:
  targetRef:
    kind: MeshService
    name: web-backend
  from:
    - targetRef:
        kind: MeshService
        name: web-frontend
```

This will log all the inbound traffic from `web-frontend` in `web-backend`.

#### Only outbound

```yaml
type: MeshLogging
mesh: default
spec:
  targetRef:
    kind: MeshService
    name: web-backend
  to:
    - targetRef:
        kind: MeshService
        name: web-queue
```

This will log all the outbound traffic to `web-queue` in `web-backend`.

#### Both inbound and outbound

```yaml
type: MeshLogging
mesh: default
spec:
  targetRef:
    kind: MeshService
    name: web-backend
  from:
    - targetRef:
        kind: MeshService
        name: web-frontend
  to:
    - targetRef:
        kind: MeshService
        name: web-queue
```

This will log all the inbound traffic from `web-frontend` and outbound traffic to `web-queue` in `web-backend`.

### Backends

#### Keep backends in the `MeshLogging` resource for now, move them into `MeshLoggingBackend` policy in the future

For now new Policy `MeshLogging` will reuse the backend definitions from `Mesh`.

```yaml
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: default
spec:
  logging:
    # MeshLogging policies may leave the `backend` field undefined.
    # In that case the logs will be forwarded into the `defaultBackend` of that Mesh.
    defaultBackend: file
    # List of logging backends that can be referred to by name
    # from MeshLogging policies of that Mesh.
    backends:
      - name: logstash
        type: tcp
        conf:
          address: 127.0.0.1:5000
      - name: file
        type: file
        conf:
          path: /tmp/access.log
---
type: MeshLogging
mesh: default
spec:
  targetRef:
    kind: MeshService
    name: web-backend
  from:
    - targetRef:
        kind: MeshService
        name: web-frontend
      default:
        backend: logstash # Forward the logs into the logging backend named `logstash`.
```

In the future we will introduce `MeshLoggingBackend` policy that would hold that data.
Definitions there will have precedence over definitions in `Mesh`.
In the future the user will re-create (or maybe we could do this automatically) equivalent `MeshLoggingBackend` policies.

##### Positive Consequences 

* easier to implement

##### Negative Consequences

* breaks the barrier between new / old policies by reusing old policies resources
* possibly more work for the user in the future

#### Move backend to the `MeshLoggingBackend` policy now

Moving backend to a new policy called `MeshLoggingBackend` immediately.
That policy would be neither `inbound` nor `outbound` it would just store backend definitions.
It would support `targetRef` of `Mesh` but could probably support more in the future.
This allows us to have consistent naming (`enabledBackend`) across all policies and
would make it easier to have per zone backends.

```yaml
type: MeshLoggingBackend
mesh: default
spec:
  targetRef:
    kind: Mesh
    name: default
  enabledBackend: file # instead of defaultBackend
  # List of logging backends that can be referred to by name
  # from MeshLogging policies of that Mesh.
  default:
    backends:
      - name: logstash
        type: tcp
        conf:
          address: 127.0.0.1:5000
      - name: file
        type: file
        conf:
          path: /tmp/access.log
---
type: MeshLogging
mesh: default
spec:
  targetRef:
    kind: MeshService
    name: web-backend
  from:
    - targetRef:
        kind: MeshService
        name: web-frontend
      default:
        backend: logstash # Forward the logs into the logging backend named `logstash`.
```

##### Positive Consequences

* cleaner solution, there is separation between old and new policies
* consistency in regards to `enabledBackend`
* easier to have per zone backends

##### Negative Consequences

* possibly more work to implement

#### Embed backend in `MeshLogging` policy

Another option would be to embed `backend` field inside the `MeshLogging` policy.

```yaml
type: MeshLogging
mesh: default
spec:
  targetRef:
    kind: MeshService
    name: web-backend
    tags:
      kuma.io/zone: us-east
  from:
    - targetRef:
        kind: MeshService
        name: web-frontend
      default:
        backend:
          name: logstash # Forward the logs into the logging backend named `logstash`.
          type: tcp
          conf:
            address: 127.0.0.1:5000
```

##### Positive Consequences

* consistency in regards to `enabledBackend`
* easier to have per zone backends

##### Negative Consequences

* less dry

### Other configuration options

#### Format

Format is currently defined on a `backend`, but in reality it's not really tied to a backend 
(multiple backends can have the same format).

##### Leave `format` in the `backend`

###### Positive Consequences

* users are familiar with this

###### Negative Consequences

* less dry

##### Move `format` outside of `backend`

```yaml

default:
  formats:
    - name: "long-format"
      value: '{"start_time": "%START_TIME%", "source": "%KUMA_SOURCE_SERVICE%", "destination": "%KUMA_DESTINATION_SERVICE%", "source_address": "%KUMA_SOURCE_ADDRESS_WITHOUT_PORT%", "destination_address": "%UPSTREAM_HOST%", "duration_millis": "%DURATION%", "bytes_received": "%BYTES_RECEIVED%", "bytes_sent": "%BYTES_SENT%"}'
    - name: "short-format"
      value: '{"start_time": "%START_TIME%", "source": "%KUMA_SOURCE_SERVICE%", "destination": "%KUMA_DESTINATION_SERVICE%"}'
  backends:
    - name: logstash
      format: "short-format"
    - name: debug-file
      format: "long-format"
```

###### Positive Consequences

* more dry

###### Negative Consequences

* new for the user

#### Type

Stays the in the same place with the `backend`.

### gRPC support

One of the aspects of this new policy is better [support for gGRPC](https://github.com/kumahq/kuma/issues/3324). 
It seems that nothing [has changed](https://github.com/kumahq/kuma/issues/3324#issuecomment-1008991839) in Envoy and
there is no new features for gRPC apart from operators that already existed
(`GRPC_STATUS` - introduced [2 years ago](https://github.com/envoyproxy/envoy/blame/v1.17.0/docs/root/configuration/observability/access_log/usage.rst#L379),
`GRPC_STATUS_NUMBER` - introduced [4 months ago](https://github.com/envoyproxy/envoy/blame/main/docs/root/configuration/observability/access_log/usage.rst#L598)).
It might be worth explicitly pointing out these operators in our docs.

## Examples

More full-fledged one using all the options will be described once the desired implementation is chosen.

### Apples and oranges 

```yaml
type: MeshHTTPRoute
name: backend-apple-path
spec:
  http:
    rules:
      - matches:
          - path:
              match: PREFIX
              value: /apples
```
```yaml
type: MeshHTTPRoute
name: web-orange-path
spec:
  http:
    rules:
      - matches:
          - path:
              match: PREFIX
              value: /oranges
```
```yaml
type: TrafficLog
mesh: mesh-1
name: tl-1
spec:
  targetRef:
    kind: MeshHTTPRoute
    name: backend-apple-path
  to:
    - targetRef:
        kind: MeshHTTPRoute
        name: web-orange-path
      default:
        backends:
          - name: logstash
            format: ""
          - name: debug-file
  from:
    - targetRef:
        kind: Mesh
      default:
        backends:
          - name: file
```

Outcome:
- Every request that comes on `/apples` will be logged to "file".
- Every request that goes to `/oranges` will be logged to "logstash".
