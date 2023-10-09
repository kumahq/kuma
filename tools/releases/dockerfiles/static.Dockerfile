FROM gcr.io/distroless/static-debian11:debug-nonroot@sha256:cdb2034c38b2f2bd0a99f08191a44831a04220c81aab97b2397d2ecf1082db5f

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
