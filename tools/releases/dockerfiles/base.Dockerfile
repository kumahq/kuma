FROM gcr.io/distroless/base-nossl-debian11:debug-nonroot@sha256:934b713496a9ed100550aaa58636270c4d69c27040e44f2aed1fa39594c45eba

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
