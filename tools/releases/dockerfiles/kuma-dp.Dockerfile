ARG ARCH
FROM kumahq/envoy:no-push-$ARCH AS envoy
FROM kumahq/base-nossl-debian12:no-push-$ARCH
ARG ARCH

COPY /build/artifacts-linux-$ARCH/kuma-dp/kuma-dp \
    /build/artifacts-linux-$ARCH/coredns/coredns \
    /usr/bin/

COPY --from=envoy /envoy /usr/bin/envoy

ENTRYPOINT ["/usr/bin/kuma-dp"]
