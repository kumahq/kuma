FROM debian:12@sha256:b4042f895d5d1f8df415caebe7c416f9dbcf0dc8867abb225955006de50b21f3 as envoy
ARG ARCH

COPY /build/artifacts-linux-$ARCH/envoy/envoy /envoy
