load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
load(":repository_locations.bzl", "REPOSITORY_LOCATIONS")

def _default_konvoy_build_config_impl(ctx):
    ctx.file("WORKSPACE", "")
    ctx.file("BUILD.bazel", "")
    ctx.symlink(ctx.attr.config, "konvoy_build_config.bzl")

_default_konvoy_build_config = repository_rule(
    implementation = _default_konvoy_build_config_impl,
    attrs = {
        "config": attr.label(default = "@konvoy//bazel:konvoy_build_config.bzl"),
    },
)

def konvoy_dependencies():
    # Treat Konvoy's overall build config as an external repo, so projects that
    # build Konvoy as a subcomponent can easily override the config.
    if "konvoy_build_config" not in native.existing_rules().keys():
        _default_konvoy_build_config(name = "konvoy_build_config")

    _envoy()
    _konvoy_filter()
    _istio_proxy()

def _envoy():
    http_archive(
        name = "envoy",
        **REPOSITORY_LOCATIONS["envoy"]
    )

def _konvoy_filter():
    native.local_repository(
        name = "konvoy_filter",
        **REPOSITORY_LOCATIONS["konvoy_filter"]
    )

def _istio_proxy():
    http_archive(
        name = "istio_proxy",
        **REPOSITORY_LOCATIONS["istio_proxy"]
    )
