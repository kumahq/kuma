# Developer documentation

## Getting Envoy's code

* this repository depends on `Envoy` as a submodule
* to check out `Envoy`'s code, run 
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

## Including/excluding Envoy extensions from Konvoy

At the moment, `Konvoy`'s build configuration is optimized for faster development cycle.

In particular, almost all `Envoy` extensions (such as, `filters`) 
have been excluded from the build.

If you need to bring some of those extensions back, edit [envoy_build_config/extensions_build_config.bzl].

See `Envoy`'s [Disabling extensions][disabling-extensions] guide for further details.

## Testing Konvoy

To run the `Konvoy` integration tests, execute:

```bash
$ bazel test //test/extensions/filters/http/konvoy:konvoy_integration_test
```

[local_docker_build]: https://www.envoyproxy.io/docs/envoy/latest/install/sandboxes/local_docker_build
[quick-start-bazel-build-for-developers]: https://github.com/envoyproxy/envoy/blob/master/bazel/README.md#quick-start-bazel-build-for-developers
[disabling-extensions]: https://github.com/envoyproxy/envoy/blob/master/bazel/README.md#disabling-extensions
