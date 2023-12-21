FROM gcr.io/distroless/static-debian11:debug-nonroot@sha256:1e5b9bb417e6f4ec664b56d9a73148ce5b4662895fbfddce2f27b945364d7948

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
