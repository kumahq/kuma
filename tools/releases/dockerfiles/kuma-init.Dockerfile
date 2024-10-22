<<<<<<< HEAD
# until there is a distroless iptables image we have to use something else
FROM ubuntu:jammy-20240530@sha256:19478ce7fc2ffbce89df29fea5725a8d12e57de52eb9ea570890dc5852aac1ac
=======
FROM gcr.io/k8s-staging-build-image/distroless-iptables:v0.6.4
>>>>>>> d5963e709 (chore(kuma-init): use distroless image (#5945))
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
