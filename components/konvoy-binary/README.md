# Konvoy binary

Build configuration to assemble Envoy binary that includes 3rd party extensions,
such as Konvoy, Istio, etc.

## Default Konvoy binary

By default, `Konvoy` binary includes:

* [Envoy] (including all upstream extensions)
* [Konvoy] extensions
* [Istio] extensions

## Custom Konvoy binary

### How To exclude an extension

In order to exclude extensions from `Konvoy binary`, you need to override
`konvoy_build_config` Bazel repository.

The overall approach is similar to [disabling extensions][disabling-extensions] in upstream Envoy. 

Essentially, you need to copy and edit 
[bazel/konvoy_build_config.bzl](bazel/konvoy_build_config.bzl) file.

E.g., the following configuration will exclude [Istio] extensions from `Konvoy binary`:

```python
BUILD_CONFIG = dict(
    konvoy_filter = dict(
        #
        # Konvoy's extensions to Envoy
        #
        EXTENSIONS = {
            #
            # HTTP filters
            #

            "envoy.filters.http.konvoy":                 "//source/extensions/filters/http/konvoy:konvoy_config",

            #
            # Network filters
            #

            "envoy.filters.network.konvoy":              "//source/extensions/filters/network/konvoy:konvoy_config",
        }
    ),
    istio_proxy = dict(
        #
        # Istio's extensions to Envoy
        #
        EXTENSIONS = {
            #
            # Do NOT include Istio extensions into Konvoy binary
            #
        }
    ),
)
```   

### How to include a 3rd party Envoy extension

In order to include a 3rd party Envoy extension into `Konvoy binary`, 
you need to update Bazel configuration as follows:

1. Add a link to a 3rd party Envoy extension into 
[bazel/repository_locations.bzl](bazel/repository_locations.bzl)
2. Update definition of `konvoy_dependencies()` in [bazel/repositories.bzl](bazel/repositories.bzl)
3. Update [WORKSPACE](WORKSPACE)
4. Update [bazel/konvoy_build_config.bzl](bazel/konvoy_build_config.bzl)

E.g., take a look at a configuration for [Istio] extensions:

1. [bazel/repository_locations.bzl](bazel/repository_locations.bzl)
    ```python
    ISTIO_PROXY_GIT_SHA = "23e050c3300b9769987fc8d5aa7fd0925c630055"
    ISTIO_PROXY_SHA256 = "2b9671250646ac37a3b291fdf026dc0d85c34c2482e2ddb408809b53bc6ee7b6"
    
    istio_proxy = dict(
        sha256 = ISTIO_PROXY_SHA256,
        strip_prefix = "proxy-" + ISTIO_PROXY_GIT_SHA,
        urls = ["https://github.com/istio/proxy/archive/" + ISTIO_PROXY_GIT_SHA + ".tar.gz"],
    ),
    ```
2. [bazel/repositories.bzl](bazel/repositories.bzl)    
    ```python
    def konvoy_dependencies():
       ...
       _istio_proxy()
 
    def _istio_proxy():
        http_archive(
            name = "istio_proxy",
            **REPOSITORY_LOCATIONS["istio_proxy"]
        )
    ```
3. [WORKSPACE](WORKSPACE)
    ```python
    #
    # Istio Proxy dependencies
    #
    
    load("@istio_proxy//:repositories.bzl", "googletest_repositories", "mixerapi_dependencies")
    googletest_repositories()
    mixerapi_dependencies()
    ```
4. [bazel/konvoy_build_config.bzl](bazel/konvoy_build_config.bzl)
    ```python
    istio_proxy = dict(
        #
        # Istio's extensions to Envoy
        #
        EXTENSIONS = {
            #
            # HTTP filters
            #

            "istio_authn":                               "//src/envoy/http/authn:filter_lib",
            "jwt-auth":                                  "//src/envoy/http/jwt_auth:http_filter_factory",
            "mixer/http":                                "//src/envoy/http/mixer:filter_lib",

            #
            # Network filters
            #

            "forward_downstream_sni":                    "//src/envoy/tcp/forward_downstream_sni:config_lib",
            "mixer/tcp":                                 "//src/envoy/tcp/mixer:filter_lib",
            "sni_verifier":                              "//src/envoy/tcp/sni_verifier:config_lib",
            "envoy.filters.network.tcp_cluster_rewrite": "//src/envoy/tcp/tcp_cluster_rewrite:config_lib",
        }
    ),    
    ```

[Envoy]: https://github.com/envoyproxy/envoy
[Konvoy]: ../konvoy-filter
[Istio]: https://github.com/istio/proxy
[disabling-extensions]: https://github.com/envoyproxy/envoy/blob/master/bazel/README.md#disabling-extensions
