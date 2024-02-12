# until there is a distroless iptables image we have to use something else
FROM ubuntu:jammy@sha256:2b7412e6465c3c7fc5bb21d3e6f1917c167358449fecac8176c6e496e5c1f05f
ARG ARCH

RUN apt-get update && \
    apt-get install --no-install-recommends -y iptables=1.8.7-1ubuntu5.2 iproute2=5.15.0-1ubuntu2 && \
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
