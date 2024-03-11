# use only when root is really needed
FROM gcr.io/distroless/base-nossl-debian11:debug@sha256:534434d03f90ed54c7043d5580438d5d5cc8f7322f444bb01f2c1862e8f3c82f

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
