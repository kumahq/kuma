FROM gcr.io/k8s-staging-build-image/distroless-iptables:v0.9.2@sha256:c2c638dd6f7ce3d2eb79b52e207afd82c1953ca01d72972168ed4bd0a55ebe7e
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
