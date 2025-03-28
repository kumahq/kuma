# use only when root is really needed
FROM gcr.io/distroless/base-nossl-debian12:debug@sha256:1368c7b5edff36920f9169f87f8e8108f532b6f343ab68ce7895a184bbfdb930

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
