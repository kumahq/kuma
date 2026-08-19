FROM gcr.io/distroless/static-debian12:debug-nonroot@sha256:2b3c67db50828b3b13277c0e00ec9550cac00dd2b7e0683235c7770c2227e0df

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
