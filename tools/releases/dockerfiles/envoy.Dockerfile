FROM debian:13.7@sha256:9cc080028c43b27d2074d63a5f9caf7166d731494965616c1a6d2827a004585c AS envoy
ARG ARCH

COPY /build/artifacts-linux-$ARCH/envoy/envoy /envoy
