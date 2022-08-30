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

#### Backends

Currently, tracing support requires defining _backends_ in the `Mesh` resource. A backend defines where traces are sent, the _provider_. It also defines sampling rates. The `TrafficTrace` resource references these backends and contains no configuration of its own.

### Variant 1

This MADR proposes leaving the _provider config_ in the `Mesh`.
It leaves the door open for defining providers inline or in a separate
resource.

However, we propose configuring sampling and tag options in the `Tracing`
resource.

#### Custom tags

Envoy can set tags in traces using values from:

- Literal value
- Request header
- Environment variable
- Metadata

We propose allowing users to configure custom tags using only literal values or
request headers. Using Kuma leaves environment variables and Envoy metadata
opaque for users so the use cases for configuring them (at least directly)
in Kuma are limited.

#### Proposed schema

##### `targetRef`

This is a new policy so it's based on `targetRef`. Envoy tracing is configured
on the HTTP connection manager so `Tracing` has a single `spec.targetRef` field
that selects what it applies to. It does not use `to` or `from` fields.

All logging configuration happens under `spec.default` so that users are able to
override settings with more specific `targetRef`s.

Resources supported by `spec.targetRef` are `Mesh`, `MeshSubset`, `MeshService`,
`MeshServiceSubset`, `MeshGatewayRoute` and eventually the future
evolution of `TrafficRoute`, which we'll call `MeshTrafficRoute` in this
document.

##### Backends

We propose keeping backends definable in the `Mesh` resource as is.

```yaml
spec:
  default:
    backend:
      name: <backend defined in `Mesh`>
```

In particular, these backends contain the [_provider
configs_](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/trace/v3/http_tracer.proto#envoy-v3-api-msg-config-trace-v3-tracing-http).

The rationale here is that the provider is likely to be the same provider for all
traces of a `Mesh` so it makes sense to configure it on the `Mesh`.

###### Further provider config options

There are two more possibilities worth mentioning.

We could easily offer defining a `Tracing`-specific backend inline:

```yaml
spec:
  default:
    backendConfig:
      type: <backend type>
      <type-specific config>
```

We could create an additional resource `TracingConfig` (name tbd) that can be referenced
from `Tracing`:

```yaml
spec:
  default:
    configRef:
      name: <resource-name>
```

##### More knobs

This MADR proposes any additional options be configurable on `Tracing` resources
directly.

###### Sampling

At the moment, tracing backends only support Envoy's
[`overall_sampling`](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/network/http_connection_manager/v3/http_connection_manager.proto#envoy-v3-api-field-extensions-filters-network-http-connection-manager-v3-httpconnectionmanager-tracing-overall-sampling)
via the `sampling` field.

Question: should we allow the following instead?

```yaml
spec:
  default:
    sampling:
      overall: <percentage>
      client: <percentage>
      random: <percentage>
```

If users set the simple `sampling` field we have now,
do they expect it to limit traces started via `x-client-trace-id`?

##### Tags

Tags can be configured as well:

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

##### Method/path specific

Users can set the `spec.targetRef` to be a `MeshGatewayRoute` or
`MeshTrafficRoute`.

TODO: is this enough?

### Examples

All examples assume Kubernetes.

#### Simple

This configures all `Dataplane` inbounds with `kuma.io/service: backend` to
trace a maximum of 80% of traffic and adds the custom tag `team` with a value of
`core` to spans.

```yaml
apiVersion: kuma.io/v1alpha1
kind: TrafficLog
metadata:
  name: all
  labels:
    kuma.io/mesh: default
spec:
  targetRef:
    kind: MeshService
    name: backend
  default:
    backend: jaeger
    sampling:
      overall: 80
    tags:
      - name: team
        literal: core
```

#### Route-specific

This configures any listeners matched by the `MeshGatewayRoute` `prod`
to trace a maximum of 80% of traffic and adds the custom tag `env`
with a value from `x-env` and a default value of `prod`.

```yaml
apiVersion: kuma.io/v1alpha1
kind: TrafficLog
metadata:
  name: all
  labels:
    kuma.io/mesh: default
spec:
  targetRef:
    kind: MeshGatewayRoute
    name: prod
  default:
    backend: jaeger
    sampling:
      overall: 80
    tags:
      - name: env
        header:
          name: x-env
          default: prod
```
