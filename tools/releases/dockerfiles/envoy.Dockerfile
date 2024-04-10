FROM debian:12.5@sha256:b37bc259c67238d814516548c17ad912f26c3eed48dd9bb54893eafec8739c89 as envoy
ARG ARCH

COPY /build/artifacts-linux-$ARCH/envoy/envoy /envoy
