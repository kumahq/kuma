FROM gcr.io/distroless/base-nossl-debian13:debug-nonroot@sha256:4d040a10bd12b8a0a44346381a1cbbedf3965dcda822b0ae4314ce1bf53a2111

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
