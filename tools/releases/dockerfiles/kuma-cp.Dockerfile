ARG ARCH
FROM kumahq/static-debian13:no-push-$ARCH
ARG ARCH

COPY /build/artifacts-linux-${ARCH}/kuma-cp/kuma-cp /usr/bin

ENTRYPOINT ["/usr/bin/kuma-cp"]
