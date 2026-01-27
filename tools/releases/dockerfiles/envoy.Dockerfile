FROM debian:13.2@sha256:c71b05eac0b20adb4cdcc9f7b052227efd7da381ad10bb92f972e8eae7c6cdc9 AS envoy
ARG ARCH

COPY /build/artifacts-linux-$ARCH/envoy/envoy /envoy
