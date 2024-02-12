# use only when root is really needed
FROM gcr.io/distroless/base-nossl-debian11:debug@sha256:70d27f6659eebe180c82d343c99fdf000938c7574254f22a53743a7f72b9bb1b

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
