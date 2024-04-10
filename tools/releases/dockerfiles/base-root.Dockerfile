# use only when root is really needed
FROM gcr.io/distroless/base-nossl-debian11:debug@sha256:4cba3ac67aa5b17e3833bf0b53bf40f8345155a0df76eec5a624d1cfe61f0789

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
