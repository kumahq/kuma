FROM gcr.io/distroless/base-nossl-debian12:debug-nonroot@sha256:d86c78b580ebcd04f3606c8201fd1ea76e457c21059cc1d2a17695b6f9ebf121

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
