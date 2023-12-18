# use only when root is really needed
FROM gcr.io/distroless/base-nossl-debian11:debug@sha256:61c9d7a17c185b67fad756ed5efaa19cba2ee7af666a4ba8f99339b4560a8539

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
