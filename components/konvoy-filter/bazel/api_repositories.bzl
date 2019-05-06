def _default_konvoy_filter_api_impl(ctx):
    ctx.file("WORKSPACE", "")
    ctx.file("BUILD.bazel", "")
    api_dirs = [
        "envoy",
    ]
    for d in api_dirs:
        ctx.symlink(ctx.path(ctx.attr.api).dirname.get_child(d), d)

_default_konvoy_filter_api = repository_rule(
    implementation = _default_konvoy_filter_api_impl,
    attrs = {
        "api": attr.label(default = "@konvoy_filter//api:BUILD"),
    },
)

def konvoy_filter_api_dependencies():
    # Treat the data plane API as an external repo, this simplifies exporting the API to
    # https://github.com/envoyproxy/data-plane-api.
    if "konvoy_filter_api" not in native.existing_rules().keys():
        _default_konvoy_filter_api(name = "konvoy_filter_api")
