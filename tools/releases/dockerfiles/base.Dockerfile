FROM gcr.io/distroless/base-nossl-debian12:debug-nonroot@sha256:83b8737817cda240f3f75e9acaa2b6fa547e2693e7e8dc4594c7ce98a0e5765e

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
