# use only when root is really needed
FROM gcr.io/distroless/base-nossl-debian12:debug@sha256:84c4a85988fed534e53dc708615c83dbe44b1e78c06ab8ddf29cb8d9a8b4026b

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
