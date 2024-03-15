FROM gcr.io/distroless/base-nossl-debian11:debug-nonroot@sha256:4e4e517e8ab604acc25b25620011cb617a77d9efc72839d61f0aca2eaf07ec8c

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
