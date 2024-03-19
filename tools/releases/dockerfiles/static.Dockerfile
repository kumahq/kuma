FROM gcr.io/distroless/static-debian11:debug-nonroot@sha256:c7c26a49568781eb04040f48c305b676f290dbaa9112b2f9de02c8df57a881b6

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
