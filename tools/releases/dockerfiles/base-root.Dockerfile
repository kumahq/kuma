# use only when root is really needed
FROM gcr.io/distroless/base-nossl-debian12:debug@sha256:1321f45efc120268506fa83b0b6ac8e9086c1048e7c95253c41de79eb3e1f8b0

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
