package v3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
)

var _ = Describe("TcpProxyConfigurer", func() {

	type testCase struct {
		listenerName     string
		listenerProtocol xds.SocketAddressProtocol
		listenerAddress  string
		listenerPort     uint32
		statsName        string
		clusters         []envoy_common.Cluster
		expected         string
	}

	DescribeTable("should generate proper Envoy config with metadata",
		func(given testCase) {
			// when
			listener, err := NewListenerBuilder(envoy_common.APIV3).
				Configure(InboundListener(given.listenerName, given.listenerAddress, given.listenerPort, given.listenerProtocol)).
				Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
					Configure(TcpProxyWithMetadata(given.statsName, given.clusters...)))).
				Build()
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(listener)
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("basic tcp_proxy with a single destination cluster", testCase{
			listenerName:    "inbound:192.168.0.1:8080",
			listenerAddress: "192.168.0.1",
			listenerPort:    8080,
			statsName:       "localhost:8080",
			clusters: []envoy_common.Cluster{envoy_common.NewCluster(
				envoy_common.WithService("localhost:8080"),
				envoy_common.WithWeight(200),
				envoy_common.WithTags(map[string]string{"version": "v1"}),
			)},
			expected: `
        name: inbound:192.168.0.1:8080
        trafficDirection: INBOUND
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
              metadataMatch:
                filterMetadata:
                  envoy.lb:
                    version: v1
              statPrefix: localhost_8080
`,
		}),
		Entry("basic tcp_proxy with weighted destination clusters", testCase{
			listenerName:    "inbound:127.0.0.1:5432",
			listenerAddress: "127.0.0.1",
			listenerPort:    5432,
			statsName:       "db",
			clusters: []envoy_common.Cluster{
				envoy_common.NewCluster(
					envoy_common.WithService("db"),
					envoy_common.WithWeight(10),
					envoy_common.WithTags(map[string]string{"kuma.io/service": "db", "version": "v1"}),
				),
				envoy_common.NewCluster(
					envoy_common.WithService("db"),
					envoy_common.WithWeight(90),
					envoy_common.WithTags(map[string]string{"kuma.io/service": "db", "version": "v2"}),
				),
			},
			expected: `
            address:
              socketAddress:
                address: 127.0.0.1
                portValue: 5432
            enableReusePort: false
            filterChains:
            - filters:
              - name: envoy.filters.network.tcp_proxy
                typedConfig:
                  '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                  statPrefix: db
                  weightedClusters:
                    clusters:
                    - metadataMatch:
                        filterMetadata:
                          envoy.lb:
                            version: v1
                      name: db
                      weight: 10
                    - metadataMatch:
                        filterMetadata:
                          envoy.lb:
                            version: v2
                      name: db
                      weight: 90
            name: inbound:127.0.0.1:5432
            trafficDirection: INBOUND`,
		}),
	)

	DescribeTable("should generate proper Envoy config without metadata",
		func(given testCase) {
			// when
			listener, err := NewListenerBuilder(envoy_common.APIV3).
				Configure(InboundListener(given.listenerName, given.listenerAddress, given.listenerPort, given.listenerProtocol)).
				Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
					Configure(TcpProxy(given.statsName, given.clusters...)))).
				Build()
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(listener)
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("basic tcp_proxy with a single destination cluster", testCase{
			listenerName:    "inbound:192.168.0.1:8080",
			listenerAddress: "192.168.0.1",
			listenerPort:    8080,
			statsName:       "localhost:8080",
			clusters: []envoy_common.Cluster{
				envoy_common.NewCluster(
					envoy_common.WithService("localhost:8080"),
					envoy_common.WithWeight(200),
				),
			},

			expected: `
        name: inbound:192.168.0.1:8080
        trafficDirection: INBOUND
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
`,
		}),
		Entry("basic tcp_proxy with weighted destination clusters", testCase{
			listenerName:    "inbound:127.0.0.1:5432",
			listenerAddress: "127.0.0.1",
			listenerPort:    5432,
			statsName:       "db",
			clusters: []envoy_common.Cluster{
				envoy_common.NewCluster(
					envoy_common.WithService("db-0"),
					envoy_common.WithWeight(10),
				),
				envoy_common.NewCluster(
					envoy_common.WithService("db-1"),
					envoy_common.WithWeight(90),
				)},
			expected: `
            address:
              socketAddress:
                address: 127.0.0.1
                portValue: 5432
            enableReusePort: false
            filterChains:
            - filters:
              - name: envoy.filters.network.tcp_proxy
                typedConfig:
                  '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                  statPrefix: db
                  weightedClusters:
                    clusters:
                    - name: db-0
                      weight: 10
                    - name: db-1
                      weight: 90
            name: inbound:127.0.0.1:5432
            trafficDirection: INBOUND`,
		}),
	)
})
