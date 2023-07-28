FROM gcr.io/distroless/base-nossl-debian11:debug-nonroot@sha256:d1518145ce30d024ec65ec3251aa5ec645449022eb62fb201d0bbdb04ba6ffa8

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
