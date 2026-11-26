# use only when root is really needed
FROM gcr.io/distroless/base-nossl-debian13:latest@sha256:f21fdfe4ede467d9ea8a6caaec1f8381ea4502243a3e8bb1c43df80c6fbaaeae

COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /tools/releases/templates/NOTICE \
    /kuma/
