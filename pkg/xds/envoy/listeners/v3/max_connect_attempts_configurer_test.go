package v3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
)

var _ = Describe("MaxConnectAttemptsConfigurer", func() {
	type testCase struct {
		listenerName       string
		listenerAddress    string
		listenerPort       uint32
		listenerProtocol   core_xds.SocketAddressProtocol
		statsName          string
		clusters           []envoy_common.Cluster
		maxConnectAttempts uint32
		expected           string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// given
			retry := &core_mesh.RetryResource{
				Meta: nil,
				Spec: &mesh_proto.Retry{
					Sources:      nil,
					Destinations: nil,
					Conf: &mesh_proto.Retry_Conf{
						Tcp: &mesh_proto.Retry_Conf_Tcp{
							MaxConnectAttempts: given.maxConnectAttempts,
						},
					},
				},
			}

			// when
			listener, err := NewListenerBuilder(envoy_common.APIV3).
				Configure(OutboundListener(given.listenerName, given.listenerAddress, given.listenerPort, given.listenerProtocol)).
				Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
					Configure(TcpProxy(given.statsName, given.clusters...)).
					Configure(MaxConnectAttempts(retry)))).
				Build()
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(listener)
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("basic tcp_proxy", testCase{
			listenerName:    "outbound:127.0.0.1:5432",
			listenerAddress: "127.0.0.1",
			listenerPort:    5432,
			statsName:       "db",
			clusters: []envoy_common.Cluster{envoy_common.NewCluster(
				envoy_common.WithService("db"),
				envoy_common.WithWeight(200),
			)},
			maxConnectAttempts: 5,
			expected: `
            name: outbound:127.0.0.1:5432
            trafficDirection: OUTBOUND
            address:
              socketAddress:
                address: 127.0.0.1
                portValue: 5432
            filterChains:
            - filters:
              - name: envoy.filters.network.tcp_proxy
                typedConfig:
                  '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                  cluster: db
                  statPrefix: db
                  maxConnectAttempts: 5
`,
		}),
	)
})
