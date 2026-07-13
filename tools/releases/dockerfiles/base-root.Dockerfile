# use only when root is really needed
FROM gcr.io/distroless/base-nossl-debian12:debug@sha256:2b89aea887933b8595f62c3d7e05a2718a0594e21e506952ce55b68b537f4fb6

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
