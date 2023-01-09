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

#### Top level

Top-level targetRef can have all available kinds:
```yaml
targetRef:
 kind: Mesh|MeshSubset|MeshService|MeshServiceSubset
 name: ...
```
Fault injection is an inbound policy so only `from` should be configured.

#### From level

We can allow configuration from different sources.
```yaml
from:
 - targetRef:
     kind: Mesh|MeshSubset|MeshService|MeshServiceSubset
     name: ...
```
The only issue is `MeshHealthCheck` which doesn't set the `x-kuma-tags` flag and might not match the specific fault injection configuration. We have 2 options:

* set `x-kuma-tags` header by default for `MeshHealthCheck`
* change the scope of `From` to `Mesh` so the policy will affect all sources

#### Configuration

```yaml
default:
  disabled: false
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

#### **Result**

```yaml
type: MeshFaultInjection
mesh: example
name: example-fault-injection
spec:
 targetRef:
   kind: Mesh|MeshSubset|MeshService|MeshServiceSubset
   name: example-fault-injection
 from:
   - targetRef:
       kind: Mesh|MeshSubset|MeshService|MeshServiceSubset
       name: backend
       mesh: example
     default:
       disabled: false
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
        abort:
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
        delay:
          percentage: 50.5
          value: 5s
```

#### All services to one service fault injection disabled

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
        disabled: true
        delay:
          percentage: 50.5
          value: 5s
```
