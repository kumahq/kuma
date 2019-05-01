# Konvoy filter

`Envoy` filter that pipes requests through a side car process over gRPC-based protocol.

Effectively, `Konvoy filter` enables extending `Envoy` in a programming language 
of your choice.   

## Overview

Technically speaking, `Konvoy` consists of 2 `Envoy` filters and 2 gRPC Services:
1. `Konvoy http filter` + [`Http Konvoy (gRPC Service)`](api/envoy/service/konvoy/v2alpha/http_konvoy_service.proto) : for piping HTTP requests
2. `Konvoy network filter` + [`Network Konvoy (gRPC Service)`](api/envoy/service/konvoy/v2alpha/network_konvoy_service.proto) : for piping L4 payload data

`Konvoy filters` integrate into `Envoy` machinery and pipe L4 or L7 request data though a corresponding
gRPC Service.

As a user of `Konvoy`, you can provide custom implementations of [`Http Konvoy`](api/envoy/service/konvoy/v2alpha/http_konvoy_service.proto)
and [`Network Konvoy`](api/envoy/service/konvoy/v2alpha/network_konvoy_service.proto)
gRPC services to tailor `Envoy` for your personal needs.

## Building

To build `Konvoy` (`Envoy` + `Konvoy filter`) static binary:

1. `git submodule update --init`
2. `make build/binary`

If you're new to `Bazel`, see [Developer Guide](DEVELOPER.md) on how to set up 
local dev environment or, alternatively,
how to use a `Docker` image that comes with all the required tools pre-installed.

## Running

To run `Konvoy` with a demo configuration:

1. Start [Konvoy demo gRPC server][konvoy-grpc-demo-java]
   * `${KONVOY_GRPC_DEMO_JAVA_HOME}/build/install/konvoy-grpc-demo-java/bin/konvoy-demo-server`
2. Start `Konvoy` (`Envoy` + `Konvoy filter`) 
   * `make run/demo`
3. Enable verbose logging in `Konvoy filter`
   * `curl -XPOST http://localhost:9901/logging?misc=trace`
4. Make arbitrary requests to `http://localhost:10000` (reverse proxied to `mockbin.org`), e.g.
   * `curl http://localhost:10000`
5. Observe communication between `Konvoy filter` and `Konvoy gRPC Service` in the logs

See [Developer Guide](DEVELOPER.md) for further details.

## Observing

To scrape metrics related to `Konvoy filter`:

`curl -s http://localhost:9901/stats |grep -e 'konvoy\.'`

```
konvoy.http.demo-grpc-server.request_total: 142167
konvoy.http.demo-grpc-server.request_total_stream_exchange_latency_ms: 79061
konvoy.http.demo-grpc-server.request_total_stream_latency_ms: 82015
konvoy.http.demo-grpc-server.request_total_stream_start_latency_ms: 193
konvoy.http.demo-grpc-server.request_stream_exchange_latency_ms: P0(0,0) P25(0,0) P50(0,0) P75(1.07191,1.07191) P90(2.07696,2.07696) P95(3.09033,3.09033) P99(8.05311,8.05311) P99.5(13.88,13.88) P99.9(81.5587,81.5587) P100(140,140)
konvoy.http.demo-grpc-server.request_stream_latency_ms: P0(0,0) P25(0,0) P50(0,0) P75(1.07469,1.07469) P90(2.08013,2.08013) P95(3.09371,3.09371) P99(8.05959,8.05959) P99.5(13.8883,13.8883) P99.9(81.6704,81.6704) P100(140,140)
konvoy.http.demo-grpc-server.request_stream_start_latency_ms: P0(0,0) P25(0,0) P50(0,0) P75(0,0) P90(0,0) P95(0,0) P99(0,0) P99.5(0,0) P99.9(1.03402,1.03402) P100(5.1,5.1)

konvoy.network.demo-grpc-server.cx_active: 1
konvoy.network.demo-grpc-server.cx_error: 0
konvoy.network.demo-grpc-server.cx_total: 3
konvoy.network.demo-grpc-server.cx_total_stream_exchange_latency_ms: 46125
konvoy.network.demo-grpc-server.cx_total_stream_latency_ms: 46137
konvoy.network.demo-grpc-server.cx_total_stream_start_latency_ms: 11
konvoy.network.demo-grpc-server.cx_stream_exchange_latency_ms: P0(nan,21000) P25(nan,21500) P50(nan,22000) P75(nan,24500) P90(nan,24800) P95(nan,24900) P99(nan,24980) P99.5(nan,24990) P99.9(nan,24998) P100(nan,25000)
konvoy.network.demo-grpc-server.cx_stream_latency_ms: P0(nan,21000) P25(nan,21500) P50(nan,22000) P75(nan,24500) P90(nan,24800) P95(nan,24900) P99(nan,24980) P99.5(nan,24990) P99.9(nan,24998) P100(nan,25000)
konvoy.network.demo-grpc-server.cx_stream_start_latency_ms: P0(nan,0) P25(nan,0) P50(nan,0) P75(nan,11.5) P90(nan,11.8) P95(nan,11.9) P99(nan,11.98) P99.5(nan,11.99) P99.9(nan,11.998) P100(nan,12)
```

