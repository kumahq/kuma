FROM gcr.io/distroless/static-debian12:debug-nonroot@sha256:eb3506377f0d8b0e173bbe61f1f7f13066ee6a7f8a7fd637c4a7e13a2799d092

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
