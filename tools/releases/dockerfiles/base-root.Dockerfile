# use only when root is really needed
FROM gcr.io/distroless/base-nossl-debian11:debug@sha256:1dcd82e95d13f1fbe8977f5b798d3912453f885d54f96380c7242ab2bd3c87b7

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