## Configuring

### Konvoy http filter

`Konvoy http filter` must be used in the context of `envoy.http_connection_manager` 
and precede `envoy.router`.

A custom [`Http Konvoy (gRPC Service)`](api/envoy/service/konvoy/v2alpha/http_konvoy_service.proto)
can inspect/transform HTTP request message and let `envoy.router` to dispatch it upstream.

---

To configure `Konvoy http filter`:
1. Add `Konvoy http filter` to the `filter chain` for a particular HTTP route configuration
2. Add a `cluster` for [`Http Konvoy (gRPC Service)`](api/envoy/service/konvoy/v2alpha/http_konvoy_service.proto) (typically deployed as a side car process) 

E.g., here is an excerpt from the demo configuration: 

```yaml
...

http_filters:
#
# Konvoy filter to pipe HTTP request through a side car process 
#
- name: konvoy
  typed_config:
    "@type": type.googleapis.com/envoy.config.filter.http.konvoy.v2alpha.Konvoy
    stat_prefix: demo-grpc-server
    grpc_service:
      envoy_grpc:
        cluster_name: konvoy_side_car
    per_service_config:
      # Configuration specific to a custom "Http Konvoy" Service implementation,
      # i.e. "Demo Http Konvoy" Service.
      http_konvoy:
        # Configuration defined via a YAML/JSON file has to use `google.protobuf.Struct`
        # instead of a Proto type natively supported by a particular Konvoy Server.    
        "@type": type.googleapis.com/google.protobuf.Struct
        value:
          # HTTP request header to inject
          header_name: via-konvoy-service
          header_value: demo-http-konvoy    
- name: envoy.router
  config: {}

...

clusters:
#
# "Http Konvoy" and "Network Konvoy" gRPC Services deployed as a side car process
#
- name: konvoy_side_car
  connect_timeout: 0.25s
  type: STATIC
  dns_lookup_family: V4_ONLY
  hosts:
  - socket_address:
      address: 127.0.0.1
      port_value: 8980
  lb_policy: ROUND_ROBIN
  http2_protocol_options: {}
  
...
```

### Konvoy network filter

`Konvoy network filter` can be used:
* either as a terminal filter in the chain
* or as an intermediate filter followed by `envoy.tcp_proxy`

In the former case, a custom [`Network Konvoy (gRPC Service)`](api/envoy/service/konvoy/v2alpha/network_konvoy_service.proto)
is responsible for the entire processing and should only return back 
a stream of bytes to send back downstream.

In the latter case, a custom [`Network Konvoy (gRPC Service)`](api/envoy/service/konvoy/v2alpha/network_konvoy_service.proto)
can inspect/transform request stream and let `envoy.tcp_proxy` to dispatch it upstream.      

---

To configure `Konvoy network filter`:
1. Add `Konvoy network filter` to the `filter chain`
2. Add a `cluster` for [`Network Konvoy (gRPC Service)`](api/envoy/service/konvoy/v2alpha/network_konvoy_service.proto) (typically deployed as a side car process) 

E.g., here is an excerpt from the demo configuration: 

```yaml
- name: listener_1
  address:
    socket_address:
      protocol: TCP
      address: 0.0.0.0
      port_value: 10001
  filter_chains:
  - filters:
    #
    # Konvoy filter to pipe L4 payload data through a side car process
    #
    - name: konvoy
      typed_config:
        "@type": type.googleapis.com/envoy.config.filter.network.konvoy.v2alpha.Konvoy
        stat_prefix: demo-grpc-server
        grpc_service:
          envoy_grpc:
            cluster_name: konvoy_side_car
        per_service_config:
          # Configuration specific to a custom "Network Konvoy" Service implementation,
          # i.e. "Demo Network Konvoy" Service.      
          network_konvoy:
            # Configuration defined via a YAML/JSON file has to use `google.protobuf.Struct`
            # instead of a Proto type natively supported by a particular Konvoy Server.    
            "@type": type.googleapis.com/google.protobuf.Struct
            value:
              # Delay to apply
              fixed_delay: 2s
    - name: envoy.tcp_proxy
      typed_config:
        "@type": type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
        stat_prefix: post_konvoy_tcp_proxy
        cluster: service_mockbin

...

clusters:
#
# "Http Konvoy" and "Network Konvoy" gRPC Services deployed as a side car process
#
- name: konvoy_side_car
  connect_timeout: 0.25s
  type: STATIC
  dns_lookup_family: V4_ONLY
  hosts:
  - socket_address:
      address: 127.0.0.1
      port_value: 8980
  lb_policy: ROUND_ROBIN
  http2_protocol_options: {}
  
...
```

## Testing

To run `Konvoy` integration tests:

```bash
$ make run/tests
```

## Verifying Test Coverage

To verify test coverage:

```bash
$ make collect/coverage
```

To open coverage report in a browser:

```bash
open generated/coverage/coverage.html
```

## Developing

See [Developer Guide](DEVELOPER.md) for further details.

[konvoy-grpc-demo-java]: https://github.com/Kong/konvoy-grpc-demo-java
