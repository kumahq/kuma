licenses(["notice"])  # Apache 2

load(
    "@envoy//bazel:envoy_build_system.bzl",
    "envoy_cc_binary",
    "envoy_cc_test",
)

envoy_cc_binary(
    name = "konvoy",
    repository = "@envoy",
    deps = [
        "//source/extensions/filters/network/konvoy:konvoy_config",
        "//source/extensions/filters/http/konvoy:konvoy_config",
        "@envoy//source/exe:envoy_main_entry_lib",
    ],
)

sh_test(
    name = "konvoy_binary_test",
    srcs = ["konvoy_binary_test.sh"],
    data = [":konvoy"],
)
