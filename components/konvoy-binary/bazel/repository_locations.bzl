ENVOY_GIT_SHA = "b8bcb11f5fb8c5ae19535444a9de10a207d9cec6"
ENVOY_SHA256 = "ee0b9ef5707604b41d6439f4a3b8ea05cfe728705d2d24f75a43e3ea74845ff7"

ISTIO_PROXY_GIT_SHA = "23e050c3300b9769987fc8d5aa7fd0925c630055"
ISTIO_PROXY_SHA256 = "2b9671250646ac37a3b291fdf026dc0d85c34c2482e2ddb408809b53bc6ee7b6"

REPOSITORY_LOCATIONS = dict(
    envoy = dict(
        sha256 = ENVOY_SHA256,
        strip_prefix = "envoy-" + ENVOY_GIT_SHA,
        urls = ["https://github.com/envoyproxy/envoy/archive/" + ENVOY_GIT_SHA + ".tar.gz"],
    ),
    konvoy_filter = dict(
        path = "../konvoy-filter",
    ),
    istio_proxy = dict(
        sha256 = ISTIO_PROXY_SHA256,
        strip_prefix = "proxy-" + ISTIO_PROXY_GIT_SHA,
        urls = ["https://github.com/istio/proxy/archive/" + ISTIO_PROXY_GIT_SHA + ".tar.gz"],
    ),
)
