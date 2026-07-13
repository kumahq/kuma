FROM gcr.io/k8s-staging-build-image/distroless-iptables:v0.9.3@sha256:c9995b6b72ef54e8586ffb5dbee77573953992a103b304af7952523b4bac7ff5
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
