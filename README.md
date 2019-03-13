# Konvoy filter

Envoy filter that delegates request processing to a side car process
acting as a gRPC service.

Effectively, Konvoy filter enables extending Envoy in a programming language 
of your choice.   

## Building

To build the Envoy static binary:

1. `git submodule update --init`
2. `bazel build //:envoy`

## Running

To run the `Konvoy` with a demo configuration:

1. Start demo `Konvoy` gRPC Service (see [Konvoy demo gRPC server][konvoy-grpc-demo-java])
2. `bazel run -- //:envoy -c $(pwd)/configs/konvoy.yaml `
3. Make arbitrary requests to `http://localhost:10000` (reverse proxied to `www.google.com`)
and observe communication between `Konvoy Filter` and `Konvoy gRPC Service` in the logs

E.g.,

HTTP request
```
curl -XGET http://localhost:10000/search -d q=example
```

Envoy logs
```
[2019-03-13 14:32:10.962][2411720][info][filter] [source/extensions/filters/http/konvoy/konvoy.cc:37] konvoy-filter: forwarding request headers to Konvoy (side car):
':authority', 'localhost:10000'
':path', '/search'
':method', 'GET'
'user-agent', 'curl/7.54.0'
'accept', '*/*'
'content-length', '9'
'content-type', 'application/x-www-form-urlencoded'
'x-forwarded-proto', 'http'
'x-request-id', 'e511a1db-3905-4692-a0a0-29f5b72e6836'

[2019-03-13 14:32:10.962][2411720][info][filter] [source/extensions/filters/http/konvoy/konvoy.cc:56] konvoy-filter: forwarding request body to Konvoy (side car):
9 bytes, end_stream=false
[2019-03-13 14:32:10.962][2411720][info][filter] [source/extensions/filters/http/konvoy/konvoy.cc:56] konvoy-filter: forwarding request body to Konvoy (side car):
0 bytes, end_stream=true
[2019-03-13 14:32:10.962][2411720][info][filter] [source/extensions/filters/http/konvoy/konvoy.cc:84] konvoy-filter: forwarding is finished
[2019-03-13 14:32:10.964][2411720][info][filter] [source/extensions/filters/http/konvoy/konvoy.cc:88] konvoy-filter: received message from Konvoy (side car):
1
[2019-03-13 14:32:10.965][2411720][info][filter] [source/extensions/filters/http/konvoy/konvoy.cc:88] konvoy-filter: received message from Konvoy (side car):
2
[2019-03-13 14:32:10.966][2411720][info][filter] [source/extensions/filters/http/konvoy/konvoy.cc:88] konvoy-filter: received message from Konvoy (side car):
3
[2019-03-13 14:32:10.967][2411720][info][filter] [source/extensions/filters/http/konvoy/konvoy.cc:94] konvoy-filter: received close signal from Konvoy (side car):
status = 0, message = 
``` 

Demo gRPC server logs
```
Mar 13, 2019 2:32:10 PM com.konghq.konvoy.demo.KonvoyServer$KonvoyService$1 onNext
INFO: onNext: request_headers {
  headers {
    headers {
      key: ":authority"
      value: "localhost:10000"
    }
    headers {
      key: ":path"
      value: "/search"
    }
    headers {
      key: ":method"
      value: "GET"
    }
    headers {
      key: "user-agent"
      value: "curl/7.54.0"
    }
    headers {
      key: "accept"
      value: "*/*"
    }
    headers {
      key: "content-length"
      value: "9"
    }
    headers {
      key: "content-type"
      value: "application/x-www-form-urlencoded"
    }
    headers {
      key: "x-forwarded-proto"
      value: "http"
    }
    headers {
      key: "x-request-id"
      value: "e511a1db-3905-4692-a0a0-29f5b72e6836"
    }
  }
}

Mar 13, 2019 2:32:10 PM com.konghq.konvoy.demo.KonvoyServer$KonvoyService$1 onNext
INFO: onNext: request_body_chunk {
  bytes: "q=example"
}

Mar 13, 2019 2:32:10 PM com.konghq.konvoy.demo.KonvoyServer$KonvoyService$1 onNext
INFO: onNext: request_trailers {
}

Mar 13, 2019 2:32:10 PM com.konghq.konvoy.demo.KonvoyServer$KonvoyService$1 onCompleted
INFO: onCompleted
```

## Testing

To run the `Konvoy` integration test:

`bazel test //test/extensions/filters/http/konvoy:konvoy_integration_test`

## How it works

The [Envoy repository](https://github.com/envoyproxy/envoy/) is provided as a submodule.
The [`WORKSPACE`](WORKSPACE) file maps the `@envoy` repository to this local path.

The [`BUILD`](BUILD) file introduces a new Envoy static binary target, `envoy`,
that links together the new filter and `@envoy//source/exe:envoy_main_lib`. The
`envoy` filter registers itself during the static initialization phase of the
Envoy binary as a new filter.

## How to write and use Konvoy HTTP filter

- The main task is to write a class that implements the interface
 [`Envoy::Http::StreamDecoderFilter`][StreamDecoderFilter] as in
 [`konvoy.h`](source/extensions/filters/http/konvoy/konvoy.h) and [`konvoy.cc`](source/extensions/filters/http/konvoy/konvoy.cc),
 which contains functions that handle http headers, data, and trailers.
 To write encoder filters or decoder/encoder filters
 you need to implement 
 [`Envoy::Http::StreamEncoderFilter`][StreamEncoderFilter] or
 [`Envoy::Http::StreamFilter`][StreamFilter] instead.
- You also need a class that implements 
 `Envoy::Server::Configuration::NamedHttpFilterConfigFactory`
 to enable the Envoy binary to find your filter,
 as in [`konvoy_config.h`](source/extensions/filters/http/konvoy/config.h).
 It should be linked to the Envoy binary by modifying [`BUILD`][BUILD] file.
- Finally, you need to modify the Envoy config file to add `konvoy` filter to the
 filter chain for a particular HTTP route configuration. For instance, if you
 wanted to change [the front-proxy example][front-envoy.yaml] to chain our
 `konvoy` filter, you'd need to modify its config to look like

```yaml
http_filters:
- name: konvoy          # before envoy.router because order matters!
  config:
    grpc_service:
      envoy_grpc:
        cluster_name: konvoy_side_car
- name: envoy.router
  config: {}
...
clusters:
#
# Konvoy side car
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
```
 
[StreamDecoderFilter]: https://github.com/envoyproxy/envoy/blob/b2610c84aeb1f75c804d67effcb40592d790e0f1/include/envoy/http/filter.h#L300
[StreamEncoderFilter]: https://github.com/envoyproxy/envoy/blob/b2610c84aeb1f75c804d67effcb40592d790e0f1/include/envoy/http/filter.h#L413
[StreamFilter]: https://github.com/envoyproxy/envoy/blob/b2610c84aeb1f75c804d67effcb40592d790e0f1/include/envoy/http/filter.h#L462
[BUILD]: BUILD
[front-envoy.yaml]: https://github.com/envoyproxy/envoy/blob/b2610c84aeb1f75c804d67effcb40592d790e0f1/examples/front-proxy/front-envoy.yaml#L28
[konvoy-grpc-demo-java]: https://github.com/Kong/konvoy-grpc-demo-java
