FROM gcr.io/distroless/base-nossl-debian11:debug-nonroot@sha256:ae2d0ffab317ec2f30c9309c7116b4bbf7ac3a2aad3eb64d24e4c1cb3f406378

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
