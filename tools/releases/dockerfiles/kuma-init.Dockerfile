FROM gcr.io/k8s-staging-build-image/distroless-iptables:v0.8.7@sha256:cdac80d8683f9d8d3aa177383e3a5ab166df91d4e884199023d4a5b088f6015f
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
