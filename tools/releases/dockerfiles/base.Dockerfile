FROM gcr.io/distroless/base-nossl-debian11:debug-nonroot@sha256:33768a8ee3915687881d0a41580901373d8f924416b046d524ce0a5bee8fdf18

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
