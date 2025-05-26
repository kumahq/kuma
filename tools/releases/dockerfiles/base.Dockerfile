FROM gcr.io/distroless/base-nossl-debian12:debug-nonroot@sha256:081a62d414528059b08f5054fefd56e3df3c2d48bdb4090281baa76b6240ca59

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
