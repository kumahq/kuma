# Fault Injection

## Context

Fault Injection is a useful feature that allows to test resiliency of the service. 
Envoy provides full-featured support of Fault Injection so we intended 
to implement same functionality in Kuma.

## Proposed configuration model

```yaml
type: FaultInjection
mesh: default
name: fi-1
sources:
  - match:
      service: frontend
      region: aws
      version: 3
destinations:
  - match:
      service: backend
      version: 5
conf:
  delay:
    percentage: 35.5 # from 0.0 to 100.0
    value: 50ms
  abort:
    percentage: 87.5
    httpStatus: 500
  responseBandwidth: # response_rate_limit in envoy
    percentage: 87.5
    limit: 500kbps # bps, mbps
```

## Considerations

There are two possible places to configure Envoy proxy: 
- outbound listener of source
- inbound listener of destination

### Fault Injection on the source outbound

Probably will be much easier to implement Fault Injections for external service in the future.

**Limitations:** unable to differentiate destinations somehow besides service's name. 
Potentially could be resolved the same way like `TrafficRouter`, but this is undesirable. 

### Fault Injection on the destination inbound

More preferable way to implement Fault Injection. But problem with traffic differentiation is 
still exist. But in that case problem could be solved using HTTP Headers.

Kuma can define and document a set of headers like this:
```bash
x-kuma-tag-service: frontend
x-kuma-tag-region: aws
x-kuma-tag-version: 3
```

Source outbound listeners will be configured to set such headers.   
