# L4 HealthChecks

## Context

* at the moment, default configuration for dataplanes doesn't include health checks
* without health checks (or, alternatively, without health statuses delivered via `EDS`), Envoy cannot exclude a non-responsive upstream endpoint from the list of valid targets for load balancing
* there are several ways how Envoy can be configured to obtain health statuses for upstream endpoints
  * Envoy can do "active" health checks
    * for TCP it means that Envoy will periodically try establishing a connection
  * Envoy can do "passive" health checks
    * for TCP it means that Envoy will be remembering statuses of recent connection attempts
  * Envoy can receive health statuses from the Control Plane as part of EDS API

## Requirements

1. support `Envoy -> upstream` "health checks"
2. support `Envoy -> local app` "health checks"
3. forward results of "health checks" to the Control Plane
4. aggregate results of "health checks" in the Control Plane
5. include "health statuses" into EDS configuration sent to Envoy

## Proposed configuration model

### Universal mode

```yaml
type: HealthCheck
name: rule-name
mesh: default
sources:
- match:
    service: '*' # values other than '*' can be set here, but will it be used in practice?
    # NOTE: other tags can be set here, but will it be ever used in practice?
destinations:
- match:
    # NOTE: only `service` tag can be used here
    service: backend
conf:
  activeChecks: # configuration for a TCP check that only verifies whether a connection can be established
    timeout: 1s
    interval: 5s # how often health check requests should be made
    unhealthyThreshold: 3
    healthyThreshold: 2
  passiveChecks:
    unhealthyThreshold: 3
    penaltyInterval: 10s # for how long endpoint should be considered unhealthy
```

## Design

### Requirement #1: Support `Envoy -> upstream` "health checks"

* In order to generate `Cluster`s for a given `Dataplane`:
  1. List all `HealthCheck`s in the same `mesh` as `Dataplane`
  2. For each *outbound* interface of the `Dataplane`
     * Go through `HealthCheck`s and select those where
       * `destinations` selector matches `service` tag of that *outbound* interface
       * `sources` selector matches tags on one of the *inbound* interfaces of that `Dataplane`
     * Order matched `rule`s and keep only 1 "best match"
     * If there is no "best match" `rule`
       * Generate a `Cluster` without "active" and "passive" health checks
     * If there is a "best match" `rule`
       * Generate a `Cluster` with "active" and "passive" health checks according to that rule

### Requirement #2: Support `Envoy -> local app` "health checks"

TODO:

### Requirement #3: Forward results of "health checks" to the Control Plane

TODO:

### Requirement #4: Aggregate results of "health checks" in the Control Plane

TODO:

### Requirement #5: Include "health statuses" into EDS configuration sent to Envoy

TODO:

## Consideration Notes

### Requirement #3: Forward results of "health checks" to the Control Plane

