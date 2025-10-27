# use only when root is really needed
FROM gcr.io/distroless/base-nossl-debian12:debug@sha256:12dbb4f46c5f712fe3da1c7a441602ee91eb87a5d46b0e725b4440b852000538

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
