# use only when root is really needed
FROM gcr.io/distroless/base-nossl-debian12:debug@sha256:9e9eb25f82ef4c496b1f3f16b73d6d114764ae39d0fce3b341a4af0ee0166fbe

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
