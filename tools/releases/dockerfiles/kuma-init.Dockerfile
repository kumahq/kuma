# until there is a distroless iptables image we have to use something else
FROM ubuntu:lunar@sha256:a7c141915b6648277343c50356614414ff4244ba86ae38ab27a1f3fb9ffee5b3
ARG ARCH

RUN apt-get update && \
    apt-get install --no-install-recommends -y iptables=1.8.7-1ubuntu7 iproute2=6.1.0-1ubuntu2 adduser=3.129ubuntu1 && \
    rm -rf /var/lib/apt/lists/*

COPY /build/artifacts-linux-$ARCH/kumactl/kumactl /usr/bin

# this will be from a base image once it is done
COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /kuma/

COPY /tools/releases/templates/NOTICE-kumactl /kuma/NOTICE

RUN update-alternatives --set iptables /usr/sbin/iptables-legacy && \
    adduser --system --disabled-password --group kumactl --uid 5678

ENTRYPOINT ["/usr/bin/kumactl"]
CMD ["install", "transparent-proxy"]
