FROM gcr.io/distroless/base-nossl-debian13:debug-nonroot@sha256:37b8b2d30e81bf028eae7c1e8c65d775d46b2a21a2b76cde6bf8bf59ef0e3289

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
