ARG ARCH
FROM kumahq/envoy:no-push-$ARCH as envoy
FROM kumahq/base-nossl-debian11:no-push-$ARCH
ARG ARCH

COPY /build/artifacts-linux-$ARCH/kuma-dp/kuma-dp \
    /build/artifacts-linux-$ARCH/coredns/coredns \
    /usr/bin/

COPY --from=envoy /envoy /usr/bin/envoy
COPY --from=envoy /glibc/glibc-compat /usr/bin/glibc-compat

ENTRYPOINT ["/usr/bin/kuma-dp"]
