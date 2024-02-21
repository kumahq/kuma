# use only when root is really needed
FROM gcr.io/distroless/base-nossl-debian11:debug@sha256:729fa645347060e2fd36ad8956a1ea36c0191cb3de39f7b8a94d92da7ecddef6

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/busybox/busybox", "sh", "-c"]
