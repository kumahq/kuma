FROM gcr.io/distroless/base-nossl-debian12:debug-nonroot@sha256:ccb2092b96479df8f9ace9675596f40fb124862c1f07acbfd86bf2f854883a7f

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
