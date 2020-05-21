# Circuit Breaker

## Context

### Design pattern

Circuit breaker is a popular design pattern in a distributed system world. It protects caller from long hanging in timeout if something unexpectedly goes wrong on the server side. Circuit breaker has 2 states - OPEN and CLOSED (sometimes also HALF OPEN). An initial state is CLOSED and all requests are proxied successfully to the server. If some threshold is reached (success rate or several consecutive errors) then circuit breaker switches to OPEN state. In OPEN state circuit breaker returns error immediately without an actual requesting of the server. Eventually, after a while, circuit breaker switches to CLOSED state again. 

### Envoy support 

Envoy allows to configure `cluster.CircuitBreakers` on each Cluster:
```json
{
  "thresholds": []
}
```
Where `thresholds` is an array of elements:
```json
{
  "priority": "...",
  "max_connections": "{...}",
  "max_pending_requests": "{...}",
  "max_requests": "{...}",
  "max_retries": "{...}",
  "retry_budget": "{...}",
  "track_remaining": "...",
  "max_connection_pools": "{...}"
}
```
So circuit breaker in Envoy is a bunch of thresholds which prevents service from opening more new connections and create more new requests than configured. Also, it could limit the number of retries that happen in parallel. 

That functionality of Envoy doesn't allow implementing design pattern described above. In order to make it work there is another useful Envoy feature called `cluster.OutlierDetection`:
```json
{
  "consecutive_5xx": "{...}",
  "interval": "{...}",
  "base_ejection_time": "{...}",
  "max_ejection_percent": "{...}",
  "enforcing_consecutive_5xx": "{...}",
  "enforcing_success_rate": "{...}",
  "success_rate_minimum_hosts": "{...}",
  "success_rate_request_volume": "{...}",
  "success_rate_stdev_factor": "{...}",
  "consecutive_gateway_failure": "{...}",
  "enforcing_consecutive_gateway_failure": "{...}",
  "split_external_local_origin_errors": "...",
  "consecutive_local_origin_failure": "{...}",
  "enforcing_consecutive_local_origin_failure": "{...}",
  "enforcing_local_origin_success_rate": "{...}",
  "failure_percentage_threshold": "{...}",
  "enforcing_failure_percentage": "{...}",
  "enforcing_failure_percentage_local_origin": "{...}",
  "failure_percentage_minimum_hosts": "{...}",
  "failure_percentage_request_volume": "{...}"
}
```

It gives us an ability to configure criteria and conditions for how long a specific Endpoint will be ejected from the load balancing set. That's the thing I believe people expect to see when it comes to Circuit breakers.

## Proposed configuration model

```yaml
type: CircuitBreaker
mesh: default
name: cb-1
sources:
  - match:
      service: frontend
      region: aws
      version: 3
destinations:
  - match:
      service: backend
conf:
  interval: 1s # time interval between ejection analysis sweeps
  baseEjectionTime: 30s # the base time that a host is ejected for. The real time is equal to the base time multiplied by the number of times the host has been ejected
  maxEjectionPercent: 20 # the maximum percent of an upstream cluster that can be ejected due to outlier detection
  splitExternalAndLocalErrors: false # enables Split Mode in which local and external errors are distinguished 
  detectors:
    errors: # detects consecutive errors
      total: 20 # errors with status code 5xx and locally originated errors, in Split Mode - just errors with status code 5xx
      gateway: 10 # subset of 'total' related to gateway errors (502, 503 or 504 status code)
      local: 7 # takes into account only in Split Mode, number of locally originated errors
    standardDeviation: # detection based on success rate, aggregated from every host in the cluser
      requestVolume: 10 # ignore hosts with less number of requests than 'requestVolume'
      minimumHosts: 5 # won't count success rate for cluster if number of hosts with required 'requestVolume' is less than 'minimumHosts'
      factor: 1.9 # resulting threshold = mean - (stdev * success_rate_stdev_factor)
    failure: # detection based on success rate, but threshold is set explicitly (unlike 'standardDeviation')
      requestVolume: 10 # ignore hosts with less number of requests than 'requestVolume'
      minimumHosts: 5 # won't count success rate for cluster if number of hosts with required 'requestVolume' is less than 'minimumHosts'
      threshold: 85 # eject host if failure percentage of a given host is greater than or equal to this value
```

Fields `interval`, `baseEjectionTime`, `maxEjectionPercent` and `splitExternalAndLocalErrors will be mapped respectively to corresponding Envoy fields. 
Detectors will be mapped in the following way:

- `errors.total` -> `consecutive_5xx`
- `errors.gateway` -> `consecutive_gateway_failure`
- `errors.local` -> `consecutive_gateway_failure`
- `standardDeviation.minimumHosts` -> `success_rate_minimum_hosts`
- `standardDeviation.requestVolume` -> `success_rate_request_volume`
- `standardDeviation.factor` -> `success_rate_stdev_factor`
- `failure.minimumHosts` -> `failure_percentage_minimum_hosts`
- `failure.requestVolume` -> `failure_percentage_request_volume`
- `failure.threshold` -> `failure_percentage_threshold`
