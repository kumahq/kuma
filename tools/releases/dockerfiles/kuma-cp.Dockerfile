ARG ARCH
FROM kumahq/static-debian12:no-push-$ARCH
ARG ARCH

COPY --chmod=0755 /build/artifacts-linux-${ARCH}/kuma-cp/kuma-cp /usr/bin

ENTRYPOINT ["/usr/bin/kuma-cp"]
