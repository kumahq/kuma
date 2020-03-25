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

Probably will be much easier to implement Fault Injections for external services in the future.

**Limitations:** unable to differentiate destinations somehow besides service's name. 
Potentially could be resolved the same way like `TrafficRoute`, but this is undesirable because 
we have to generate cluster for every destination combination which leads to complexity of reading and
aggregating of metrics. Also this is performance hit on having such combinations in envoy configs, they become bigger. 

### Fault Injection on the destination inbound

More preferable way to implement Fault Injection. But problem with traffic differentiation 
still has to be solved. The idea is to implement matching by HTTP Headers.

Envoy Fault Injection filter allows to specify regex for headers and be applied only for matched ones.   
Kuma can reserve header `x-kuma-match` and configure source proxy to set it on every request. We can 
consider the format for that header, but probably the simplest one is URL-style: `service=frontend&version=0.1`. 

**_Why don't just have separate header for each tag?_**

_Alternatively we could reserve header pattern like `x-kuma-tag-TAG_NAME` and store each tag separately. 
But in that case we need envoy to have ability to remove headers with specified prefix (or with regex).
There is no such functionality yet._

On the destination side proxy we will configure regex for matching:
```
(?=.*version=0\.1.*)(?=.*service=frontend.*)
```
So it will match:
- `service=frontend&version=0.1` 
- `version=0.1&service=frontend` 
- `service=frontend&version=0.1&tag=customtag`
- ...

In other words all key-values pairs that necessarily contain provided version and service.
Also on destination side we will configure HTTP filter that removes specified headers. 
So application will see the request as it was sent by source. 

_Note: envoy will replace header if some malicious application tries to specify it by itself pretending to be identified like
another service. So that feature requres to have mTLS enabled._ 

## Summary
At that point we decided to move forward with implementation of the second option - Fault Injection on the destination inbound. 
That gives us more flexibility for traffic management with less efforts. As soon as External Services support will be added 
to Kuma for L7 traffic, we will come back to Fault Injections and to that proposal.