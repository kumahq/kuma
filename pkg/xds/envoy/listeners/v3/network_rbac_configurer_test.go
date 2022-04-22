package v3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/xds"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
)

var _ = Describe("NetworkRbacConfigurer", func() {

	type testCase struct {
		listenerName     string
		listenerProtocol xds.SocketAddressProtocol
		listenerAddress  string
		listenerPort     uint32
		statsName        string
		clusters         []envoy_common.Cluster
		rbacEnabled      bool
		permission       *core_mesh.TrafficPermissionResource
		expected         string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			listener, err := NewListenerBuilder(envoy_common.APIV3).
				Configure(InboundListener(given.listenerName, given.listenerAddress, given.listenerPort, given.listenerProtocol)).
				Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
					Configure(TcpProxy(given.statsName, given.clusters...)).
					Configure(NetworkRBAC(given.listenerName, given.rbacEnabled, given.permission)))).
				Build()
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(listener)
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("basic tcp_proxy with network RBAC enabled", testCase{
			listenerName:    "inbound:192.168.0.1:8080",
			listenerAddress: "192.168.0.1",
			listenerPort:    8080,
			statsName:       "localhost:8080",
			clusters: []envoy_common.Cluster{envoy_common.NewCluster(
				envoy_common.WithService("localhost:8080"),
				envoy_common.WithWeight(200),
			)},
			rbacEnabled: true,
			permission: &core_mesh.TrafficPermissionResource{
				Meta: &test_model.ResourceMeta{
					Name: "tp-1",
					Mesh: "default",
				},
				Spec: &mesh_proto.TrafficPermission{
					Sources: []*mesh_proto.Selector{
						{
							Match: map[string]string{
								"kuma.io/service": "web1",
								"version":         "1.0",
							},
						},
					},
					Destinations: []*mesh_proto.Selector{
						{
							Match: map[string]string{
								"kuma.io/service": "backend1",
								"env":             "dev",
							},
						},
					},
				},
			},
			expected: `
            address:
              socketAddress:
                address: 192.168.0.1
                portValue: 8080
            enableReusePort: false
            filterChains:
            - filters:
              - name: envoy.filters.network.rbac
                typedConfig:
                  '@type': type.googleapis.com/envoy.extensions.filters.network.rbac.v3.RBAC
                  rules:
                    policies:
                      tp-1:
                        permissions:
                        - any: true
                        principals:
                        - andIds:
                            ids:
                            - authenticated:
                                principalName:
                                  exact: kuma://version/1.0
                            - authenticated:
                                principalName:
                                  exact: spiffe://default/web1
                  statPrefix: inbound_192_168_0_1_8080.
              - name: envoy.filters.network.tcp_proxy
                typedConfig:
                  '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                  cluster: localhost:8080
                  statPrefix: localhost_8080
            name: inbound:192.168.0.1:8080
            trafficDirection: INBOUND
`,
		}),
		Entry("basic tcp_proxy with network RBAC disabled", testCase{
			listenerName:    "inbound:192.168.0.1:8080",
			listenerAddress: "192.168.0.1",
			listenerPort:    8080,
			statsName:       "localhost:8080",
			clusters: []envoy_common.Cluster{envoy_common.NewCluster(
				envoy_common.WithService("localhost:8080"),
				envoy_common.WithWeight(200),
			)},
			rbacEnabled: false,
			permission: &core_mesh.TrafficPermissionResource{
				Meta: &test_model.ResourceMeta{
					Name: "tp-1",
					Mesh: "default",
				},
				Spec: &mesh_proto.TrafficPermission{
					Sources: []*mesh_proto.Selector{
						{
							Match: map[string]string{
								"kuma.io/service": "web1",
								"version":         "1.0",
							},
						},
					},
					Destinations: []*mesh_proto.Selector{
						{
							Match: map[string]string{
								"kuma.io/service": "backend1",
								"env":             "dev",
							},
						},
					},
				},
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
	)
})
