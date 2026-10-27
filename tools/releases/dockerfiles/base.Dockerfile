FROM gcr.io/distroless/base-nossl-debian13:debug-nonroot@sha256:68f210b8b9d3105847b270a1aef56abfcd8ee1f8bb54921999e451f42d0dad0c

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
