package listeners_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/xds/envoy/listeners"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"

	test_model "github.com/Kong/kuma/pkg/test/resources/model"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
	envoy_common "github.com/Kong/kuma/pkg/xds/envoy"
)

var _ = Describe("NetworkRbacConfigurer", func() {

	type testCase struct {
		listenerName    string
		listenerAddress string
		listenerPort    uint32
		statsName       string
		clusters        []envoy_common.ClusterSubset
		rbacEnabled     bool
		permission      *mesh_core.TrafficPermissionResource
		expected        string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			listener, err := NewListenerBuilder().
				Configure(InboundListener(given.listenerName, given.listenerAddress, given.listenerPort)).
				Configure(FilterChain(NewFilterChainBuilder().
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
			clusters:        []envoy_common.ClusterSubset{{ClusterName: "localhost:8080", Weight: 200}},
			rbacEnabled:     true,
			permission: &mesh_core.TrafficPermissionResource{
				Meta: &test_model.ResourceMeta{
					Name: "tp-1",
					Mesh: "default",
				},
				Spec: mesh_proto.TrafficPermission{
					Sources: []*mesh_proto.Selector{
						{
							Match: map[string]string{
								"service": "web1",
								"version": "1.0",
							},
						},
					},
					Destinations: []*mesh_proto.Selector{
						{
							Match: map[string]string{
								"service": "backend1",
								"env":     "dev",
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
            filterChains:
            - filters:
              - name: envoy.filters.network.rbac
                typedConfig:
                  '@type': type.googleapis.com/envoy.config.filter.network.rbac.v2.RBAC
                  rules:
                    policies:
                      tp-1:
                        permissions:
                        - any: true
                        principals:
                        - authenticated:
                            principalName:
                              exact: spiffe://default/web1
                  statPrefix: inbound_192_168_0_1_8080.
              - name: envoy.tcp_proxy
                typedConfig:
                  '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                  cluster: localhost:8080
                  statPrefix: localhost_8080
`,
		}),
		Entry("basic tcp_proxy with network RBAC disabled", testCase{
			listenerName:    "inbound:192.168.0.1:8080",
			listenerAddress: "192.168.0.1",
			listenerPort:    8080,
			statsName:       "localhost:8080",
			clusters:        []envoy_common.ClusterSubset{{ClusterName: "localhost:8080", Weight: 200}},
			rbacEnabled:     false,
			permission: &mesh_core.TrafficPermissionResource{
				Meta: &test_model.ResourceMeta{
					Name: "tp-1",
					Mesh: "default",
				},
				Spec: mesh_proto.TrafficPermission{
					Sources: []*mesh_proto.Selector{
						{
							Match: map[string]string{
								"service": "web1",
								"version": "1.0",
							},
						},
					},
					Destinations: []*mesh_proto.Selector{
						{
							Match: map[string]string{
								"service": "backend1",
								"env":     "dev",
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
            filterChains:
            - filters:
              - name: envoy.tcp_proxy
                typedConfig:
                  '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                  cluster: localhost:8080
                  statPrefix: localhost_8080
`,
		}),
	)
})
