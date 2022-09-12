# MeshAccessLog

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/4733

# Context and Problem Statement

[New policy matching MADR](./005-policy-matching.md) introduces a new approach how Kuma applies configuration to proxies.
Rolling out strategy implies creating a new set of policies that satisfy a new policy matching.
Current MADR aims to define a `MeshAccessLog`.

# Considered Options

* Create a MeshAccessLog policy:
  * Move backend to the `MeshAccessLogBackend` policy now and allow inlining in the `MeshAccessLog` policy itself.
  * Move backend to the `MeshAccessLogBackend` policy now.
  * Keep backends in the `Mesh` resource for now, move them into `MeshAccessLogBackend` policy in the future.
  * Embed backend in `MeshAccessLog` policy.

# Decision Outcome

Move backend to the `MeshAccessLogBackend` policy now and allow inlining in the logging policy itself.

## Decision Drivers

* Being able to assign RBAC permissions to an observability role
(if it was in the `Mesh` object then the "observability" role would have full `Mesh` permissions).
* The option to inline gives more flexibility and is easier to apply one policy than two.
* Moving the policy now minimises the work needed in the future.

## Overview

`MeshAccessLog` allows users to log incoming and outgoing traffic for services.
This feature is useful in the following scenarios:
* Auditing - services that contain important data might want to log access to its resources for auditing purposes.
* Debugging - an engineer might want to see traffic going in/out of a service to debug misbehaviour.
* Insights - logs are useful when trying to understand users and their behaviour.

## Specification

### Naming

#### Policy

During the video meeting we decided on `MeshAccessLog`.

##### Decision Drivers

* Matches the Envoy naming.

##### Alternatives considered

During the weekly meeting we started talking about the appropriate name for this policy, here are the names mentioned:

- `MeshLogging`
- `MeshLog`
- `MeshTrafficLog`
- `TrafficLogging`
- `Logging`
- `MeshAccessLog`

#### Backend

During the video meeting we decided on `GlobalAccessLogBackend` for the global scoped policy
and `MeshAccessLogBackend` for the `Mesh` scoped one.

### Matching

There are 3 parts to matching, each meaning a different part of the request flow:
- `spec.targetRef` - where the policy attaches to, `to` and `from` are relative to this element(s)
- `spec.to.targetRef` - outbound(s) of the `spec.targetRef`
- `spec.from.targetRef` - inbound(s) of the `spec.targetRef`

#### Top level

`spec.targetRef` can have the following kinds: `Mesh|MeshSubset|MeshService|MeshServiceSubset|MeshGatewayRoute|MeshHTTPRoute`
with the following caveats:

`MeshGatewayRoute` can only have `from` (there is no outbound listener).

`MeshHTTPRoute` here is an inbound route and can only have `from` (`to` always goes to the application).

