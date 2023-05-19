# Rate limit policy
 
- Status: accepted
 
Technical Story: https://github.com/kumahq/kuma/issues/4740
 
## Context and Problem Statement
 
We want to create a [new policy matching compliant](https://github.com/kumahq/kuma/blob/22c157d4adac7f518b1b49939c7e9ea4d2a1876c/docs/madr/decisions/005-policy-matching.md)
resource for managing rate limiting.
 
Rate Limiting in Kuma is implemented with Envoy's [rate limiting
support](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/local_rate_limit_filter).
There is an issue with the configuration of TCP rate limiting because [current filter](https://www.envoyproxy.io/docs/envoy/latest/configuration/listeners/network_filters/local_rate_limit_filter#config-network-filters-local-rate-limit) works on listener level which can only
allow limiting connections on the specific listener but without checking the origin of the request.
 
## Considered Options
 
* Create a MeshRateLimit with basic TCP rate limit support
* Create a MeshHTTPRateLimit and MeshTCPRateLimit with basic TCP rate limit support
* Create advanced TCP rate limiting per requested service

## Decision Outcome
 
Chosen option: create a MeshRateLimit with basic TCP rate limit support
 
## Decision Drivers
 
- Replace existing policies with new policy-matching compliant resources
- Add support for TCP connection limiting at the basic level
- Possibility to add global rate limiting in the future
 
## Considered Options
 
### Naming
 
- **MeshRateLimit**
- MeshHTTPRateLimit and MeshTCPRateLimit
 
## Solution
 
### Current configuration
Below is a sample rate-limiting configuration.
 
```yaml
spec:
 sources:
   - match:
       kuma.io/service: "redis_kuma-demo_svc_6379"   
 destinations:
   - match:
       kuma.io/service: "demo-app_kuma-demo_svc_5000"
 conf:
   http:
     requests: 5
     interval: 10s
     onRateLimit:
       status: 423
       headers:
         - key: "x-Kuma-rate-limited"
           value: "true"
           append: true
```
We are using [Envoy's local rate limit](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/local_rate_limit_filter) which supports only HTTP configuration.
 
The configuration translates to the Envoy configuration on the route and the listener filter:
 
```yaml
{
 "match": {
   "prefix": "/",
   "headers": [
   {
     "name": "x-kuma-tags",
     "safe_regex_match": {
     "google_re2": {},
     "regex": ".*&kuma.io/service=[^&]*redis_kuma-demo_svc_6379[,&].*"
     }
   }
   ]
 },
 "route": {
   "cluster": "localhost:5000",
   "timeout": "0s"
 },
 "typed_per_filter_config": {
   "envoy.filters.http.local_ratelimit": {
   "@type": "type.googleapis.com/envoy.extensions.filters.http.local_ratelimit.v3.LocalRateLimit",
   ...
   }
 }
}
 
```
 
### Specification
 
The differences between TCP and HTTP rate limiting in `targetRef` and `from` requires different validation of objects.
We create one policy which can configure both TCP and HTTP rate limiting. Because, we can't identify incoming traffic when using TCP users can only set the TCP section when using the `Mesh` type of targetRef in the `from` section. For http these types of targetRef can be used: `Mesh|MeshSubset|MeshService|MeshServiceSubset`.
 
#### **HTTP Rate Limit**
 
Rate limiting can be configured on both HTTP connection managers and routes. Thanks to the header `x-kuma-tags`, which is propagated and set by Envoy, we can recognize the origin of the HTTP request.
 
#### Top level
 
Top-level targetRef can have all available kinds:
```yaml
targetRef:
 kind: Mesh|MeshSubset|MeshService|MeshServiceSubset|MeshGatewayRoute|MeshHTTPRoute
 name: ...
```
Rate limiting is an inbound policy so only `from` should be configured.
 
#### From level
 
```yaml
from:
 - targetRef:
     kind: Mesh|MeshSubset|MeshService|MeshServiceSubset
     name: ...
```
 
Matching on MeshGatewayRoute and MeshHTTPRoute does not make sense (there isn't a route that a request originates from).
 
#### Configuration
 
```yaml
default:
  local:
    http:
      requests: 5
      interval: 10s
      onRateLimit:
        status: 423
        headers:
          - key: "x-kuma-rate-limited"
            value: "true"
            append: true
```

#### Rate limiting of External Services
`ExternalServices` are services running outside of the mesh, hence we are not able to configure their inbounds. We are going to support rate limiting of `ExternalServices` only when `ZoneEgress` is enabled.
 
#### **TCP Connection Rate Limit**
 
Rate limiting on TCP connection is more complicated because there is no easy way to match only for specific clients.
Based on alternatives we decided to make it only available for all the clients.
 
#### Top level
 
```yaml
targetRef:
 kind: Mesh|MeshSubset|MeshService|MeshServiceSubset
 name: ...
```
Matching on MeshGatewayRoute and MeshHTTPRoute does not make sense because we are configuring TCP listeners.
Rate limiting is an inbound policy so only `from` should be configured.
 
#### From level
 
```yaml
from:
 - targetRef:
     kind: Mesh
     name: ...
```
 
There is no easy way to match requesting services with the specific filter chain that's why we decided to support only configuring TCP rate limiting from all services.
 
#### Configuration
 
```yaml
default:
  local:
    tcp:
      connections: 100
      interval: 10s
```
 
#### **Result**
 
```yaml
type: MeshRateLimit 
mesh: example 
name: example-rate-limit
spec:
 targetRef:
   kind: Mesh|MeshSubset|MeshService|MeshServiceSubset|MeshGatewayRoute|MeshHTTPRoute # or Mesh|MeshSubset|MeshService|MeshServiceSubset for TCP configuration
   name: example-rate-limit
 from:
   - targetRef:
       kind: Mesh|MeshSubset|MeshService|MeshServiceSubset # or Mesh for TCP configuration
       name: backend
       mesh: example
     default: 
       local:
         http:
           disabled: false
           requests: 5
           interval: 10s
           onRateLimit:
             status: 423
             headers:
               - key: "x-kuma-rate-limited"
                 value: "true"
                 append: true
         tcp:
           disabled: false
           connections: 100
           interval: 10s
```
 
### Considered Options
 
#### TCP rate limit for a specific origin
 
Configuration of connection limits is done on listeners. In case of matching from which service a request arrived, we
need to get some information about the request. There are two options:
 
* use SourceIP filter matching
* use SNI filter matching
 
##### Use SourceIP matching
 
In this case, we need to configure a listener filter matcher with a list of all possible IPs from which requests can arrive. In the case of one service, it can be many IPs. Matching on many ips might not be the most efficient way and each change of IP requires listener reload.
 
##### Use SNI filter matching
 
It requires a change of SNI that is set on cluster `service_name{mesh=name}` to something different `service_name{mesh=name}{origin=requesting_service}` to later match it in the filter chain. That is not the best idea and can cause issues in multi-zone setups. Apart from that, it requires mTLS to be enabled.

### Examples
#### Service to service http rate limit

```yaml
type: MeshRateLimit  
mesh: default  
name: default-rate-limit  
spec:  
  targetRef:  
    kind: MeshService  
    name: backend
  from:
    - targetRef:
        kind: MeshService
        name: frontend
      default:
        local:
          http:
            requests: 5
            interval: 10s
            onRateLimit:
              status: 423
              headers:
                - key: "x-kuma-rate-limited"
                  value: "true"
                  append: true
```

#### All services to one service http rate limit

```yaml
type: MeshRateLimit  
mesh: default  
name: default-rate-limit  
spec:  
  targetRef:  
    kind: MeshService  
    name: backend
  from:
    - targetRef:
        kind: Mesh
        name: default
      default:
        local:
          http:
            disabled: false
            requests: 5
            interval: 10s
            onRateLimit:
              status: 423
              headers:
                - key: "x-kuma-rate-limited"
                  value: "true"
                  append: true
```

#### All services to specific service TCP rate limit

```yaml
type: MeshRateLimit  
mesh: default  
name: default-rate-limit  
spec:  
  targetRef:  
    kind: MeshService  
    name: backend
  from:
    - targetRef:
        kind: Mesh
        name: default
      default:
        local:
          tcp:
            disabled: false
            connections: 5
            interval: 10s
```

#### All services to specific service TCP rate limit and HTTP rate limit

```yaml
type: MeshRateLimit  
mesh: default  
name: default-rate-limit  
spec:  
  targetRef:  
    kind: MeshService  
    name: backend
  from:
    - targetRef:
        kind: Mesh
        name: default
      default:
        local:
          tcp:
            disabled: false
            connections: 5
            interval: 10s
          http:
            false: true
            requests: 5
            interval: 10s
            onRateLimit:
              status: 423
              headers:
                - key: "x-kuma-rate-limited"
                  value: "true"
                  append: true
```

#### Disable parent HTTP rate limit in backend service from frontend service

```yaml
type: MeshRateLimit  
mesh: default  
name: default-rate-limit  
spec:  
  targetRef:  
    kind: MeshService  
    name: backend
  from:
    - targetRef:
        kind: MeshService
        name: frontend
      default:
        local:
          http:
            disabled: true
```