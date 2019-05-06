licenses(["notice"])  # Apache 2

# Out-of-process request processing L7 HTTP filter

load(
    "@envoy//bazel:envoy_build_system.bzl",
    "envoy_cc_library",
    "envoy_package",
)

envoy_package()

envoy_cc_library(
    name = "konvoy",
    srcs = ["konvoy.cc"],
    hdrs = ["konvoy.h"],
    repository = "@envoy",
    deps = [
        "@envoy//include/envoy/grpc:async_client_interface",
        "@envoy//include/envoy/server:filter_config_interface",
        "@envoy//include/envoy/stats:stats_macros",
        "@envoy//source/common/buffer:buffer_lib",
        "@envoy//source/common/common:enum_to_int",
        "@envoy//source/common/http:header_map_lib",
        "@envoy//source/common/http:header_utility_lib",
        "@konvoy_filter_api//envoy/config/filter/http/konvoy/v2alpha:konvoy_cc",
        "@konvoy_filter_api//envoy/service/konvoy/v2alpha:http_konvoy_service_cc",
        ":konvoy_proto_utils_lib",
    ],
)

envoy_cc_library(
    name = "konvoy_config",
    srcs = ["config.cc"],
    hdrs = ["config.h"],
    repository = "@envoy",
    deps = [
        ":konvoy",
        "@envoy//include/envoy/registry",
        "@envoy//source/extensions/filters/http:well_known_names",
        "@envoy//source/extensions/filters/http/common:factory_base_lib",
    ],
)

envoy_cc_library(
    name = "konvoy_proto_utils_lib",
    srcs = ["proto_utils.cc"],
    hdrs = ["proto_utils.h"],
    repository = "@envoy",
    deps = [
        "@envoy//include/envoy/buffer:buffer_interface",
        "@envoy//include/envoy/http:header_map_interface",
        "@konvoy_filter_api//envoy/service/konvoy/v2alpha:http_konvoy_service_cc",
    ],
)
