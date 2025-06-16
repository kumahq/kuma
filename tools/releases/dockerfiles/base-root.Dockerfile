# use only when root is really needed
FROM gcr.io/distroless/base-nossl-debian12:debug@sha256:479e2ee4e0128106370963843337c27c70cc3451ea17a1c06f013665c2fe91e7

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
