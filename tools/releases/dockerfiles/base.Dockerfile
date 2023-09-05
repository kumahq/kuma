FROM gcr.io/distroless/base-nossl-debian11:debug-nonroot@sha256:895b0a6e273b777d9422d739b8222538b5d16ab97c61c1322c379f723c7ae4b2

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
