FROM gcr.io/distroless/base-nossl-debian11:debug-nonroot@sha256:083ed82263bc4872beca0f09a85041d8aa9d2f0214692a79483c63a5c779b8ed

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
