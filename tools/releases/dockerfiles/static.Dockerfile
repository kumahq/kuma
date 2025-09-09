FROM gcr.io/distroless/static-debian12:debug-nonroot@sha256:f556e84ce8fb32fdd5a4478550b720330a22f13ff52dab68d7a7bb3a4266829d

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
