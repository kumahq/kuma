# use only when root is really needed
FROM gcr.io/distroless/base-nossl-debian13:debug@sha256:bff74e6c262d39b8fe406835e52a7ccd484b6d35f3779641595c53514d1260f9

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
