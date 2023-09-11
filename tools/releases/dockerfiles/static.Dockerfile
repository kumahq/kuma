FROM gcr.io/distroless/static-debian11:debug-nonroot@sha256:a8017a4f68c33e1489b4ad2b88dd0e8ddbe420b0c7a5c60716c19304b0f5883e

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
