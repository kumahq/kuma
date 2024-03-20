# use only when root is really needed
FROM gcr.io/distroless/base-nossl-debian11:debug@sha256:133d7a918e622e6d8b986f070f1ac3abf78926aef9b915543442ce37f7df2d85

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
