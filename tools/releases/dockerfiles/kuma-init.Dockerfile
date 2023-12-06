# until there is a distroless iptables image we have to use something else
FROM ubuntu:jammy@sha256:8eab65df33a6de2844c9aefd19efe8ddb87b7df5e9185a4ab73af936225685bb
ARG ARCH

RUN apt-get update && \
    apt-get install --no-install-recommends -y iptables=1.8.7-1ubuntu5.1 iproute2=5.15.0-1ubuntu2 && \
    rm -rf /var/lib/apt/lists/*

COPY /build/artifacts-linux-$ARCH/kumactl/kumactl /usr/bin

# this will be from a base image once it is done
COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /kuma/

COPY /tools/releases/templates/NOTICE /kuma/NOTICE

RUN update-alternatives --set iptables /usr/sbin/iptables-legacy && \
    adduser --system --disabled-password --group kumactl --uid 5678

ENTRYPOINT ["/usr/bin/kumactl"]
CMD ["install", "transparent-proxy"]
