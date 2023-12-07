#  Fault Injection policy

- Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/4734

## Context and Problem Statement

We want to create a [new policy matching compliant](https://github.com/kumahq/kuma/blob/22c157d4adac7f518b1b49939c7e9ea4d2a1876c/docs/madr/decisions/005-policy-matching.md)
resource for managing fault injection.

FaultInjection in Kuma is implemented with Envoy's [fault injection
support](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/fault_filter).

## Considered Options

* Create a MeshFaultInjection

## Decision Outcome

Chosen option: create a MeshFaultInjection

## Decision Drivers

- Replace existing policies with new policy-matching compliant resources

## Solution

### Current configuration
Below is a sample fault injection configuration.

```yaml
spec:
  sources:
    - match:
        kuma.io/service: frontend_default_svc_80
        kuma.io/protocol: http
  destinations:
    - match:
        kuma.io/service: backend_default_svc_80
        kuma.io/protocol: http
  conf:
    abort:
      httpStatus: 500
      percentage: 50
    delay:
      percentage: 50.5
      value: 5s
    responseBandwidth:
      limit: 50 mbps
      percentage: 50
```

The configuration translates to the Envoy configuration in Http's filters:

```yaml
{
"name": "envoy.filters.http.fault",
"typed_config": {
  "@type": "type.googleapis.com/envoy.extensions.filters.http.fault.v3.HTTPFault",
  "delay": {
  "fixed_delay": "5s",
  "percentage": {
    "numerator": 505000,
    "denominator": "TEN_THOUSAND"
  }
  },
  "abort": {
  "http_status": 500,
  "percentage": {
    "numerator": 50
  }
  },
  "headers": [
  {
    "name": "x-kuma-tags",
    "safe_regex_match": {
    "google_re2": {},
    "regex": ".*&kuma.io/protocol=[^&]*http[,&].*&kuma.io/service=[^&]*frontend_default_svc_80[,&].*"
    }
  }
  ],
  "response_rate_limit": {
  "fixed_limit": {
    "limit_kbps": "50000"
  },
  "percentage": {
    "numerator": 50
  }
  }
}
```

### Specification

Fault injection supports only `HTTP` protocol and is configured in `HttpFilter`. We can use the header `x-kuma-tags` to match the specific source. In the case of `MeshHealthCheck` which doesn't set `x-kuma-tags` we can put `x-kuma-tags` or create `MeshFaultInjection` that allows only `Mesh` in `From` section.

Conclusion:

* set `x-kuma-tags` during health checking 

#### Top level

Top-level targetRef can have all available kinds:
```yaml
targetRef:
 kind: Mesh|MeshSubset|MeshService|MeshServiceSubset|MeshGatewayRoute
 name: ...
```
Fault injection can be configured on inbound and outbound.

#### From level

We can allow configuration from different sources.
```yaml
from:
 - targetRef:
     kind: Mesh|MeshSubset|MeshService|MeshServiceSubset
     name: ...
```
We'll use `x-kuma-tags` headers to select traffic from origin. `MeshHealthCheck` will need to send these headers by implementing: https://github.com/kumahq/kuma/issues/5718

#### To level

The `to` level is only allowed for `spec.kind: MeshGateway` and only `to[].targetRef.kind: Mesh` is permitted. This is because fault injection is configured on a listener and `MeshGateway`'s have only one kind of listener.

```yaml
spec:
  targetRef:
    kind: MeshGateway
    name: edge
to:
 - targetRef:
     kind: Mesh
     name: ...
```

#### Configuration

```yaml
default:
  http:
    - abort:
        httpStatus: 500
        percentage: 50  # we are going to introduce out own type which maps valus
    - abort:
        httpStatus: 404
        percentage: 5
    - delay:
        percentage: 50.5
        value: 5s
    - responseBandwidth:
        limit: 50 mbps
        percentage: 50
```

To support percentage we introduce a custom type to handle decimal values correctly see [#5717](https://github.com/kumahq/kuma/issues/5717).

#### **Result**

```yaml
type: MeshFaultInjection
mesh: example
name: example-fault-injection
spec:
 targetRef:
   kind: Mesh|MeshSubset|MeshService|MeshServiceSubset|MeshGatewayRoute
   name: example-fault-injection
 to:
   - targetRef:
       kind: Mesh|MeshSubset|MeshService|MeshServiceSubset
       name: backend
       mesh: example
     default:
       http:
        - abort:
            httpStatus: 500
            percentage: 50
        - delay:
            percentage: 50.5
            value: 5s
        - responseBandwidth:
            limit: 50 mbps
            percentage: 50
 from:
   - targetRef:
       kind: Mesh|MeshSubset|MeshService|MeshServiceSubset
       name: backend
       mesh: example
     default:
       http:
        - abort:
            httpStatus: 500
            percentage: 50
        - delay:
            percentage: 50.5
            value: 5s
        - responseBandwidth:
            limit: 50 mbps
            percentage: 50
```

### Considered Options

Instead of applying HttpFilters, we could use [matching API](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/advanced/matching/matching_api) but unfortunately, it's still under development.

### Examples
#### Service to service fault injection

```yaml
type: MeshFaultInjection
mesh: default
name: default-fault-injection
spec:
  targetRef:
    kind: MeshService
    name: backend
  from:
    - targetRef:
        kind: MeshService
        name: frontend
      default:
        http:
         - abort:
            httpStatus: 500
            percentage: 50
```

#### All services to one service fault injection

```yaml
type: MeshFaultInjection
mesh: default
name: default-fault-injection
spec:
  targetRef:
    kind: MeshService
    name: backend
  from:
    - targetRef:
        kind: Mesh
        name: default
      default:
        http:
        - delay:
            percentage: 50.5
            value: 5s
```

#### Requests to one service failing

```yaml
type: MeshFaultInjection
mesh: default
name: default-fault-injection
spec:
  targetRef:
    kind: MeshService
    name: backend
  to:
    - targetRef:
        kind: MeshService
        name: backend2
      default:
        http:
        - abort:
            httpStatus: 500
            percentage: 50
```

#### Service to service fault injection with list of faults

```yaml
type: MeshFaultInjection
mesh: default
name: default-fault-injection
spec:
  targetRef:
    kind: MeshService
    name: backend
  from:
    - targetRef:
        kind: MeshService
        name: frontend
      default:
        http:
         - abort:
             httpStatus: 500
             percentage: 2.5
         - abort:
             httpStatus: 500
             percentage: 10
         - delay:
             value: 5s
             percentage: 5
```
