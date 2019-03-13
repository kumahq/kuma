licenses(["notice"])  # Apache 2

load(
    "@envoy//bazel:envoy_build_system.bzl",
    "envoy_cc_test",
    "envoy_package",
)

envoy_package()

envoy_cc_test(
    name = "konvoy_integration_test",
    srcs = ["konvoy_integration_test.cc"],
    repository = "@envoy",
    deps = [
        "//source/extensions/filters/http/konvoy:konvoy_config",
        "@envoy//test/integration:http_integration_lib",
    ],
)
