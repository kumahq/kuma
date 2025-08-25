FROM gcr.io/distroless/static-debian12:debug-nonroot@sha256:0895d6fc256a6938a60c87d92e1148eec0d36198bff9c5d3082e6a56db7756bd

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
