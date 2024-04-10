FROM gcr.io/distroless/static-debian11:debug-nonroot@sha256:459f8abafd03779490115cdf9bdf4f208d39ca8e66dcee3a78baa73b18fca504

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
