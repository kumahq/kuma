FROM debian:13.0@sha256:6d87375016340817ac2391e670971725a9981cfc24e221c47734681ed0f6c0f5 AS envoy
ARG ARCH

COPY /build/artifacts-linux-$ARCH/envoy/envoy /envoy
