# use only when root is really needed
FROM gcr.io/distroless/base-nossl-debian12:debug@sha256:1a14fd3ffe3745e5523faa1740904dfd851c324957e230ea2db31601e2f537ec

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
