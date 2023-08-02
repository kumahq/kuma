ARG ARCH
FROM kumahq/base-busybox:no-push-$ARCH
ARG ARCH

COPY /build/artifacts-linux-${ARCH}/kuma-cp/kuma-cp /usr/bin

ENTRYPOINT ["/usr/bin/kuma-cp"]
