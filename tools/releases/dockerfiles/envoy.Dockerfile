FROM debian:12.9@sha256:35286826a88dc879b4f438b645ba574a55a14187b483d09213a024dc0c0a64ed AS envoy
ARG ARCH

COPY /build/artifacts-linux-$ARCH/envoy/envoy /envoy
