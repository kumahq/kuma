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
)
