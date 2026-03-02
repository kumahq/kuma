# use only when root is really needed
FROM gcr.io/distroless/base-nossl-debian12:debug@sha256:eacaa3a43627a7936de63d04028b034f36f21b930cc6f50c4650d63f4672fed6

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
