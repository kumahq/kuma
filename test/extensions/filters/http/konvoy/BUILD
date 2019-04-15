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

envoy_cc_test(
    name = "konvoy_test",
    srcs = ["konvoy_test.cc"],
    repository = "@envoy",
    deps = [
        "//source/extensions/filters/http/konvoy:konvoy",
        "@envoy//source/common/protobuf:utility_lib",
        "@envoy//test/mocks/grpc:grpc_mocks",
        "@envoy//test/mocks/http:http_mocks",
        "@envoy//test/mocks/stats:stats_mocks",
        "@envoy//test/test_common:simulated_time_system_lib",
        "@envoy//test/test_common:utility_lib",
    ],
)

envoy_cc_test(
    name = "config_test",
    srcs = ["config_test.cc"],
    repository = "@envoy",
    deps = [
        "//source/extensions/filters/http/konvoy:konvoy_config",
        "@envoy//test/mocks/server:server_mocks",
    ],
)
