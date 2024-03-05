FROM gcr.io/distroless/base-nossl-debian11:debug-nonroot@sha256:5b76b43a16c9b8ebde713a1202c773fe7981f527d9c6c5916f38ba97529c2fc2

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
