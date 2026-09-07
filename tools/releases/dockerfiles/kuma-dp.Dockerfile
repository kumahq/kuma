ARG ARCH
FROM kumahq/envoy:no-push-$ARCH AS envoy
FROM kumahq/base-nossl-debian12:no-push-$ARCH
ARG ARCH

COPY --chmod=0755 /build/artifacts-linux-$ARCH/kuma-dp/kuma-dp \
    /usr/bin/

COPY --from=envoy /envoy /usr/bin/envoy

ENTRYPOINT ["/usr/bin/kuma-dp"]
