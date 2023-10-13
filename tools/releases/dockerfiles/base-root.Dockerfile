# use only when root is really needed
FROM gcr.io/distroless/base-nossl-debian11:debug@sha256:1ae8df52aedbab54421655679dd1830f3da74115b932ac6ad3477bb4f8346bd1

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
