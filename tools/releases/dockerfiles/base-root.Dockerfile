# use only when root is really needed
FROM gcr.io/distroless/base-nossl-debian11:debug@sha256:d66c60eff6c55972af9e661a57c1afe96ef4ddfa4fff37b625a448df41a15820

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
