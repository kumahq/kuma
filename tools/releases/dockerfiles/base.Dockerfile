FROM gcr.io/distroless/base-nossl-debian12:debug-nonroot@sha256:8c89c12a1900f0cc2b00e47244bb0c50add49081475ebf5ad3170f4358874e00

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
