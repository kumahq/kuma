FROM gcr.io/distroless/base-nossl-debian11:debug-nonroot@sha256:3e155791fd07828864f610e509acd34d67563dcdb69c7c641d2b2f737e5c9223

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
