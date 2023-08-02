ARG ARCH
FROM kumahq/base-busybox:no-push-$ARCH
ARG ARCH

# override NOTICE
COPY /tools/releases/templates/NOTICE-kumactl /kuma/NOTICE
COPY /build/artifacts-linux-$ARCH/kumactl/kumactl /usr/bin/

ENTRYPOINT ["/usr/bin/kumactl"]
