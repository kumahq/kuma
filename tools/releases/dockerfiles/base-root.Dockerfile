# use only when root is really needed
FROM gcr.io/distroless/base-nossl-debian13:debug@sha256:d085c34ea66f46cdd1ff439c01f88c6875f6f49adeb2c1fcc7a22d33e1af515e

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
