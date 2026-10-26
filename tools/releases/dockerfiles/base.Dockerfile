FROM gcr.io/distroless/base-nossl-debian13:debug-nonroot@sha256:a6b55ccdc539958a55408946fa5d3e3475137ff6b5e50153b71cdd35a4cf7ecf

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
