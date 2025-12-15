ARG ARCH
FROM kumahq/envoy:no-push-$ARCH AS envoy
FROM ghcr.io/spiffe/spire-server:1.13.3 AS spire_server
FROM ghcr.io/spiffe/spire-agent:1.13.3 AS spire_agent
# Built in github.com/kumahq/ci-tools
FROM ghcr.io/kumahq/ubuntu-netools:main@sha256:9c4e99b4496093762c3b953bee7ad02c6ec8afc12c0ebc4b5b9e2855d5af0891

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
