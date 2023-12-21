FROM debian:12.4@sha256:bac353db4cc04bc672b14029964e686cd7bad56fe34b51f432c1a1304b9928da as envoy
ARG ARCH

COPY /build/artifacts-linux-$ARCH/envoy/envoy /envoy
