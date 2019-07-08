package server

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	core_discovery "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/discovery"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/memory"
	util_cache "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/cache"
	util_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/proto"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/model"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/template"
)

var _ = Describe("Reconcile", func() {
	Describe("templateSnapshotGenerator", func() {

		gen := templateSnapshotGenerator{
			ProxyTemplateResolver: &simpleProxyTemplateResolver{
				ResourceStore:        memory.NewStore(),
				DefaultProxyTemplate: template.TransparentProxyTemplate,
			},
		}

		type testCase struct {
			proxy    *model.Proxy
			expected string
		}

		DescribeTable("Generate Snapshot per Envoy Node",
			func(given testCase) {
				// when
				s, err := gen.GenerateSnapshot(given.proxy)

				// then
				Expect(err).ToNot(HaveOccurred())

				// then
				resp := util_cache.ToDeltaDiscoveryResponse(s)
				actual, err := util_proto.ToYAML(resp)

				Expect(err).ToNot(HaveOccurred())
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("should support Nodes without metadata", testCase{
				proxy: &model.Proxy{
					Id: model.ProxyId{Name: "side-car", Namespace: "default"},
					Workload: model.Workload{
						Version: "1",
					},
				},
				expected: `
                resources:
                - name: pass_through
                  resource:
                    '@type': type.googleapis.com/envoy.api.v2.Cluster
                    connectTimeout: 5s
                    lbPolicy: ORIGINAL_DST_LB
                    name: pass_through
                    type: ORIGINAL_DST
                  version: "1"
                - name: catch_all
                  resource:
                    '@type': type.googleapis.com/envoy.api.v2.Listener
                    address:
                      socketAddress:
                        address: 0.0.0.0
                        portValue: 15001
                    filterChains:
                    - filters:
                      - name: envoy.tcp_proxy
                        typedConfig:
                          '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                          cluster: pass_through
                          statPrefix: pass_through
                    name: catch_all
                    useOriginalDst: true
                  version: "1"
`,
			}),
			Entry("should support Nodes with 1 IP and 1 Port", testCase{
				proxy: &model.Proxy{
					Id: model.ProxyId{Name: "side-car", Namespace: "default"},
					Workload: model.Workload{
						Version: "3",
						Endpoints: []core_discovery.WorkloadEndpoint{
							{Address: "192.168.0.1", Port: 8080},
						},
					},
				},
				expected: `
                resources:
                - name: localhost:8080
                  resource:
                    '@type': type.googleapis.com/envoy.api.v2.Cluster
                    connectTimeout: 5s
                    loadAssignment:
                      clusterName: localhost:8080
                      endpoints:
                      - lbEndpoints:
                        - endpoint:
                            address:
                              socketAddress:
                                address: 127.0.0.1
                                portValue: 8080
                    name: localhost:8080
                    type: STATIC
                  version: "3"
                - name: pass_through
                  resource:
                    '@type': type.googleapis.com/envoy.api.v2.Cluster
                    connectTimeout: 5s
                    lbPolicy: ORIGINAL_DST_LB
                    name: pass_through
                    type: ORIGINAL_DST
                  version: "3"
                - name: catch_all
                  resource:
                    '@type': type.googleapis.com/envoy.api.v2.Listener
                    address:
                      socketAddress:
                        address: 0.0.0.0
                        portValue: 15001
                    filterChains:
                    - filters:
                      - name: envoy.tcp_proxy
                        typedConfig:
                          '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                          cluster: pass_through
                          statPrefix: pass_through
                    name: catch_all
                    useOriginalDst: true
                  version: "3"
                - name: inbound:192.168.0.1:8080
                  resource:
                    '@type': type.googleapis.com/envoy.api.v2.Listener
                    address:
                      socketAddress:
                        address: 192.168.0.1
                        portValue: 8080
                    deprecatedV1:
                      bindToPort: false
                    filterChains:
                    - filters:
                      - name: envoy.tcp_proxy
                        typedConfig:
                          '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                          cluster: localhost:8080
                          statPrefix: localhost:8080
                    name: inbound:192.168.0.1:8080
                  version: "3"
`,
			}),
			Entry("should support Nodes with 1 IP and N Port(s)", testCase{
				proxy: &model.Proxy{
					Id: model.ProxyId{Name: "side-car", Namespace: "default"},
					Workload: model.Workload{
						Version: "4",
						Endpoints: []core_discovery.WorkloadEndpoint{
							{Address: "192.168.0.1", Port: 8080},
							{Address: "192.168.0.1", Port: 8443},
						},
					},
				},
				expected: `
                resources:
                - name: localhost:8080
                  resource:
                    '@type': type.googleapis.com/envoy.api.v2.Cluster
                    connectTimeout: 5s
                    loadAssignment:
                      clusterName: localhost:8080
                      endpoints:
                      - lbEndpoints:
                        - endpoint:
                            address:
                              socketAddress:
                                address: 127.0.0.1
                                portValue: 8080
                    name: localhost:8080
                    type: STATIC
                  version: "4"
                - name: localhost:8443
                  resource:
                    '@type': type.googleapis.com/envoy.api.v2.Cluster
                    connectTimeout: 5s
                    loadAssignment:
                      clusterName: localhost:8443
                      endpoints:
                      - lbEndpoints:
                        - endpoint:
                            address:
                              socketAddress:
                                address: 127.0.0.1
                                portValue: 8443
                    name: localhost:8443
                    type: STATIC
                  version: "4"
                - name: pass_through
                  resource:
                    '@type': type.googleapis.com/envoy.api.v2.Cluster
                    connectTimeout: 5s
                    lbPolicy: ORIGINAL_DST_LB
                    name: pass_through
                    type: ORIGINAL_DST
                  version: "4"
                - name: catch_all
                  resource:
                    '@type': type.googleapis.com/envoy.api.v2.Listener
                    address:
                      socketAddress:
                        address: 0.0.0.0
                        portValue: 15001
                    filterChains:
                    - filters:
                      - name: envoy.tcp_proxy
                        typedConfig:
                          '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                          cluster: pass_through
                          statPrefix: pass_through
                    name: catch_all
                    useOriginalDst: true
                  version: "4"
                - name: inbound:192.168.0.1:8080
                  resource:
                    '@type': type.googleapis.com/envoy.api.v2.Listener
                    address:
                      socketAddress:
                        address: 192.168.0.1
                        portValue: 8080
                    deprecatedV1:
                      bindToPort: false
                    filterChains:
                    - filters:
                      - name: envoy.tcp_proxy
                        typedConfig:
                          '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                          cluster: localhost:8080
                          statPrefix: localhost:8080
                    name: inbound:192.168.0.1:8080
                  version: "4"
                - name: inbound:192.168.0.1:8443
                  resource:
                    '@type': type.googleapis.com/envoy.api.v2.Listener
                    address:
                      socketAddress:
                        address: 192.168.0.1
                        portValue: 8443
                    deprecatedV1:
                      bindToPort: false
                    filterChains:
                    - filters:
                      - name: envoy.tcp_proxy
                        typedConfig:
                          '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                          cluster: localhost:8443
                          statPrefix: localhost:8443
                    name: inbound:192.168.0.1:8443
                  version: "4"
`,
			}),
			Entry("should support Nodes with M IP(s) and N Port(s)", testCase{
				proxy: &model.Proxy{
					Id: model.ProxyId{Name: "side-car", Namespace: "default"},
					Workload: model.Workload{
						Version: "5",
						Endpoints: []core_discovery.WorkloadEndpoint{
							{Address: "192.168.0.1", Port: 8080},
							{Address: "192.168.0.1", Port: 8443},
							{Address: "192.168.0.2", Port: 8080},
							{Address: "192.168.0.2", Port: 8443},
						},
					},
				},
				expected: `
                resources:
                - name: localhost:8080
                  resource:
                    '@type': type.googleapis.com/envoy.api.v2.Cluster
                    connectTimeout: 5s
                    loadAssignment:
                      clusterName: localhost:8080
                      endpoints:
                      - lbEndpoints:
                        - endpoint:
                            address:
                              socketAddress:
                                address: 127.0.0.1
                                portValue: 8080
                    name: localhost:8080
                    type: STATIC
                  version: "5"
                - name: localhost:8443
                  resource:
                    '@type': type.googleapis.com/envoy.api.v2.Cluster
                    connectTimeout: 5s
                    loadAssignment:
                      clusterName: localhost:8443
                      endpoints:
                      - lbEndpoints:
                        - endpoint:
                            address:
                              socketAddress:
                                address: 127.0.0.1
                                portValue: 8443
                    name: localhost:8443
                    type: STATIC
                  version: "5"
                - name: pass_through
                  resource:
                    '@type': type.googleapis.com/envoy.api.v2.Cluster
                    connectTimeout: 5s
                    lbPolicy: ORIGINAL_DST_LB
                    name: pass_through
                    type: ORIGINAL_DST
                  version: "5"
                - name: catch_all
                  resource:
                    '@type': type.googleapis.com/envoy.api.v2.Listener
                    address:
                      socketAddress:
                        address: 0.0.0.0
                        portValue: 15001
                    filterChains:
                    - filters:
                      - name: envoy.tcp_proxy
                        typedConfig:
                          '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                          cluster: pass_through
                          statPrefix: pass_through
                    name: catch_all
                    useOriginalDst: true
                  version: "5"
                - name: inbound:192.168.0.1:8080
                  resource:
                    '@type': type.googleapis.com/envoy.api.v2.Listener
                    address:
                      socketAddress:
                        address: 192.168.0.1
                        portValue: 8080
                    deprecatedV1:
                      bindToPort: false
                    filterChains:
                    - filters:
                      - name: envoy.tcp_proxy
                        typedConfig:
                          '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                          cluster: localhost:8080
                          statPrefix: localhost:8080
                    name: inbound:192.168.0.1:8080
                  version: "5"
                - name: inbound:192.168.0.1:8443
                  resource:
                    '@type': type.googleapis.com/envoy.api.v2.Listener
                    address:
                      socketAddress:
                        address: 192.168.0.1
                        portValue: 8443
                    deprecatedV1:
                      bindToPort: false
                    filterChains:
                    - filters:
                      - name: envoy.tcp_proxy
                        typedConfig:
                          '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                          cluster: localhost:8443
                          statPrefix: localhost:8443
                    name: inbound:192.168.0.1:8443
                  version: "5"
                - name: inbound:192.168.0.2:8080
                  resource:
                    '@type': type.googleapis.com/envoy.api.v2.Listener
                    address:
                      socketAddress:
                        address: 192.168.0.2
                        portValue: 8080
                    deprecatedV1:
                      bindToPort: false
                    filterChains:
                    - filters:
                      - name: envoy.tcp_proxy
                        typedConfig:
                          '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                          cluster: localhost:8080
                          statPrefix: localhost:8080
                    name: inbound:192.168.0.2:8080
                  version: "5"
                - name: inbound:192.168.0.2:8443
                  resource:
                    '@type': type.googleapis.com/envoy.api.v2.Listener
                    address:
                      socketAddress:
                        address: 192.168.0.2
                        portValue: 8443
                    deprecatedV1:
                      bindToPort: false
                    filterChains:
                    - filters:
                      - name: envoy.tcp_proxy
                        typedConfig:
                          '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                          cluster: localhost:8443
                          statPrefix: localhost:8443
                    name: inbound:192.168.0.2:8443
                  version: "5"
`,
			}),
		)
	})
})
