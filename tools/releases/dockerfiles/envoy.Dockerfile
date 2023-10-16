FROM debian:12@sha256:7d3e8810c96a6a278c218eb8e7f01efaec9d65f50c54aae37421dc3cbeba6535 as envoy
ARG ARCH

COPY /build/artifacts-linux-$ARCH/envoy/envoy /envoy
