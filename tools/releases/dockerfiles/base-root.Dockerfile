# use only when root is really needed
FROM gcr.io/distroless/base-nossl-debian12:debug@sha256:94b009dc8031fbdcb3038ba6452fbdb267ddb8cf2512538c1a9e556ef3cf9568

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
