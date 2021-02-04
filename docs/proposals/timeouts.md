# Timeout policy

## Context

Envoy allows users to configure plenty of different timeouts. All of them are applied to the different parts of configuration, all of them protects different stages of Connect / Request / Response process. There are the most important timeouts that Envoy allows to configure:

* [1] [connect_timeout](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/cluster/v3/cluster.proto#envoy-v3-api-field-config-cluster-v3-cluster-connect-timeout) - specifies the amount of time Envoy will wait for an upstream TCP connection to be established.
  
  Place: cluster
  
  Required: true
  
  Default: no
  
  Protocol: TCP


* [2] [idle_timeout](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/network/tcp_proxy/v3/tcp_proxy.proto#envoy-v3-api-field-extensions-filters-network-tcp-proxy-v3-tcpproxy-idle-timeout) - is defined as the period in which there are no bytes sent or received on either the upstream or downstream connection.
  
  Place: TCPProxy
  
  Required: false (0 value for disabling)
  
  Default: 1h
  
  Protocol: TCP


* [3] [idle_timeout](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/protocol.proto#envoy-v3-api-field-config-core-v3-httpprotocoloptions-idle-timeout) - the idle timeout is the time at which a downstream or upstream connection will be terminated if there are no active streams.
  
  Place: HttpConnectionManager, Cluster
  
  Required: false
  
  Default: 1h
  
  Protocol: HTTP


* [4] [request_timeout](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/network/http_connection_manager/v3/http_connection_manager.proto#envoy-v3-api-field-extensions-filters-network-http-connection-manager-v3-httpconnectionmanager-request-timeout) - the amount of time that Envoy will wait for the entire request to be received. The timer is activated when the request is initiated, and is disarmed when the last byte of the request is sent upstream (i.e. all decoding filters have processed the request), OR when the response is initiated. If not specified or set to 0, this timeout is disabled. This timeout is not compatible with streaming requests.
  
  Place: HttpConnectionManager
  
  Required: false (0 value for disabling)
  
  Default: not specified
  
  Protocol: HTTP


* [5] [max_stream_duration](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/protocol.proto#envoy-v3-api-field-config-core-v3-httpprotocoloptions-max-stream-duration) - Total duration to keep alive an HTTP request/response stream. If the time limit is reached the stream will be reset independent of any other timeouts.
  
  Place: HttpConnectionManager, Cluster
  
  Required: false
  
  Default: not specified
  
  Protocol: HTTP


* [6] [stream_idle_timeout](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/network/http_connection_manager/v3/http_connection_manager.proto#envoy-v3-api-field-extensions-filters-network-http-connection-manager-v3-httpconnectionmanager-stream-idle-timeout) - when an initiated request is being sent too slowly in a way that even the headers are not delivered at required intervals, Envoy will respond with a 408 error code.
  
  Place: HttpConnectionManager
  
  Required: false (0 value for disabling)
  
  Default: 5 min
  
  Protocol: HTTP


* [7] [timeout](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-field-config-route-v3-routeaction-timeout) - this spans between the point at which the entire downstream request (i.e. end-of-stream) has been processed and when the upstream response has been completely processed.
  
  Place: Route
  
  Required: false (0 value for disabling)
  
  Default: 15s
  
  Protocol: HTTP


* [8] [idle_timeout](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-field-config-route-v3-routeaction-idle-timeout) - the way to override [stream_idle_timeout](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/network/http_connection_manager/v3/http_connection_manager.proto#envoy-v3-api-field-extensions-filters-network-http-connection-manager-v3-httpconnectionmanager-stream-idle-timeout)  [6]
  
  Place: Route
  
  Required:  false (0 value for disabling)
  
  Default: not specified
  
  Protocol: HTTP


* [9] [per_try_timeout](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-field-config-route-v3-retrypolicy-per-try-timeout) - can be configured when using retries so that individual tries using a shorter timeout than the overall request timeout described above. This timeout only applies before any part of the response is sent to the downstream, which normally happens after the upstream has sent response headers. This timeout can be used with streaming endpoints to retry if the upstream fails to begin a response within the timeout.
  
  Place: Route
  
  Required: false
  
  Default: not specified
  
  Protocol: HTTP


* [10] [max_stream_duration](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-msg-config-route-v3-routeaction-maxstreamduration) - the way to override [max_stream_duration](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/protocol.proto#envoy-v3-api-field-config-core-v3-httpprotocoloptions-max-stream-duration)  [5]
  
  Place: Route
  
  Required:  false (0 value for disabling)
  
  Default: not specified
  
  Protocol: HTTP


* [11] [request_headers_timeout](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/network/http_connection_manager/v3/http_connection_manager.proto#envoy-v3-api-field-extensions-filters-network-http-connection-manager-v3-httpconnectionmanager-request-headers-timeout) - The amount of time that Envoy will wait for the request headers to be received. The timer is activated when the first byte of the headers is received, and is disarmed when the last byte of the headers has been received.
  
  Place: HttpConnectionManager
  
  Required: false (0 value for disabling)
  
  Default: not specified
  
  Protocol: HTTP


* [12] [transport_socket_connect_timeout](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/listener/v3/listener_components.proto#envoy-v3-api-field-config-listener-v3-filterchain-transport-socket-connect-timeout) - If present and nonzero, the amount of time to allow incoming connections to complete any transport socket negotiations. If this expires before the transport reports connection establishment, the connection is summarily closed.
  
  Place: FilterChain
  
  Required: false
  
  Default: not specified

  Protocol: TCP


## Use cases

1. **As a**  service owner
   
   **I want to** propagate default recommended timeouts for applications which consumes my service
   
   **so that** they know approximately how much time it will take to request my service


2. **As a** service owner
   
   **I want to** set service side timeouts
   
   **so that**  I can protect my service from abusive application which is trying to consume it


3. **As a** service consumer
   
   **I want to** set different timeouts for different outbounds of my application
   
   **so that** I can granularly make decisions regarding the time to wait for specific service


4. **As a** service consumer
   
   **I want to** be able to overrider existing default recommended timeouts for the service that I consume
   
   **so that** I can control outgoing timeouts myself


## Proposed solution

In order to satisfy all the use cases we can introduce 2 new policies: `Timeout` and `InboundTimeout`.

### Timeout

Service A -> SRC Envoy -> -> -> DST Envoy -> Service B

Timeout policy is a connection policy so it will be applied to the pair of service (A and B in the example). 
The idea of the policy is to instruct Service A how to request Service B. Example:

```yaml
type: Timeout
mesh: default
name: default-timeouts-srv-B
sources:
- match:
    kuma.io/service: '*'
destinations:
- match:
    kuma.io/service: 'srv-B'
conf:
  # connectTimeout defines time to establish connection, 'connect_timeout' on Cluster [1]
  connectTimeout: 10s 
  tcp:
    # 'idle_timeout' on TCPProxy [2]
    idleTimeout: 1h
  http:
    # 'timeout' on Route [7]
    requestTimeout: 5s
    # 'idle_timeout' on Cluster [3]
    idleTimeout: 1h 
  grpc:
    # 'stream_idle_timeout' on HttpConnectionManager [6]
    streamIdleTimeout: 5m
    # 'max_stream_duration' on Cluster [5]
    maxStreamDuration: 30m
``` 

As an owner of Service B you have some SLA regarding the upper boundaries of the request/response time, so you can apply
this `default-timeouts-srv-B` and every service which is trying to consume Service B will have those timeouts by default.

At the same time if you are an owner of Service A you can override this policy for your service and put different timeout 
values. From the Service B perspective `default-timeouts-srv-B` policy has an advisory nature, every service can override it,
`Timeout` policy is not purposed to protect Service B. In order to protect Service B from abusive clients with long
requests it’s better to use `InboundTimeouts`.

### InboundTimeout

Unlike the `Timeout` this policy is matched for the Dataplane rather than pair of services (applied only to the Dataplanes of the Service B). `InboundTimeout` has the same configuration, but values are mapped to the different Envoy timeouts:

```yaml
type: InboundTimeout
mesh: default
name: timeouts-srv-B
selectors:
  - match:
      kuma.io/service: srv-B
conf:
  connectTimeout: 10s # 'transport_socket_connect_timeout' on FilterChain [12]
  tcp:
    # 'idle_timeout' on TCPProxy [2]
    idleTimeout: 1h
  http:
    # 'request_timeout' on HttpConnectionManager [4]
    requestTimeout: 10s
    # 'request_headers_timeout' on HttpConnectionManager [11]
    requestHeadersTimeout: 5s
    # 'idle_timeout' on HttpConnectionManager [3]
    idleTimeout: 1h
  grpc:
    # 'stream_idle_timeout' on HttpConnectionManager [6]
    streamIdleTimeout: 5s
    # 'max_stream_duration' on HttpConnectionManager [5]
    maxStreamDuration: 30m
``` 

This policy aims to protect your service from the long living requests, if the service is not designed to support them.

### Uncovered timeouts

There are 3 uncovered timeouts - [8], [9], [10].

[8] and [10] allows to override existing timeouts on per-route basis, probably it makes sense to implement them in L7 
TrafficRoutes.

[9] - probably makes sense to implement in the scope of Retry policy. 

