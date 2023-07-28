# use only when root is really needed
FROM gcr.io/distroless/base-nossl-debian11:debug@sha256:6579e1f772c215415dc5921eeebd80c5178957657139b5b5bddb63e572a1c588

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
