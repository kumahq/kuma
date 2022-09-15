# Traffic tracing policy

- Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/4732

## Context and Problem Statement

We want to create a [new policy matching compliant](https://github.com/kumahq/kuma/blob/22c157d4adac7f518b1b49939c7e9ea4d2a1876c/docs/madr/decisions/005-policy-matching.md)
resource for managing traffic tracing.

## Decision Drivers

- Replace existing policies with new policy matching compliant resources
- [Path/Method filtering on traces](https://github.com/kumahq/kuma/issues/3335)
- [Custom tags](https://github.com/kumahq/kuma/issues/3275)
- Not blocking [OpenTelemetry tracing](https://github.com/kumahq/kuma/issues/3690)

## Considered Options

The new resource will be called `Tracing` in the rest of this document.

Tracing in Kuma is implemented with Envoy's [tracing
support](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/observability/tracing).
At the moment, Kuma supports Zipkin and Datadog.

### Naming

Some potential names:

- `Tracing`
- `Traces`
- `MeshTrafficTrace`
- `MeshTrace`

### Backends

Currently, tracing support requires defining _backends_ in the `Mesh` resource.
A backend defines where traces are sent, the _provider_. For `TrafficTrace`, it also contains the sampling rates settings.
The `TrafficTrace` resource references these backends and contains no configuration of its own.

The proposed `Tracing` policy follows the decisions about backends as laid out in
[the MADR for `MeshAccessLog`](docs/madr/decisions/009-tracing-policy.md#backends). This
means creating a new `TracingBackend` resource and allowing inline backend
definitions.

However, we propose moving configuring sampling and tag configuration
from the backend (the `TrafficTrace` status quo) to the `Tracing` resource.
Backends can contain the same settings but they can be set or overriden in
specific `Tracing` resources.

We continue to support Zipkin and Datadog:

```
apiVersion: kuma.io/v1alpha1
kind: TracingBackend
metadata:
 name: jaeger
spec:
  zipkin:
    url: http://jaeger-collector.mesh-observability:9411/api/v2/spans
```

```
apiVersion: kuma.io/v1alpha1
kind: TracingBackend
metadata:
 name: datadog
spec:
  datadog:
    address: trace-svc.datadog.svc.cluster.local
    port: 8126
```

### `targetRef`

This is a new policy so it's based on `targetRef`. Envoy tracing is configured
on the HTTP connection manager so `Tracing` has a single `spec.targetRef` field
that selects what it applies to. It does not use `to` or `from` fields.

All logging configuration happens under `spec.default` so that users are able to
override settings with more specific `targetRef`s.

Resources supported by `spec.targetRef` are `Mesh`, `MeshSubset`, `MeshService`,
`MeshServiceSubset`, `MeshGatewayRoute` and eventually the future
evolution of `TrafficRoute`, which we'll call `MeshTrafficRoute` in this
document.

### Policy-specific configuration

This MADR proposes additional options be configurable on `Tracing` resources
directly.

#### Sampling

At the moment, tracing backends only support Envoy's
[`overall_sampling`](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/network/http_connection_manager/v3/http_connection_manager.proto#envoy-v3-api-field-extensions-filters-network-http-connection-manager-v3-httpconnectionmanager-tracing-overall-sampling)
via the `sampling` field.

This MADR proposes instead the following:

```yaml
spec:
 default:
  sampling:
   overall: <percentage>
   client: <percentage>
   random: <percentage>
```

with the justification that it's not clear to users
that when they set the simple `sampling` field we have now,
whether traces started via `x-client-trace-id` are limited.
This also enables disabling `x-client-trace-id`.

#### Tags

Envoy can set tags in traces using values from:

- Literal value
- Request header
- Environment variable
- Metadata

We propose allowing users to configure custom tags using only literal values or
request headers. Using Kuma leaves environment variables and Envoy metadata
opaque for users so the use cases for configuring them (at least directly)
in Kuma are limited.

Tags can be configured as follows:

```yaml
spec:
 default:
  tags:
   - name:
     # exactly one of the following keys must be set
     literal: ...
     header:
      name:
      default:
```

#### Method/path specific

Users can set the `spec.targetRef` to be a `MeshGatewayRoute` or
`MeshTrafficRoute`.

Both paths and methods are matchable in `MeshGatewayRoute` and we can ensure
that `MeshTrafficRoute` also supports this matching.

## Examples

All examples assume Kubernetes.

### Simple

This configures all `Dataplane` inbounds with `kuma.io/service: backend` to
trace a maximum of 80% of traffic and adds the custom tag `team` with a value of
`core` to spans. The backend is specified in a `TracingBackend` object.

```yaml
apiVersion: kuma.io/v1alpha1
kind: Tracing
metadata:
 name: all
 labels:
  kuma.io/mesh: default
spec:
 targetRef:
  kind: MeshService
  name: backend
 default:
  backend:
   ref:
    kind: TracingBackend
    name: jaeger
  sampling:
   overall: 80
  tags:
   - name: team
     literal: core
```

### Inline backend configuration

This specifies the backend directly:

```yaml
apiVersion: kuma.io/v1alpha1
kind: Tracing
metadata:
 name: all
 labels:
  kuma.io/mesh: default
spec:
 targetRef:
  kind: MeshService
  name: backend
 default:
  backend:
   zipkin:
    url: http://jaeger-collector.mesh-observability:9411/api/v2/spans
  sampling:
   overall: 80
```

### Route-specific

This configures any listeners matched by the `MeshGatewayRoute` `prod`
to trace a maximum of 80% of traffic and adds the custom tag `env`
with a value from `x-env` and a default value of `prod`.

```yaml
apiVersion: kuma.io/v1alpha1
kind: Tracing
metadata:
 name: all
 labels:
  kuma.io/mesh: default
spec:
 targetRef:
  kind: MeshGatewayRoute
  name: prod
 default:
  backend:
   ref:
    kind: TracingBackend
    name: jaeger
  sampling:
   overall: 80
  tags:
   - name: env
     header:
      name: x-env
      default: prod
```
