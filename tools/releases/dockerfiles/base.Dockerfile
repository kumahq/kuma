FROM gcr.io/distroless/base-nossl-debian12:debug-nonroot@sha256:ef708361af252a61d6d9bf32c6605f530971c9da3c22b1a403fbddc1a6d12ee6

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
