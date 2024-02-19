FROM gcr.io/distroless/base-nossl-debian11:debug-nonroot@sha256:8e2bb0958121b9ec2d70fe16145670e11ad9c68c5e234a8bdcf863f2a02ee16e

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
