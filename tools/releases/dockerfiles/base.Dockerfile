FROM gcr.io/distroless/base-nossl-debian12:debug-nonroot@sha256:393a39691519a85382db0f75cedffc6f0208911484d138b2afc46c7d218e46af

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
