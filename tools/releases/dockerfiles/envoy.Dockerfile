FROM debian:12@sha256:fab22df37377621693c68650b06680c0d8f7c6bf816ec92637944778db3ca2c0 as envoy
ARG ARCH

COPY /build/artifacts-linux-$ARCH/envoy/envoy /envoy
