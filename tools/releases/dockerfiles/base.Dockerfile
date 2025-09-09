FROM gcr.io/distroless/base-nossl-debian12:debug-nonroot@sha256:6c2580fdeefdd4a30c79bbb831462d8126acea9e003e8a49ed1ba6a3d1204ba3

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
