FROM debian:12@sha256:9f76a008888da28c6490bedf7bdaa919bac9b2be827afd58d6eb1b916e1e5918 as envoy
ARG ARCH

COPY /build/artifacts-linux-$ARCH/envoy/envoy /envoy
