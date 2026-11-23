# use only when root is really needed
FROM gcr.io/distroless/base-nossl-debian13:debug@sha256:4d896520c621a590b5c8f3d930386765199d6fb4f1a000e1a88dde707fd897ee

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
