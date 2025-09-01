# use only when root is really needed
FROM gcr.io/distroless/base-nossl-debian12:debug@sha256:7557eb8546b0e4b6ed845e33c55f8411e534885045c65e9ffe786f533a4ca8fb

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
