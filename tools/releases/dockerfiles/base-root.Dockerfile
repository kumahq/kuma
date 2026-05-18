# use only when root is really needed
FROM gcr.io/distroless/base-nossl-debian12:debug@sha256:35a3865dc6080a6a29c0fc4022c9c780ce619b923edb0bea580f2e7d7852343b

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
