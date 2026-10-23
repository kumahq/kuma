FROM gcr.io/distroless/static-debian13:debug-nonroot@sha256:06bfc1ed6f52f8efecd4e68e5f146fc06891b1c40f2d1654f1518392ba89815a

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
