# use only when root is really needed
FROM gcr.io/distroless/base-nossl-debian12:debug@sha256:1b5a39adf37ba68a459a3cb412c99d0f30227c27dd12b38e0757d9aa1d403130

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
