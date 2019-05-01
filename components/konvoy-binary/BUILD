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
        "@envoy//source/exe:envoy_main_entry_lib",

        # Konvoy Filter
        "@konvoy_filter//source/extensions/filters/network/konvoy:konvoy_config",
        "@konvoy_filter//source/extensions/filters/http/konvoy:konvoy_config",
    ],
)

sh_test(
    name = "konvoy_binary_test",
    srcs = ["konvoy_binary_test.sh"],
    data = [":konvoy"],
)
