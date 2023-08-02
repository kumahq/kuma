FROM gcr.io/k8s-staging-build-image/distroless-iptables:0.26.0@sha256:7ee5c2ea5b0bacb43a2acee7e4a38816d8de193488abe486a252bdf19e03967d
ARG ARCH

COPY /build/artifacts-linux-$ARCH/kumactl/kumactl /usr/bin

# this will be from a base image once it is done
COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /kuma/

COPY /tools/releases/scripts/iptables /usr/local/sbin/
COPY /tools/releases/templates/NOTICE-kumactl /kuma/NOTICE

ENTRYPOINT ["/usr/bin/kumactl"]
CMD ["install", "transparent-proxy"]
