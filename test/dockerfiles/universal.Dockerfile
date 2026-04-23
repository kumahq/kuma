ARG ARCH
FROM kumahq/envoy:no-push-$ARCH AS envoy
FROM ghcr.io/spiffe/spire-server:1.14.5@sha256:295af37293ffef6530ba0cafce055c41fd4d26dbabaca24944c029bbc8291214 AS spire_server
FROM ghcr.io/spiffe/spire-agent:1.14.5@sha256:efb8d29ace803852c8f3d0001c3d183102bfda2c6ff52e4a93a067269a2caa5e AS spire_agent
# Built in github.com/kumahq/ci-tools
FROM ghcr.io/kumahq/ubuntu-netools:main@sha256:487f66a9386f17fb2ba4cf5271bbf6ce8c79daadb3ad8dc80406acddb99fb110

ARG ARCH

RUN useradd -u 5678 -U kuma-dp

RUN mkdir /kuma
RUN mkdir /spire
RUN echo "# use this file to override default configuration of \`kuma-cp\`" > /kuma/kuma-cp.conf \
    && chmod a+rw /kuma/kuma-cp.conf

ADD /build/artifacts-linux-$ARCH/kuma-cp/kuma-cp /usr/bin
ADD /build/artifacts-linux-$ARCH/kuma-dp/kuma-dp /usr/bin
COPY --from=envoy /envoy /usr/bin/envoy
COPY --from=envoy /envoy /usr/bin/envoy
COPY --from=spire_agent /opt/spire/bin/spire-agent /usr/bin
COPY --from=spire_server /opt/spire/bin/spire-server /usr/bin
ADD /test/dockerfiles/spire-server.conf /spire
ADD /test/dockerfiles/spire-agent.conf /spire
ADD /build/artifacts-linux-$ARCH/coredns/coredns /usr/bin
ADD /build/artifacts-linux-$ARCH/kumactl/kumactl /usr/bin
ADD /build/artifacts-linux-$ARCH/test-server/test-server /usr/bin
ADD /test/server/certs/server.crt /kuma
ADD /test/server/certs/server.key /kuma
