FROM debian:13.4@sha256:35b8ff74ead4880f22090b617372daff0ccae742eb5674455d542bef71ef1999 AS envoy
ARG ARCH

COPY /build/artifacts-linux-$ARCH/envoy/envoy /envoy
