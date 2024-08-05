FROM debian:12.6@sha256:45f2e735295654f13e3be10da2a6892c708f71a71be845818f6058982761a6d3 AS envoy
ARG ARCH

COPY /build/artifacts-linux-$ARCH/envoy/envoy /envoy
