FROM gcr.io/distroless/static-debian12:debug-nonroot@sha256:8c28702f8a20280cd84526f1abc50c6a91f933e5c3bf792e3e47fd1263146ed7

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
