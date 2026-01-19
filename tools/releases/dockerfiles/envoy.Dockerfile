FROM debian:13.3@sha256:5cf544fad978371b3df255b61e209b373583cb88b733475c86e49faa15ac2104 AS envoy
ARG ARCH

COPY /build/artifacts-linux-$ARCH/envoy/envoy /envoy
