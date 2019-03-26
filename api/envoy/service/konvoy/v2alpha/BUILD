load("@envoy_api//bazel:api_build_system.bzl", "api_proto_library_internal")

licenses(["notice"])  # Apache 2

api_proto_library_internal(
    name = "http_konvoy_service",
    srcs = [
        "http_konvoy_service.proto",
    ],
    has_services = 1,
    require_py = 0,
    visibility = ["//visibility:public"],
    deps = [
        "@envoy_api//envoy/api/v2/core:base_export",
    ],
)

api_proto_library_internal(
    name = "network_konvoy_service",
    srcs = [
        "network_konvoy_service.proto",
    ],
    has_services = 1,
    require_py = 0,
    visibility = ["//visibility:public"],
    deps = [
        "@envoy_api//envoy/api/v2/core:base_export",
    ],
)
