FROM cgr.dev/chainguard/busybox:latest-glibc@sha256:a5779421f99aee79005c3f598dba1b43b8d83a7c92fbd7e2df244fa674909329

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/

SHELL ["/bin/busybox", "sh", "-c"]
