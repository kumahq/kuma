FROM gcr.io/distroless/static-debian13:debug-nonroot@sha256:12c2026b21e05857a4c730b4dbf812becf02b396cb7916166c8305f45f140ab0

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
