FROM gcr.io/distroless/base-nossl-debian11:debug-nonroot@sha256:848708456cb7b4ba31650d36e59642f53a02d55f9d0a8623bee868e8fb21b0c9

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
