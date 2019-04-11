licenses(["notice"])  # Apache 2

load(
    "@envoy//bazel:envoy_build_system.bzl",
    "envoy_cc_test",
    "envoy_package",
)

envoy_package()

envoy_cc_test(
    name = "config_test",
    srcs = ["config_test.cc"],
    repository = "@envoy",
    deps = [
        "//source/extensions/filters/network/konvoy:konvoy_config",
        "@envoy//test/mocks/server:server_mocks",
    ],
)

envoy_cc_test(
    name = "konvoy_test",
    srcs = ["konvoy_test.cc"],
    repository = "@envoy",
    deps = [
        "//source/extensions/filters/network/konvoy:konvoy",
        "@envoy//test/mocks/grpc:grpc_mocks",
        "@envoy//test/mocks/network:network_mocks",
        "@envoy//test/test_common:simulated_time_system_lib",
    ],
)
