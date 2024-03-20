FROM gcr.io/distroless/base-nossl-debian11:debug-nonroot@sha256:933f42b2c067a81c96bfe4fe24d21123081cbec813298db390b52226c7ec7577

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
