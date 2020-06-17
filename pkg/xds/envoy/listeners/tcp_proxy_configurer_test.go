package listeners_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/xds/envoy/listeners"

	util_proto "github.com/Kong/kuma/pkg/util/proto"
	envoy_common "github.com/Kong/kuma/pkg/xds/envoy"
)

var _ = Describe("TcpProxyConfigurer", func() {

	type testCase struct {
		listenerName    string
		listenerAddress string
		listenerPort    uint32
		statsName       string
		clusters        []envoy_common.ClusterSubset
		expected        string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			listener, err := NewListenerBuilder().
				Configure(InboundListener(given.listenerName, given.listenerAddress, given.listenerPort)).
				Configure(FilterChain(NewFilterChainBuilder().
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
			clusters: []envoy_common.ClusterSubset{
				{ClusterName: "localhost:8080", Weight: 200},
			},
			expected: `
        name: inbound:192.168.0.1:8080
        trafficDirection: INBOUND
        address:
          socketAddress:
            address: 192.168.0.1
            portValue: 8080
        filterChains:
        - filters:
          - name: envoy.tcp_proxy
            typedConfig:
              '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
              cluster: localhost:8080
              statPrefix: localhost_8080
`,
		}),
		Entry("basic tcp_proxy with weighted destination clusters", testCase{
			listenerName:    "inbound:127.0.0.1:5432",
			listenerAddress: "127.0.0.1",
			listenerPort:    5432,
			statsName:       "db",
			clusters: []envoy_common.ClusterSubset{{
				ClusterName: "db",
				Weight:      10,
				Tags:        map[string]string{"service": "db", "version": "v1"},
			}, {
				ClusterName: "db",
				Weight:      90,
				Tags:        map[string]string{"service": "db", "version": "v2"},
			}},
			expected: `
            address:
              socketAddress:
                address: 127.0.0.1
                portValue: 5432
            filterChains:
            - filters:
              - name: envoy.tcp_proxy
                typedConfig:
                  '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
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
})
