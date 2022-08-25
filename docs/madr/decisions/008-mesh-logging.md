# MeshLogging

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/4733

## Context and Problem Statement

[New policy matching MADR](./005-policy-matching.md) introduces a new approach how Kuma applies configuration to proxies.
Rolling out strategy implies creating a new set of policies that satisfy a new policy matching.
Current MADR aims to define a MeshLogging.

## Considered Options

* Create a MeshLogging policy
  * Keep backends in the `Mesh` resource for now, move them into `MeshLoggingBackend` policy in the future
  * Move backend to the `MeshLoggingBackend` policy now
  * Embed backend in `MeshLogging` policy

## Decision Outcome

Chosen option: "{option 1}", because {justification. e.g., only option, which meets k.o. criterion decision driver | which resolves force {force} | … | comes out best (see below)}.

### Overview

MeshLogging allows users to log incoming and outgoing traffic for services.
This feature is useful in the following scenarios:
* auditing - services that contain important data might want to log access to its resources for auditing purposes
* debugging - an engineer might want to see traffic going in/out of a service to debug misbehaviour
* insights - logs are useful when trying to understand users and their behaviour

### Specification

#### Matching

Top-level targetRef can have the following kinds:

```yaml
targetRef:
  kind: Mesh|MeshSubset|MeshService|MeshServiceSubset
  name: ...
```

TODO: Figure out how much work would it be to add `MeshGatewayRoute` and `MeshHTTPRoute`.

MeshLogging can be an `inbound`, or `outbound` or `inbound/outbound` policy
meaning it can have `from`, `to`, or both `from` and `to` sections.

#### Examples

##### Only inbound

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

This will log all the inbound traffic from `web-backend` in `web-backend`.

##### Only outbound

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

##### Both inbound and outbound

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

#### Backends

##### Keep backends in the `MeshLogging` resource for now, move them into `MeshLoggingBackend` policy in the future

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
  conf:
    # Forward the logs into the logging backend named `logstash`.
    backend: logstash
```

In the future we will introduce `MeshLoggingBackend` policy that would hold that data.
Definitions there will have precedence over definitions in `Mesh`.
In the future the user will re-create (or maybe we could do this automatically) equivalent `MeshLoggingBackend` policies.

###### Positive Consequences 

* easier to implement

###### Negative Consequences

* breaks the barrier between new / old policies by reusing old policies resources
* possibly more work for the user in the future

##### Move backend to the `MeshLoggingBackend` policy now

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
  enabledBackend: file
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
  conf:
    # Forward the logs into the logging backend named `logstash`.
    backend: logstash
```

###### Positive Consequences

* cleaner solution, there is separation between old and new policies
* consistency in regards to `enabledBackend`
* easier to have per zone backends

###### Negative Consequences

* possibly more work to implement

##### Embed backend in `MeshLogging` policy

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
  conf:
    # Forward the logs into the logging backend named `logstash`.
    backend:
      name: logstash
      type: tcp
      conf:
        address: 127.0.0.1:5000
```

###### Positive Consequences

* consistency in regards to `enabledBackend`
* easier to have per zone backends

###### Negative Consequences

* less dry

#### Other configuration options

All other configuration options (`format`, `type`) stay the same and
are put in the same places in the configuration tree depending on which implementation is used.

A full-fledged example using all the options will be described once the desired implementation is chosen.

#### gRPC support

One of the aspects of this new policy is better [support for gGRPC](https://github.com/kumahq/kuma/issues/3324). 
It seems that nothing [has changed](https://github.com/kumahq/kuma/issues/3324#issuecomment-1008991839) in Envoy and
there is no new features for gRPC apart from operators that already existed
(`GRPC_STATUS` - introduced [2 years ago](https://github.com/envoyproxy/envoy/blame/v1.17.0/docs/root/configuration/observability/access_log/usage.rst#L379),
`GRPC_STATUS_NUMBER` - introduced [4 months ago](https://github.com/envoyproxy/envoy/blame/main/docs/root/configuration/observability/access_log/usage.rst#L598)).
It might be worth explicitly pointing out these operators in our docs.