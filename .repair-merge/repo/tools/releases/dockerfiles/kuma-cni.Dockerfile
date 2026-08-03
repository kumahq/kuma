ARG ARCH
FROM kumahq/base-root-debian12:no-push-$ARCH
ARG ARCH

COPY /build/artifacts-linux-$ARCH/kuma-cni/kuma-cni /opt/cni/bin/kuma-cni
COPY /build/artifacts-linux-$ARCH/install-cni/install-cni /install-cni

ENV PATH=$PATH:/opt/cni/bin

WORKDIR /opt/cni/bin

ENTRYPOINT ["/install-cni"]
