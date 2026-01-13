FROM gcr.io/distroless/base-nossl-debian12:debug-nonroot@sha256:b9a1ed9775ca5c3b84f63be38d52412a5a67f870bb814ae34ea609b9bfb696b5

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
