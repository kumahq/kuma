ARG ARCH
FROM kumahq/base-nossl-debian12:no-push-$ARCH
ARG ARCH

# override NOTICE
COPY /tools/releases/templates/NOTICE /kuma/NOTICE
COPY --chmod=0755 /build/artifacts-linux-$ARCH/kumactl/kumactl /usr/bin/

ENTRYPOINT ["/usr/bin/kumactl"]
