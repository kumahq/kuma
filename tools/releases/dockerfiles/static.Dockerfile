FROM gcr.io/distroless/static-debian12:debug-nonroot@sha256:a855ba843839f3344272cb64183489d91c190af11bec454e5d17f341255944e1

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
