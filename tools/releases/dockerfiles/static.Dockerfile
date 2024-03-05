FROM gcr.io/distroless/static-debian11:debug-nonroot@sha256:0758e75e315ea514023f24a965affd2fb5ab330638cafe564d4b99724cac03eb

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
