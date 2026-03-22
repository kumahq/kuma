ARG ARCH
FROM kumahq/envoy:no-push-$ARCH AS envoy
FROM ghcr.io/spiffe/spire-server:1.14.4@sha256:27c7d356768b8641c569745e1121328affeb4aaabe0c974d33ff92dddecf30ef AS spire_server
FROM ghcr.io/spiffe/spire-agent:1.14.2@sha256:f8c40f435d42bd8b5420768b95f6b41acc695fb13cd9f9728d27c8e21e07d803 AS spire_agent
# Built in github.com/kumahq/ci-tools
FROM ghcr.io/kumahq/ubuntu-netools:main@sha256:89d4405daea10790f8f7380a73cf4ae2229689c15f290cb0b9048f18cc33da29

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
