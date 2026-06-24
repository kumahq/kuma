FROM gcr.io/distroless/base-nossl-debian12:debug-nonroot@sha256:6cec64391f19300e64ff6bbc46a867138c9797913690bafc44e4b22520dd384b

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
