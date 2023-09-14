# use only when root is really needed
FROM gcr.io/distroless/base-nossl-debian11:debug@sha256:b1f88dedef04a62c818592267d56723ab338a3071f075948dc699ae874759f2a

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
