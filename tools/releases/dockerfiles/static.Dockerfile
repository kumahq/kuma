FROM gcr.io/distroless/static-debian12:debug-nonroot@sha256:14a28be5b9e710e8d5a519cd7d76ab2714b35edb641307c07a264372ed0c81a9

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
