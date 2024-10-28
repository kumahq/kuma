FROM debian:12.7@sha256:e11072c1614c08bf88b543fcfe09d75a0426d90896408e926454e88078274fcb AS envoy
ARG ARCH

COPY /build/artifacts-linux-$ARCH/envoy/envoy /envoy
