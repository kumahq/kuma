FROM gcr.io/distroless/static-debian12:debug-nonroot@sha256:e8506419f1230f0622c8209e2dcb28c6166644b2515aa586997bc1fb01bab6a0

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
