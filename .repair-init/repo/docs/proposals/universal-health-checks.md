# Universal Health Check

## Context
Since `Kuma 1.0.4` Dataplane model has a new field `health`:

```yaml
type: Dataplane
mesh: default
name: web-01
networking:
  address: 127.0.0.1
  inbound:
    - port: 11011
      health:
        ready: true
      servicePort: 11012
      tags:
        kuma.io/service: backend
        kuma.io/protocol: http
    - port: 11013
      health:
        ready: false
      servicePort: 11014
      tags:
        kuma.io/service: backend
        kuma.io/protocol: http

```

This field is intended to be set automatically by various health checkers. When Kuma runs in Kubernetes this field is automatically set by Kuma CP based on data from `Pod.Status`. But when Kuma runs in Universal this doesn’t happen.

Thanks to Envoy HDS (Health Discovery Service)  we can configure Envoy to do active health checking of application’s ports and report status back to management server. PR [#1418](https://github.com/kumahq/kuma/pull/1418)  provides basic implementation of HDS, but it doesn’t introduce good way to configure health checks.

There are several types of health checkers - TCP, HTTP and GRPC, each of them has own set of parameters. We want flexibility in specifying those health checks on a per Dataplane basis. In order to do that we have to either extend of the existing policy or introduce new one. Current proposal aims to cover all possible options with pros and cons.

## Requirements

1. Enable/Disable application health check for all Dataplanes or only for specific ones.
2. Parametrise port and path for health checkers (if application has special endpoints for checking health)

## Possible solutions

### Extend Dataplane model

```yaml
type: Dataplane
mesh: default
name: web-01
networking:
  address: 127.0.0.1
  inbound:
    - port: 11011
      servicePort: 11012
      tags:
        kuma.io/service: backend
        kuma.io/protocol: http
probe:
  checkInbounds: true
  readiness:
    interval: 5s
    timeout: 1s
    http:
      path: /health
      port: 80
```

### Extend HealthCheck policy

New section `AppHealthChecks` for `HealthCheck` policy:

```yaml
type: HealthCheck
name: web-to-backend-check
mesh: default
sources:
- match:
    kuma.io/service: web
destinations:
- match:
    kuma.io/service: backend
conf:
  interval: 10s
  timeout: 2s
  unhealthyThreshold: 3
  healthyThreshold: 1
  protocol:
    tcp: {}
  appHealthChecks:
    ...
```

*Downsides:*
* because of the way we pick matched policy it will be problematic to support.
* it is implied that HealthCheck is matched for pair - source + destination, application health check is designed to be applied on a per Dataplane basis

### Introduce new policy Probe

```yaml
type: Probe
mesh: default
name: custom-probe
selectors:
  - match:
      kuma.io/service: '*'
conf:
  readiness:
    interval: 5s
    timeout: 1s
    http:
      path: /health
      port: 80
    tcp: 
      port: 9000
```

```yaml
type: Probe
mesh: default
name: default-probe
selectors:
  - match:
      kuma.io/service: '*'
conf:
  readiness:
    tcp: {}
```

## Solution

Extend dataplane module, specifically Inbound section:

```yaml
type: Dataplane
mesh: default
name: web-01
networking:
  address: 127.0.0.1
  inbound:
    - port: 11011
      servicePort: 11012
      tags:
        kuma.io/service: backend
        kuma.io/protocol: http
      serviceProbe:
        interval: 10s
        timeout: 3s
        healthyThreshold: 1
        unhealthThreshold: 1
        tcp: {}
```

Parameters `interval`, `timeout`, `healthyThreshold` and `unhealthyThreshold` could be omitted, default values for them could be specified in kuma-cp config. 
In the future we can add `http` checker section as well.
