FROM debian:13.1@sha256:833c135acfe9521d7a0035a296076f98c182c542a2b6b5a0fd7063d355d696be AS envoy
ARG ARCH

COPY /build/artifacts-linux-$ARCH/envoy/envoy /envoy
