# Health Check Policy

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/4735

## Context and Problem Statement

We want to create a new policy for Health Checking that uses TargetRef [matching](https://github.com/kumahq/kuma/blob/22c157d4adac7f518b1b49939c7e9ea4d2a1876c/docs/madr/decisions/005-policy-matching.md).

We want to extend the current list of supported protocols (TCP, HTTP) with GRPC.

As a reminder: this is active health checking so this policy will attach to an outbound.

## Specification

### Matching

#### From level

The way active health checking works is by sending outgoing requests, I don't think it makes sense to have "from" section.

#### Top & to level

`Top` and `to` levels look to me as mirror cases, so I think it's worth talking about them together.

`MeshGatewayRoute` and `MeshHTTPRoute` do not make sense as on its own they do not have "health",
also Envoy does set health on a route only per Cluster/Endpoint.

`MeshService` is a natural fit, but I think we should be careful when allowing other targets combinations.

##### Option 1 - allow everything

###### Pros

- Flexibility
- Can avoid DOS-ing by using reachable services

###### Cons

- Users might accidentally DOS their infra by making Everything HC service X
- Without reachable services using `targetRef=Mesh` and `to.targetRef=Mesh` does not make sense
- Cases that might produce a lot of traffic need to be documented

##### Option 2 - disallow Mesh-Mesh

###### Pros

- Flexibility (without 1 case)
- Can avoid DOS-ing by using reachable services

###### Cons

- Users might accidentally DOS their infra by making Everything HC service X
- Cases that might produce a lot of traffic need to be documented

##### Option 3 - allow only MeshService

###### Pros

- Harder to accidentally DOS infrastructure by running too aggressive health checks

###### Cons

- Inflexibility

#### Decision drivers

As we don't pose any restrictions on HC right now I would go with **option 1** and an entry in docs about possible problems.

### GRPC

GRPC support seems pretty straight forward,
just need to use a different health checking type - [GrpcHealthCheck](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/health_check.proto.html?highlight=grpc_health_check#envoy-v3-api-field-config-core-v3-healthcheck-grpc-health-check)

### Configuration options

There are some features that we currently do not expose (like pass through mode / caching, redis / thrift protocol),
but currently this MADR does not aim to change that, if you think something critical is missing from this functionality,
let us know in the PR.

#### GRPC options
For the first iteration we will add `service_name`, `authority` and leave out `initial_metadata` (can be added later).

#### Protocol selection

Pick the most specific protocol depending on the service annotation (`kuma.io/protocol` or `appProtocol`),
fallback to more general protocol when specific protocol has `disabled=true`.

Example algorithm flow:

1. service annotation = http
2. http.disabled == true
3. try tcp (more general proto)
4. tcp.disabled == false
5. use tcp

### All configuration options

```yaml
default:
  interval: 10s
  timeout: 2s
  unhealthyThreshold: 3
  healthyThreshold: 1
  initialJitter: 5s # optional
  intervalJitter: 6s # optional
  intervalJitterPercent: 10 # optional
  healthyPanicThreshold: 60 # optional, by default 50
  failTrafficOnPanic: true # optional, by default false
  noTrafficInterval: 10s # optional, by default 60s
  eventLogPath: "/tmp/health-check.log" # optional
  alwaysLogHealthCheckFailures: true # optional, by default false
  reuseConnection: false # optional, by default true
  tcp: # it will pick the protocol as described in 'protocol selection' section
    disabled: true # new, default false, can be disabled for override
    send: Zm9v # optional, empty payloads imply a connect-only health check
    receive: # optional
    - YmFy
    - YmF6
  http:
    disabled: true # new, default false, can be disabled for override
    path: /health
    requestHeadersToAdd: # optional, empty by default
    - append: false
      header:
        key: Content-Type
        value: application/json
    - header:
        key: Accept
        value: application/json
    expectedStatuses: [200, 201] # optional, by default [200]
  grpc: # new
    disabled: false # new, default false, can be disabled for override
    serviceName: "" # optional, service name parameter which will be sent to gRPC service
    authority: "" # optional, the value of the :authority header in the gRPC health check request, by default name of the cluster this health check is associated with
```

## Examples

### GRPC

```yaml
type: MeshHealthCheck
mesh: default
name: hc-all-grpc
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: Mesh
        name: default
      default:
        grpc: {}
```

### Override

```yaml
type: MeshHealthCheck
mesh: default
name: hc-all-tcp
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: Mesh
        name: default
      default:
        tcp: {}
```

```yaml
type: MeshHealthCheck
mesh: default
name: hc-front-to-back-http
spec:
  targetRef:
    kind: MeshService
    name: frontend
  to:
    - targetRef:
        kind: MeshService
        name: backend
      default:
        tcp:
          disabled: true
        http:
          path: /health
```
