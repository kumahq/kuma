FROM gcr.io/distroless/base-nossl-debian12:debug-nonroot@sha256:32d0efd5a21d0274ba4d04c5bc233fec686a4d5fe28aca3447926d5304f1b102

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
