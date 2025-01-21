FROM debian:12.9@sha256:321341744acb788e251ebd374aecc1a42d60ce65da7bd4ee9207ff6be6686a62 AS envoy
ARG ARCH

COPY /build/artifacts-linux-$ARCH/envoy/envoy /envoy
