# Traffic tracing policy

- Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/4732

## Context and Problem Statement

We want to create a [new policy matching compliant](https://github.com/kumahq/kuma/blob/22c157d4adac7f518b1b49939c7e9ea4d2a1876c/docs/madr/decisions/005-policy-matching.md)
resource for managing traffic tracing.

Tracing in Kuma is implemented with Envoy's [tracing
support](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/observability/tracing).
At the moment, Kuma supports Zipkin and Datadog.

### Sampling in Envoy

Envoy implements head-based sampling, meaning the originating service (called also `head` or `first`)
is responsible for deciding if the request is traced or not.
You can read about this in the [docs](https://github.com/envoyproxy/envoy/blame/a36b06a92634463f76689b5bb04338105527d5c7/docs/root/configuration/http/http_conn_man/headers.rst#L496):

> When the Sampled flag is either not specified or set to 1,
> the span will be reported to the tracing system.
> Once Sampled is set to 0 or 1, the same value should be consistently sent downstream.

This means that the `sampling` value in a service that is not the first one to receive traffic
does not change the behaviour.
This makes it somewhat unintuitive when targeting services lower down the chain.
The easiest solution is just to set tracing on a `Mesh` level,
but this limits the users' abilities.

## Decision Drivers

- Replace existing policies with new policy matching compliant resources
- [Path/Method filtering on traces](https://github.com/kumahq/kuma/issues/3335)
- [Custom tags](https://github.com/kumahq/kuma/issues/3275)
- Not blocking [OpenTelemetry tracing](https://github.com/kumahq/kuma/issues/3690)

## Considered Options

### Naming

Some potential names:

- ~`Tracing`~
- ~`Traces`~
- ~`MeshTrafficTrace`~
- **Chosen**: `MeshTrace`

### Backends

Currently, tracing support requires defining _backends_ in the `Mesh` resource.
A backend defines where traces are sent, the _provider_. For `TrafficTrace`, it also contains the sampling rates settings.
The `TrafficTrace` resource references these backends and contains no configuration of its own.

The proposed `MeshTrace` policy follows the decisions about backends as laid out in
[the MADR for `MeshAccessLog`](docs/madr/decisions/009-tracing-policy.md#backends). This
means creating a new `MeshTraceBackend` resource and allowing inline backend
definitions.

However, we propose moving configuring sampling and tag configuration
from the backend (the `TrafficTrace` status quo) to the `MeshTrace` resource.

In the future, we may support the same settings as defaults in backends
and overrides from specific `MeshTrace` resources.

We continue to support Zipkin and Datadog:

```
apiVersion: kuma.io/v1alpha1
kind: MeshTraceBackend
metadata:
 name: jaeger
spec:
  zipkin:
    url: http://jaeger-collector.mesh-observability:9411/api/v2/spans
```

```
apiVersion: kuma.io/v1alpha1
kind: MeshTraceBackend
metadata:
 name: datadog
spec:
  datadog:
    address: trace-svc.datadog.svc.cluster.local
    port: 8126
```

There are still open discussions about how exactly we can best represent these
sum types, for example `kind: zipkin` vs `zipkin: ...`, but we chose the field
name discriminator option as above.

#### Backend array with one element

Envoy allows configuring only 1 backend,
so the natural way of representing that would be just one object.
Unfortunately, to make merging for rule based view work properly
we had to change it to be an array `backends`
with a validation that checks if it contains only one object.

### `targetRef`

This is a new policy, so it's based on `targetRef`. Envoy tracing can be configured
on both HTTP connection managers and routes.
This proposal gives `MeshTrace` a single `spec.targetRef` field
that selects what it applies to. It does not use `to` or `from` fields.

It is theoretically possible to support `to` fields, but we omit them from this
proposal.

All tracing configuration happens under `spec.default` so that users are able to
override settings with more specific `targetRef`s.

Resources supported by `spec.targetRef` are `Mesh`, `MeshSubset`, `MeshService`,
`MeshServiceSubset`, `MeshGatewayRoute` and eventually the future
evolution of `TrafficRoute`, which we'll call `MeshTrafficRoute` in this
document.

#### Considered alternatives

We considered implementing `to` field to support a case
where we want to set different sampling for different outbound traffic.

Without `to` field trying to set sampling to `0%` to `Analytics` would not be possible:
- targeting `Client` as a `MeshService` would set sampling for both `Database` and `Analytics`
- targeting `Database` and `Analytics` would not make a difference 
because they are not the originating the traffic (see [Sampling in Envoy](#sampling-in-envoy))

```text
Gateway -> Client --(constant traffic - second in chain)------> Server         # set sampling to 10%
              \-----(once every 1h - originating)-------------> Database       # set sampling to 100%
               \----(once every 1m - originating)-------------> Analytics      # set sampling to 0%
```

Implementing `to` implicitly requires implementing `from`,
otherwise the traces would be missing inbound spans.
Implementing `from` implicitly (as a syntax sugar, not actually requiring to have `from` field) 
is not possible in a case where `to` has multiple different backends,
because we wouldn't know which backend to use.

Example:

```yaml
apiVersion: kuma.io/v1alpha2
kind: MeshTrace
spec:
  targetRef:
    kind: Mesh
# !implicit definition
#  from: 
#  - targetRef:
#     kind: Mesh
#    default:
#      backend:
#        reference:
#          kind: MeshTraceBackend
#          name: jaeger-? which to choose here?
  to:
  - targetRef:
     kind: MeshService
     name: database
    default:
      backends:
        - reference:
          kind: MeshTraceBackend
          name: jaeger-1
  - targetRef:
      kind: MeshService
      name: analytics
    default:
      backends:
        - reference:
          kind: MeshTraceBackend
          name: jaeger-2
```

This makes the config quite verbose and that's why we decided not to got down this route.

### Policy-specific configuration

This MADR proposes additional options be configurable on `MeshTrace` resources
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
   overall:
    value: <percentage>
   client:
    value: <percentage>
   random:
    value: <percentage>
```

with the justification that it's not clear to users
whether traces started via `x-client-trace-id` are limited
when they set the simple `sampling` field we have now.

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

One future feature might be adding tags from the `Dataplane` object to the
trace.

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
`core` to spans. The backend is specified in a `MeshTraceBackend` object.

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshTrace
metadata:
 name: all
 labels:
  kuma.io/mesh: default
spec:
 targetRef:
  kind: MeshService
  name: backend
 default:
  backends:
   - reference:
     kind: MeshTraceBackend
     name: jaeger
  sampling:
   overall:
     value: 80
  tags:
   - name: team
     literal: core
```

### Inline backend configuration

This specifies the backend directly:

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshTrace
metadata:
 name: all
 labels:
  kuma.io/mesh: default
spec:
 targetRef:
  kind: MeshService
  name: backend
 default:
  backends:
   - zipkin:
     url: http://jaeger-collector.mesh-observability:9411/api/v2/spans
  sampling:
   overall:
     value: 80
```

### Route-specific

This configures any listeners matched by the `MeshGatewayRoute` `prod`
to trace a maximum of 80% of traffic and adds the custom tag `env`
with a value from `x-env` and a default value of `prod`.

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshTrace
metadata:
 name: all
 labels:
  kuma.io/mesh: default
spec:
 targetRef:
  kind: MeshGatewayRoute
  name: prod
 default:
  backends:
   - reference:
     kind: MeshTraceBackend
     name: jaeger
  sampling:
   overall:
     value: 80
  tags:
   - name: env
     header:
      name: x-env
      default: prod
```
