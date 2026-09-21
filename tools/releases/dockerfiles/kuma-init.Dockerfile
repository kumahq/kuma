FROM gcr.io/k8s-staging-build-image/distroless-iptables:v0.11.0@sha256:bedc7f0c0dc8e42a316f1cd6d6ee321eb4159e1f75ac30875bdfcf13446bc7a0
ARG ARCH

COPY /build/artifacts-linux-$ARCH/kumactl/kumactl /usr/bin

# this will be from a base image once it is done
COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /kuma/

COPY /tools/releases/templates/NOTICE /kuma/NOTICE

# Copy modified system files
COPY /tools/releases/templates/passwd /etc/passwd
COPY /tools/releases/templates/group /etc/group

ENTRYPOINT ["/usr/bin/kumactl"]
CMD ["install", "transparent-proxy"]
