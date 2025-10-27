FROM gcr.io/distroless/static-debian12:debug-nonroot@sha256:7d3273ed75e3c6b4a159e215dd30187b856fdfdb3266ec7777a3fce51cecccfe

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
