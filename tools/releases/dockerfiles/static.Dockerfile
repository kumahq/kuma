FROM gcr.io/distroless/static-debian11:debug-nonroot@sha256:deb9b2c21634ea4fe9e6207be8efc75201978c6fd97bf5a5bbf246cb9bb2a154

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
