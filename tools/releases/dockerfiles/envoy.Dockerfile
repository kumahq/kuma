FROM debian:12@sha256:b91baba9c2cae5edbe3b0ff50ae8f05157e3ae6f018372dcfc3aba224acb392b as envoy
ARG ARCH

COPY /build/artifacts-linux-$ARCH/envoy/envoy /envoy
