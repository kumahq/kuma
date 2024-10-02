FROM debian:12.7@sha256:27586f4609433f2f49a9157405b473c62c3cb28a581c413393975b4e8496d0ab AS envoy
ARG ARCH

COPY /build/artifacts-linux-$ARCH/envoy/envoy /envoy
