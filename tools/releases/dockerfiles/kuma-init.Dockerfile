# until there is a distroless iptables image we have to use something else
FROM ubuntu:jammy-20240227@sha256:77906da86b60585ce12215807090eb327e7386c8fafb5402369e421f44eff17e
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

RUN adduser --system --disabled-password --group kumactl --uid 5678

# As of iptables 1.8, the iptables command line clients come in two different versions/modes:
# "legacy", which uses the kernel iptables API just like iptables 1.6 and earlier did, and
# "nft", which translates the iptables command-line API into the kernel nftables API.
#
# Because they connect to two different subsystems in the kernel, you cannot mix and match
# between them; in particular, if you are adding a new rule that needs to run either before
# or after some existing rules (such as the system firewall rules), then you need to create
# your rule with the same iptables mode as the other rules were created with, since otherwise
# the ordering may not be what you expect. (eg, if you prepend a rule using the nft-based
# client, it will still run after all rules that were added with the legacy iptables client.)
#
# In particular, this means that if you create a container image that will make changes to
# iptables rules in the host network namespace, and you want that container to be able to work
# on any host, then you need to figure out at run time which mode the host is using, and then
# also use that mode yourself. This wrapper is designed to do that for you.
#
# ref. https://github.com/kubernetes-sigs/iptables-wrappers
COPY /build/artifacts-linux-$ARCH/iptables-wrapper/iptables-wrapper-installer.sh \
     /build/artifacts-linux-$ARCH/iptables-wrapper/iptables-wrapper \
     /

RUN /iptables-wrapper-installer.sh

ENTRYPOINT ["/usr/bin/kumactl"]
CMD ["install", "transparent-proxy"]
