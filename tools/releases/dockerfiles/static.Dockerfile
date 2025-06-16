FROM gcr.io/distroless/static-debian12:debug-nonroot@sha256:edbeb7a4e79938116dc9cb672b231792e0b5ac86c56fb49781a79e54f3842c67

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
