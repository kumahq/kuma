package v3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
)

var _ = Describe("ServerSideStaticTLS", func() {
	It("should generate proper Envoy config", func() {
		// given
		certs := core_xds.ServerSideTLSCertPaths{
			CertPath: "/tmp/cert.pem",
			KeyPath:  "/tmp/key.pem",
		}

		cluster := envoy_common.NewCluster(
			envoy_common.WithService("localhost:8080"),
			envoy_common.WithWeight(200),
		)

		// when
		listener, err := NewInboundListenerBuilder(envoy_common.APIV3, "192.168.0.1", 8080, core_xds.SocketAddressProtocolTCP).
			Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, "").
				Configure(ServerSideStaticTLS(certs)).
				Configure(TcpProxyDeprecated("localhost:8080", cluster)))).
			Build()

		// then
		Expect(err).ToNot(HaveOccurred())
		actual, err := util_proto.ToYAML(listener)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(MatchYAML(`
            address:
              socketAddress:
                address: 192.168.0.1
                portValue: 8080
            enableReusePort: false
            filterChains:
            - filters:
              - name: envoy.filters.network.tcp_proxy
                typedConfig:
                  '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                  cluster: localhost:8080
                  statPrefix: localhost_8080
              transportSocket:
                name: envoy.transport_sockets.tls
                typedConfig:
                  '@type': type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.DownstreamTlsContext
                  commonTlsContext:
                    tlsCertificates:
                    - certificateChain:
                        filename: "/tmp/cert.pem"
                      privateKey:
                        filename: "/tmp/key.pem"
            name: inbound:192.168.0.1:8080
            trafficDirection: INBOUND`))
	})
})
