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

TODO:

## Design

### Requirement #1: Support `Envoy -> upstream` "health checks"

TODO:

### Requirement #2: Support `Envoy -> local app` "health checks"

TODO:

### Requirement #3: Forward results of "health checks" to the Control Plane

TODO:

### Requirement #4: Aggregate results of "health checks" in the Control Plane

TODO:

### Requirement #5: Include "health statuses" into EDS configuration sent to Envoy

TODO:

## Consideration Notes

TODO:

## Implementation notes

TODO:

## Examples

TODO:
