FROM gcr.io/distroless/static-debian11:debug-nonroot@sha256:14a5301feef87db098db662f70814b5f8af95b3bfb84a064dcf4a5b44e513f4d

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
