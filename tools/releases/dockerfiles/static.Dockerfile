FROM gcr.io/distroless/static-debian11:debug-nonroot@sha256:312a533b1f5584141a7d212ddcc1d079259a84ef68a1a5b0f522017093e3afda

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
