# use only when root is really needed
FROM gcr.io/distroless/base-nossl-debian12:debug@sha256:51c3587676d971b6744b1ac93bc7f64c6604e14b70ed87d1f353a6c060407e8a

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
