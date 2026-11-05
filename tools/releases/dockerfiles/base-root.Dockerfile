# use only when root is really needed
FROM gcr.io/distroless/base-nossl-debian13:debug@sha256:4a1b47e4dc79d5331f6b9265f8933a96351fb6341441902acf8dc8a9a6d25630

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