Matching on `MeshGatewayRoute` and `MeshHTTPRoute` can be achieved by using a
[HeaderMatcher](https://github.com/envoyproxy/envoy/blob/23a9a686bb4237934cd575d8e62d3e0df98b59ee/api/envoy/config/accesslog/v3/accesslog.proto#L207)
with a value of `:path`.
When using these targets, defining `spec.to.targetRef` is **disallowed** because
there is no way of knowing if the outgoing request is connected to an incoming route.

#### From level

`spec.from.targetRef` can only have: `Mesh` for this iteration.

In the future, general use case of targeting `MeshSubset|MeshService|MeshServiceSubset` in TCP/HTTP will be implemented the following way:
- get SPIFFE info from certificate in a custom Lua Envoy filter (or use something provided if it exists, you can get the info by calling `handle:streamInfo():downstreamSslConnection():uriSanPeerCertificate()` - see example [here](https://github.com/allegro/envoy-control/blob/087a8d3bf9b923f2013dabcf3adef2bfe2d9533f/envoy-control-core/src/main/resources/lua/ingress_client_name_header.lua#L13))
- pass SPIFFE info into dynamic metadata by calling dynamicMetadata:set(filterName, key, value) (see [docs](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/lua_filter#set))
- use [metadata_filter](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/accesslog/v3/accesslog.proto#envoy-v3-api-field-config-accesslog-v3-accesslogfilter-metadata-filter) with data from the previous point

Matching on `MeshGatewayRoute` and `MeshHTTPRoute` does not make sense (there is no `route` that a request originates **from**).

#### To level

`spec.to.targetRef` can only have: `Mesh|MeshService` for now.

In the future,
when `spec.targetRef` is `Mesh|MeshSubset|MeshService|MeshServiceSubset` we can have `MeshHTTPRoute` as an Outbound route,
and it can only have `to` (`from` always goes from the application).

### Examples

`MeshAccessLog` can be an `inbound`, or `outbound` or `inbound/outbound` policy
meaning it can have `from`, `to`, or both `from` and `to` sections.

#### Only inbound

```yaml
type: MeshAccessLog
mesh: default
name: some-logging
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
type: MeshAccessLog
mesh: default
name: some-logging
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
type: MeshAccessLog
mesh: default
name: some-logging
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

#### Global scoped version

A global scoped version will be named `GlobalAccessLogBackend`, it has the same properties, but it does not have a `mesh` property.

### Backends

#### Move backend to the `MeshAccessLogBackend` policy now and allow inlining in the `MeshAccessLog` policy itself.

This is a hybrid of the approaches mentioned below.

```yaml
type: MeshAccessLogBackend # mesh scoped
name: file-backend
mesh: default
spec:
  name: file
  type: file
  conf:
    path: /tmp/access.log
---
type: MeshAccessLog
mesh: default
name: some-logging
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
        backends:
          - type: tcp
            conf:
              address: 127.0.0.1:5000
          - type: reference
            conf: 
              kind: MeshAccessLogBackend
              name: file-backend
```

##### Positive Consequences

* Being able to assign RBAC permissions to an observability role.
* The option to inline gives more flexibility and is easier to apply one policy than two.
* Moving the policy now minimises the work needed in the future.

##### Negative Consequences

* Multiple ways to do the same thing, we need to make sure to clearly document which approach is best for which use case.

#### Alternatives considered

##### Keep backends in the `Mesh` resource for now, move them into `MeshAccessLogBackend` policy in the future

For now new Policy `MeshAccessLog` will reuse the backend definitions from `Mesh`.

```yaml
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: default
spec:
  logging:
    # MeshAccessLog policies may leave the `backend` field undefined.
    # In that case the logs will be forwarded into the `defaultBackend` of that Mesh.
    defaultBackend: file
    # List of logging backends that can be referred to by name
    # from MeshAccessLog policies of that Mesh.
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
type: MeshAccessLog
mesh: default
name: some-logging
spec:
  targetRef:
    kind: MeshService
    name: web-backend
  from:
    - targetRef:
        kind: MeshService
        name: web-frontend
      default:
        backends:
          - name: logstash
            type: reference # this is the default and only supported value atm, people can ignore this
```

In the future we will introduce `MeshAccessLogBackend` policy that would hold that data.
Definitions there will have precedence over definitions in `Mesh`.
In the future the user will re-create (or maybe we could do this automatically) equivalent `MeshAccessLogBackend` policies.

###### Positive Consequences 

* Easier to implement.

###### Negative Consequences

* Possibly more work for the user in the future.
* It's impossible to create an observability role using RBAC.
You need to give access to the whole mesh to such a person.

##### Move backend to the `MeshAccessLogBackend` policy now

Moving backend to a new policy called `MeshAccessLogBackend` immediately.
That policy would be neither `inbound` nor `outbound` it would just store backend definitions.

```yaml
type: MeshAccessLogBackend
name: logstash-backend
mesh: default
spec:
  name: logstash
  type: tcp
  conf:
    address: 127.0.0.1:5000
---
type: MeshAccessLogBackend
name: file-backend
mesh: default
spec:
  name: file
  type: file
  conf:
    path: /tmp/access.log
---
type: MeshAccessLog
mesh: default
name: default-logging
spec:
  targetRef:
    kind: MeshService
    name: web-backend
  from:
    - targetRef:
        kind: MeshService
        name: web-frontend
      default:
        backends:
          - name: logstash
            type: reference
```

###### Positive Consequences

* Less duplication.

###### Negative Consequences

* New policies for users to understand.
* More fragmentation.
* Possibly more work to implement.

##### Embed backend in `MeshAccessLog` policy

Another option would be to embed `backend` field inside the `MeshAccessLog` policy,
but still have the possibility to reference a backend defined in a `Mesh`.

```yaml
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: default
spec:
  logging:
    defaultBackend: aBackendDefinedInMesh
    backends:
      - name: aBackendDefinedInMesh
        type: tcp
        conf:
          address: 127.0.0.1:5000
--- 
type: MeshAccessLog
mesh: default
name: some-logging
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
        backends:
          - name: logstash
            type: tcp
            conf:
              address: 127.0.0.1:5000
          - type: reference
            name: aBackendDefinedInMesh
```

###### Positive Consequences

* More flexibility.

###### Negative Consequences

* Multiple ways to do the same thing, might confuse users without clear documentation on which approach is best for which use case.
* 

### Other configuration options

#### Format

Format can be tied to a backend for some backends (like a backend that only supports JSON entries)
or not (a file will accept anything text-like), so for this reason we're leaving format inside a backend.

##### Improving JSON support

The current support of the format is just a plain string.
To be more user-friendly to the users of JSON format we could leverage Envoy's [json_format](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/substitution_format_string.proto.html?highlight=json_format).

We decided to have backend in a separate entity, so we don't have to keep backwards compatibility.
This structure **will no longer** be accepted in the new policy:

```yaml
backends:
  - name: logstash
    format: '{"start_time": "%START_TIME%"}' # implicit type=string
```

New definition will look like this:

```yaml
backends:
  - name: logstash
    format:
      type: string
      value: '{"start_time": "%START_TIME%"}'
```

And the new format type could be specified by a `type` parameter with a value of `json`:

```yaml
backends:
  - name: logstash
    format:
      type: json
      value:
        - key: "start_time"
          value: "%START_TIME%"
```

The corresponding OpenAPI v3 schema looks like this:

```yaml
components:
  schemas:
    Backend:
      type: object
      properties:
        name:
          type: string
        format:
          type: object
          properties:
            type:
              type: string
              enum:
              - string
              - json
            value:
              oneOf:
              - type: array
                items:
                  type: object
                  properties: 
                    key:
                      type: string
                    value:
                      type: string
              - type: string
```

###### Positive Consequences

* more user-friendly for JSON users

###### Negative Consequences

* needs time to implement (being aware of scope creep)

##### Considered option - Move `format` outside of `backend`

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

* a service owner might use format not accepted by the backed

#### Type

Stays the in the same place with the `backend`.

### gRPC support

One of the aspects of this new policy is better [support for gGRPC](https://github.com/kumahq/kuma/issues/3324). 
It seems that nothing [has changed](https://github.com/kumahq/kuma/issues/3324#issuecomment-1008991839) in Envoy and
there is no new features for gRPC apart from operators that already existed
(`GRPC_STATUS` - introduced [2 years ago](https://github.com/envoyproxy/envoy/blame/v1.17.0/docs/root/configuration/observability/access_log/usage.rst#L379),
`GRPC_STATUS_NUMBER` - introduced [4 months ago](https://github.com/envoyproxy/envoy/blame/main/docs/root/configuration/observability/access_log/usage.rst#L598)).
It might be worth explicitly pointing out these operators in our docs.

### Considered Options

#### Additional filtering

For now, we're not implementing filtering options,
but they are likely to be needed for this feature to be useful.
Additional filtering could be implemented as a separate property near the backend:

```yaml
...
      default:
        filtering:
          headers:
            ":status": 
              op: "eq"
              value": "500" # only log requests with status code 500
        backends:
...
```

So the whole policy would look like this:

```yaml
type: MeshAccessLog
mesh: default
name: some-logging
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
        filtering:
          headers:
            ":status":
              op: "eq"
              value": "500"
        backends:
          - name: logstash
            type: tcp
            conf:
              address: 127.0.0.1:5000
          - type: reference
            name: aBackendDefinedInMesh
```

## Examples

### Default with override (version for separate entity)

```yaml
type: MeshAccessLogBackend
name: logstash-backend
mesh: default
spec:
  name: logstash
  type: tcp
  conf:
    address: 127.0.0.1:5000
---
type: MeshAccessLogBackend
name: file-backend
mesh: default
spec:
  name: file
  type: file
  conf:
    path: /tmp/access.log
---
type: MeshAccessLog
mesh: default
name: default-logging
spec:
  targetRef:
    kind: Mesh
    name: default
  from:
    - targetRef:
        kind: Mesh
        name: default
      default:
        backends:
          - type: reference
            conf:
              kind: MeshAccessLogBackend
              name: logstash-backend
  to:
    - targetRef:
        kind: Mesh
        name: default
      default:
        backends:
          - type: reference
            conf:
              kind: MeshAccessLogBackend
              name: logstash-backend
---
type: MeshAccessLog
mesh: default
name: debugging-issue
spec:
  targetRef:
    kind: MeshService
    name: web-frontend
  to:
    - targetRef:
        kind: MeshService
        name: web-backend
      default:
        backends:
          - type: reference
            conf:
              kind: MeshAccessLogBackend
              name: file-backend
```

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
type: MeshAccessLog
mesh: default
name: some-logging
spec:
  targetRef:
    kind: MeshHTTPRoute
    name: backend-apple-path
  from:
    - targetRef:
        kind: Mesh
      default:
        backends:
          - type: tcp
            conf:
              address: 127.0.0.1:5000
```

Outcome:
- Every request that comes on `/apples` will be logged to "file".

### Zones

#### In mesh

```yaml
type: MeshAccessLogBackend
name: logstash-backend
mesh: default
spec:
  name: logstash-zone-a
  type: tcp
  conf:
    address: 127.0.0.1:5000
---
type: MeshAccessLogBackend
name: file-backend
mesh: default
spec:
  name: logstash-zone-b
  type: tcp
  conf:
    address: 127.0.0.2:5000
---
type: MeshAccessLog
mesh: default
name: log-zone-a
spec:
  targetRef:
    kind: MeshSubset
    tags:
      kuma.io/zone: zone-a
  from:
    - targetRef:
        kind: Mesh
        name: default
      default:
        backends:
          - type: reference
            conf:
              kind: MeshAccessLogBackend
              name: logstash-zone-a
---
type: MeshAccessLog
mesh: default
name: log-zone-b
spec:
  targetRef:
    kind: MeshSubset
    tags:
      kuma.io/zone: zone-b
  from:
    - targetRef:
        kind: Mesh
        name: default
      default:
        backends:
          - type: reference
            conf:
              kind: MeshAccessLogBackend
              name: logstash-zone-b
```
