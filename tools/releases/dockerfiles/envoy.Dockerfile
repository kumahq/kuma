FROM debian:13.1@sha256:e623a68de39df2046af830adc3c97928bf141c104a13cffc021fce9867aa54fe AS envoy
ARG ARCH

COPY /build/artifacts-linux-$ARCH/envoy/envoy /envoy
