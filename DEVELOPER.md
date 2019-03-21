# Developer documentation

## Repository structure

* This repository contains sources of `Konvoy filter`
* The [Envoy repository](https://github.com/envoyproxy/envoy/) is provided as a submodule

## Bazel configuration

* `Bazel` is a build tool used by `Envoy`
* Since `Konvoy filter` is developed against `Envoy`'s C++ APIs
  and, in the end, is statically linked into a single binary with `Envoy`, 
  we also use `Bazel` and utilize `Envoy`'s `Bazel` configurations "as is"        
* To reuse `Envoy`'s source code already available as a Git submodule,  
  the [`WORKSPACE`](WORKSPACE) file maps the `@envoy` Bazel repository 
  to the local path of that submodule (instead of pointing to GitHub or to a tarball)
* [`BUILD`](BUILD) file introduces a new `Envoy static binary` target, named `konvoy`,
  that links together `Konvoy filter` and `@envoy//source/exe:envoy_main_lib`
* `Konvoy filter` registers itself  as a new filter 
  during the static initialization phase of the `Envoy` binary

## Fetching submodules

To fetch `Envoy`'s source code, run 
```bash
$ git submodule update --init
```    

## Using Docker for development tasks

If you don't want to (or unable to) setup native dev environment on your workstation,
you can always fallback to using `Docker`.

`Envoy` provides a `Docker` image with all required build tools pre-installed.

### Configuring Docker

* Building `Envoy` requires lots of CPU and memory resources
* If you're using `Docker for Mac`, remember to adjust default settings:
  * Open `Docker -> Preferences -> Advanced`
  * Increase amount of CPUs and memory available to `Docker Engine`  

### Launching build container 

To start a build container with source code mounted in, run 

```bash
$ ci/run_envoy_docker.sh
```

Now, you can run individual `Bazel` commands, such as

```bash
$ bazel info
```

```bash
$ bazel build //:konvoy
```   

See `Envoy`'s [Building an Envoy Docker image][local_docker_build] guide for further details. 

## Setting up native dev environment

1. On `MacOS`, install `Xcode` (`command line tools` are NOT enough)
2. Follow `Envoy`'s [Quick start Bazel build for developers][quick-start-bazel-build-for-developers]
   and install required dependencies

Once complete, you should be able to run individual `Bazel` commands directly on your workstation,
 e.g.
 
```bash
$ bazel info
```

```bash
$ bazel build //:konvoy
```   

## Building Konvoy

To build `Konvoy` (`Envoy` + `Konvoy filter`), run

```bash
$ bazel build //:konvoy
```

which will produce a binary at `bazel-bin/konvoy`.

Notice, however, that this binary is not suitable for benchmarking or production use,
since compiler options were optimized for faster build times rather than for performance
at runtime.

To build a version of `Konvoy` optimized for performance at runtime, run

```bash
$ bazel build -c opt //:konvoy
```

## Running Konvoy

To run `Konvoy` with a demo configuration:

To run `Konvoy` with a demo configuration:

1. Start demo gRPC Service (see [Konvoy demo gRPC server][konvoy-grpc-demo-java])
2. `bazel run -- //:konvoy -c $(pwd)/configs/konvoy.yaml `
3. Enable verbose logging by `Konvoy filter`
   * `curl -XPOST http://localhost:9901/logging?misc=trace`
4. Make arbitrary requests to `http://localhost:10000` (reverse proxied to `mockbin.org`), e.g.
   * `curl http://localhost:10000`
5. Observe communication between `Konvoy filter` and `Konvoy gRPC Service` in the logs

E.g.,

HTTP request
```
curl -XGET http://localhost:10000/request -d q=example
```

`Envoy` logs:
```
[2019-03-21 12:52:07.790][5422059][trace][misc] [source/extensions/filters/http/konvoy/konvoy.cc:47] konvoy-filter: forwarding request headers to Konvoy (side car):
':authority', 'localhost:10000'
':path', '/request'
':method', 'GET'
'user-agent', 'curl/7.54.0'
'accept', '*/*'
'content-length', '9'
'content-type', 'application/x-www-form-urlencoded'
'x-forwarded-proto', 'http'
'x-request-id', 'e808f4f7-568f-447f-816b-61bb63018b82'

[2019-03-21 12:52:07.790][5422059][trace][misc] [source/extensions/filters/http/konvoy/konvoy.cc:74] konvoy-filter: forwarding request body to Konvoy (side car):
9 bytes, end_stream=false, buffer_size=0
[2019-03-21 12:52:07.790][5422059][trace][misc] [source/extensions/filters/http/konvoy/konvoy.cc:74] konvoy-filter: forwarding request body to Konvoy (side car):
0 bytes, end_stream=true, buffer_size=0
[2019-03-21 12:52:07.790][5422059][trace][misc] [source/extensions/filters/http/konvoy/konvoy.cc:103] konvoy-filter: forwarding is finished
[2019-03-21 12:52:07.795][5422059][trace][misc] [source/extensions/filters/http/konvoy/konvoy.cc:107] konvoy-filter: received message from Konvoy (side car):
1
[2019-03-21 12:52:07.797][5422059][trace][misc] [source/extensions/filters/http/konvoy/konvoy.cc:107] konvoy-filter: received message from Konvoy (side car):
2
[2019-03-21 12:52:07.800][5422059][trace][misc] [source/extensions/filters/http/konvoy/konvoy.cc:107] konvoy-filter: received message from Konvoy (side car):
3
[2019-03-21 12:52:07.800][5422059][trace][misc] [source/extensions/filters/http/konvoy/konvoy.cc:168] konvoy-filter: received close signal from Konvoy (side car):
status = 0, message = 
``` 

Demo gRPC server logs:
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
      value: "/request"
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
      value: "09fe9fb0-579f-4785-b38c-d600087bc760"
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

## Testing Konvoy

To run the `Konvoy` integration tests, execute:

```bash
$ bazel test //test/extensions/filters/http/konvoy:konvoy_integration_test
```

## Making changes to the source code

Here is an overview of `Konvoy filter` source code: 

* [`konvoy.h`](source/extensions/filters/http/konvoy/konvoy.h) and 
  [`konvoy.cc`](source/extensions/filters/http/konvoy/konvoy.cc) implement
  the [`Envoy::Http::StreamDecoderFilter`][StreamDecoderFilter] interface;
  they're responsible for handling http headers, data, and trailers of received requests
* [`config.h`](source/extensions/filters/http/konvoy/config.h) and
  [`config.cc`](source/extensions/filters/http/konvoy/config.cc) implement 
  the `Envoy::Server::Configuration::NamedHttpFilterConfigFactory` interface;
  they enable the `Envoy` binary to find `Konvoy filter`
* all the above classes are linked to `Envoy` binary through the [`BUILD`][BUILD] file
* [`konvoy.proto`](api/envoy/config/filter/http/konvoy/v2alpha/konvoy.proto)
  is a `Protobuf` definition of the `Konvoy filter` configuration 
* [`http_konvoy_service.proto`](api/envoy/service/konvoy/v2alpha/http_konvoy_service.proto)
  is a `Protobuf` definition of the `Konvoy gRPC Service` implemented by a side car process
* [`extensions_build_config.bzl`](envoy_build_config/extensions_build_config.bzl)
  is a `Bazel` configuration that includes/excludes `Envoy` extensions (such as, `filters`) 
  from the build and resulting binary
* [`konvoy.yaml`](configs/konvoy.yaml) is a sample `Envoy` configuration
  that utilizes `Konvoy filter`    
 
## Including/excluding Envoy extensions from Konvoy

At the moment, `Konvoy`'s build configuration is optimized for faster development cycle.

In particular, almost all `Envoy` extensions (such as, `filters`) 
have been excluded from the build.

If you need to bring some of those extensions back, edit [envoy_build_config/extensions_build_config.bzl](envoy_build_config/extensions_build_config.bzl).

See `Envoy`'s [Disabling extensions][disabling-extensions] guide for further details.

[local_docker_build]: https://www.envoyproxy.io/docs/envoy/latest/install/sandboxes/local_docker_build
[quick-start-bazel-build-for-developers]: https://github.com/envoyproxy/envoy/blob/master/bazel/README.md#quick-start-bazel-build-for-developers
[disabling-extensions]: https://github.com/envoyproxy/envoy/blob/master/bazel/README.md#disabling-extensions
[konvoy-grpc-demo-java]: https://github.com/Kong/konvoy-grpc-demo-java
[StreamDecoderFilter]: https://github.com/envoyproxy/envoy/blob/b2610c84aeb1f75c804d67effcb40592d790e0f1/include/envoy/http/filter.h#L300
[StreamEncoderFilter]: https://github.com/envoyproxy/envoy/blob/b2610c84aeb1f75c804d67effcb40592d790e0f1/include/envoy/http/filter.h#L413
[StreamFilter]: https://github.com/envoyproxy/envoy/blob/b2610c84aeb1f75c804d67effcb40592d790e0f1/include/envoy/http/filter.h#L462
[BUILD]: BUILD
