FROM gcr.io/k8s-staging-build-image/distroless-iptables:v0.6.4
ARG ARCH

COPY /build/artifacts-linux-$ARCH/kumactl/kumactl /usr/bin

# this will be from a base image once it is done
COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /kuma/

COPY /tools/releases/templates/NOTICE-kumactl /kuma/NOTICE

ENTRYPOINT ["/usr/bin/kumactl"]
CMD ["install", "transparent-proxy"]
