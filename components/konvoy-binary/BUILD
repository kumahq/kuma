licenses(["notice"])  # Apache 2

load(
    "@envoy//bazel:envoy_build_system.bzl",
    "envoy_cc_binary",
    "envoy_cc_test",
)

load(
    "//:extensions.bzl",
    "envoy_extensions",
)

envoy_cc_binary(
    name = "konvoy",
    repository = "@envoy",
    deps = [ "@envoy//source/exe:envoy_main_entry_lib" ] + envoy_extensions()
)

sh_test(
    name = "konvoy_binary_test",
    srcs = ["konvoy_binary_test.sh"],
    data = [":konvoy"],
)
