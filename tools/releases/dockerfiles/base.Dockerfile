FROM gcr.io/distroless/base-nossl-debian12:debug-nonroot@sha256:a2206e8087a1be3a8b2d575dd82e710e87ce7271e1c985b64cc0a80d14b864a3

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
