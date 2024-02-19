FROM gcr.io/distroless/static-debian11:debug-nonroot@sha256:3d214b77f39d95f0d5994ecaa938f05abf5a1481084c2115e2fde32ce8320cd8

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
