load("@konvoy_build_config//:konvoy_build_config.bzl", "BUILD_CONFIG")

# Return list of extensions to be compiled into Konvoy.
def envoy_extensions(blacklist = dict()):
    extensions = []

    for dependency, config in BUILD_CONFIG.items():
        for name, path in config["EXTENSIONS"].items():
            if not name in blacklist.values():
                extensions.append("@" + dependency + path)

    return extensions
