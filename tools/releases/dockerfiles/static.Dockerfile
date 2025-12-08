FROM gcr.io/distroless/static-debian12:debug-nonroot@sha256:53ced3252beed1467debcf96377c78ef7bc824987cab7ed8b99c92a746547886

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
