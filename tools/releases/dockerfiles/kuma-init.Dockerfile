FROM gcr.io/k8s-staging-build-image/distroless-iptables:v0.7.3@sha256:25cc119a89bf860f964e5c47d0422eabcd260424692a2b6d4e7a74d5c3bb2231
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
