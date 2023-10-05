FROM debian:12@sha256:eaace54a93d7b69c7c52bb8ddf9b3fcba0c106a497bc1fdbb89a6299cf945c63 as envoy
ARG ARCH

COPY /build/artifacts-linux-$ARCH/envoy/envoy /envoy
