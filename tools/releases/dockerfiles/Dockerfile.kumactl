ARG ARCH
FROM kumahq/base-nossl-debian11:no-push-$ARCH
ARG ARCH

# override NOTICE
COPY /tools/releases/templates/NOTICE-kumactl /kuma/NOTICE
COPY /build/artifacts-linux-$ARCH/kumactl/kumactl /usr/bin/
