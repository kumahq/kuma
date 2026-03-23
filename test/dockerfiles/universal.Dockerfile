ARG ARCH
FROM kumahq/envoy:no-push-$ARCH AS envoy
FROM ghcr.io/spiffe/spire-server:1.14.4@sha256:27c7d356768b8641c569745e1121328affeb4aaabe0c974d33ff92dddecf30ef AS spire_server
FROM ghcr.io/spiffe/spire-agent:1.14.4@sha256:f93996a9396ec4042a48042085c2abf1a0f8f0f3b339571e06ebfebbf94bb830 AS spire_agent
# Built in github.com/kumahq/ci-tools
FROM ghcr.io/kumahq/ubuntu-netools:main@sha256:bd6c59809689d38a97bedd644b11ead272e67923c0ea7c2a9abb72598252a54d

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
