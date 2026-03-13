FROM gcr.io/distroless/base-nossl-debian12:debug-nonroot@sha256:6cca41b265fde52139389ed3002e8a35de06c48c4e584d218617b0374006bc2f

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
