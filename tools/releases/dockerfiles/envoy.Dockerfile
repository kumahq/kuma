FROM debian:12.11@sha256:bd73076dc2cd9c88f48b5b358328f24f2a4289811bd73787c031e20db9f97123 AS envoy
ARG ARCH

COPY /build/artifacts-linux-$ARCH/envoy/envoy /envoy
