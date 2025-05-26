FROM gcr.io/distroless/static-debian12:debug-nonroot@sha256:633d5fa264a127052ca34c3fdaf81ef5a58204770736df9047745919a5b318f6

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
