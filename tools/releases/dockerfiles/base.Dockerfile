FROM gcr.io/distroless/base-nossl-debian12:debug-nonroot@sha256:a3d240eff5bf9730994a71720e7415058a266d891fb9578464ccc9a945d5147c

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
