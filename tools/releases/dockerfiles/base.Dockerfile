FROM gcr.io/distroless/base-nossl-debian12:debug-nonroot@sha256:73a38455a118dc5dcd63ff83469b3b2de6b64e4747ee79e758c6c80ba392e220

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
