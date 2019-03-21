# Konvoy filter

`Envoy` filter that pipes requests through a side car process over gRPC-based protocol.

Effectively, `Konvoy filter` enables extending `Envoy` in a programming language 
of your choice.   

## Building

To build `Konvoy` (`Envoy` + `Konvoy filter`) static binary:

1. `git submodule update --init`
2. `bazel build //:konvoy`

If you're new to `Bazel`, see [Developer Guide](DEVELOPER.md) on how to set up 
local dev environment or, alternatively,
how to use a `Docker` image that comes with all the required tools pre-installed.

## Running

To run `Konvoy` with a demo configuration:

1. Start [Konvoy demo gRPC server][konvoy-grpc-demo-java]
   * `${KONVOY_GRPC_DEMO_JAVA_HOME}/build/install/konvoy-grpc-demo-java/bin/konvoy-demo-server`
2. `bazel run -- //:konvoy -c $(pwd)/configs/konvoy.yaml `
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
```

## Configuring

### Konvoy http filter

To configure `Konvoy http filter`:
1. Add `Konvoy http filter` to the `filter chain` for a particular HTTP route configuration
2. Add a `cluster` for `Http Konvoy (gRPC Service)` (typically deployed as a side car process) 

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
- name: envoy.router
  config: {}

...

clusters:
#
# Konvoy gRPC Service deployed as a side car process
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

`bazel test //test/extensions/filters/http/konvoy:konvoy_integration_test`

## Developing

See [Developer Guide](DEVELOPER.md) for further details.

[konvoy-grpc-demo-java]: https://github.com/Kong/konvoy-grpc-demo-java
