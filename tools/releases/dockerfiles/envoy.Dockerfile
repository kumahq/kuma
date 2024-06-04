FROM debian:12.5@sha256:fac2c0fd33e88dfd3bc88a872cfb78dcb167e74af6162d31724df69e482f886c as envoy
ARG ARCH

COPY /build/artifacts-linux-$ARCH/envoy/envoy /envoy