* Envoy already has a couple features we could build on top of, namely
  * [event log](https://github.com/envoyproxy/data-plane-api/blob/master/envoy/data/core/v3alpha/health_check_event.proto) for "active" health checks
  * [event log](https://github.com/envoyproxy/data-plane-api/blob/master/envoy/data/cluster/v2alpha/outlier_detection_event.proto) for "passive" health checks
  * [Health xDS](https://github.com/envoyproxy/data-plane-api/blob/master/envoy/service/discovery/v3alpha/hds.proto) for re-using Envoy as a worker node capable of doing health checks (completely unrelated to regular proxying)

* event log for "active" health checks:
  * at the moment, ONLY file output is supported
    * see an example in the end
  * configured individually for each `Cluster`

* event log for "passive" health checks:
  * at the moment, ONLY file output is supported
    * see an example in the end
  * configured globally for all `Cluster`s as part of `bootstrap` config

* Health xDS:
  * as part of this API, Envoy receives assignments and reports results back to the Control Plane over gRPC
    * see an example in the end
  * is not considered production ready yet (developed in summer 2018 by an intern)
  * does not support mTLS

Conclusions:
* we can already use `Health xDS` for `Envoy -> local app` health checks
* changes to the Envoy will be necessary to use `Health xDS` for `Envoy -> upstream` "health checks" (add support for mTLS)
* changes to the Envoy will be necessary to send event logs to the Control Plane (instead of logging to a file)

## Implementation notes

TODO:

## Examples

### Example 1. Event log for "active" health checks

```json
{
  "health_checker_type": "TCP",
  "host": {
    "socket_address": {
      "protocol": "TCP",
      "address": "127.0.0.1",
      "resolver_name": "",
      "ipv4_compat": false,
      "port_value": 8081               // endpoint #1
    }
  },
  "cluster_name": "service_responder",
  "add_healthy_event": {               // health check PASSED
    "first_check": true
  },
  "timestamp": "2019-11-02T11:04:39.858Z"
}
{
  "health_checker_type": "TCP",
  "host": {
    "socket_address": {
      "protocol": "TCP",
      "address": "127.0.0.1",
      "resolver_name": "",
      "ipv4_compat": false,
      "port_value": 8081               // endpoint #1
    }
  },
  "cluster_name": "service_responder",
  "eject_unhealthy_event": {           // health check FAILED
    "failure_type": "NETWORK"
  },
  "timestamp": "2019-11-02T11:18:39.898Z"
}
{
  "health_checker_type": "TCP",
  "host": {
    "socket_address": {
      "protocol": "TCP",
      "address": "127.0.0.1",
      "resolver_name": "",
      "ipv4_compat": false,
      "port_value": 18081              // endpoint #2
    }
  },
  "cluster_name": "service_responder",
  "health_check_failure_event": {      // health check FAILED
    "failure_type": "NETWORK",
    "first_check": true
  },
  "timestamp": "2019-11-02T11:04:39.859Z"
}
{
  "health_checker_type": "TCP",
  "host": {
    "socket_address": {
      "protocol": "TCP",
      "address": "127.0.0.1",
      "resolver_name": "",
      "ipv4_compat": false,
      "port_value": 18081              // endpoint #2
    }
  },
  "cluster_name": "service_responder",
  "add_healthy_event": {               // health check PASSED
    "first_check": false
  },
  "timestamp": "2019-11-02T11:18:39.901Z"
}
```

### Example 2. Event log for "passive" health checks

```json
{
  "type": "CONSECUTIVE_GATEWAY_FAILURE",
  "cluster_name": "service_responder",
  "upstream_url": "127.0.0.1:18081",
  "action": "EJECT",
  "num_ejections": 0,
  "enforced": false,
  "eject_consecutive_event": {},
  "timestamp": "2019-11-02T11:41:25.814Z"
}
{
  "type": "CONSECUTIVE_5XX",
  "cluster_name": "service_responder",
  "upstream_url": "127.0.0.1:18081",
  "action": "EJECT",
  "num_ejections": 1,
  "enforced": true,
  "eject_consecutive_event": {},
  "timestamp": "2019-11-02T11:41:25.814Z"
}
{
  "type": "CONSECUTIVE_5XX",
  "cluster_name": "service_responder",
  "upstream_url": "127.0.0.1:18081",
  "action": "UNEJECT",
  "num_ejections": 1,
  "enforced": false,
  "timestamp": "2019-11-02T11:41:58.713Z",
  "secs_since_last_action": "32"
}
```

### Example 3. Health xDS

#### HDS server assigns an "active" health check to a Envoy

```
clusterHealthChecks:
- clusterName: 127.0.0.1:8081
  healthChecks:
  - alwaysLogHealthCheckFailures: true
    eventLogPath: /tmp/127.0.0.1_8081.log
    healthyEdgeInterval: 3s
    healthyThreshold: 2
    interval: 3s
    noTrafficInterval: 3s
    tcpHealthCheck: {}
    timeout: 1s
    unhealthyEdgeInterval: 3s
    unhealthyInterval: 3s
    unhealthyThreshold: 3
  localityEndpoints:
  - endpoints:
    - address:
        socketAddress:
          address: 127.0.0.1
          portValue: 8081
    locality:
      region: eu
      subZone: demo
      zone: eu-west-1a
interval: 2s
```

#### Envoy reports results of "active" health check back to the HDS server

```
endpointHealthResponse:
  endpointsHealth:
  - endpoint:
      address:
        socketAddress:
          address: 127.0.0.1
          portValue: 8081
    healthStatus: UNHEALTHY

...

endpointHealthResponse:
  endpointsHealth:
  - endpoint:
      address:
        socketAddress:
          address: 127.0.0.1
          portValue: 8081
    healthStatus: HEALTHY
```

Notice that Envoy's report consists only of 1 field - "health status", one of:
* HEALTHY
* UNHEALTHY
* DRAINING
* TIMEOUT
* DEGRADED
