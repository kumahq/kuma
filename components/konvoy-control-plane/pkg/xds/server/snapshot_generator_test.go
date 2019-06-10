package server

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	konvoy_mesh "github.com/Kong/konvoy/components/konvoy-control-plane/model/api/v1alpha1"
	util_cache "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/cache"
	util_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/proto"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/generator"
)

var _ = Describe("Reconcile", func() {
	Describe("templateSnapshotGenerator", func() {

		gen := templateSnapshotGenerator{
			Template: &konvoy_mesh.ProxyTemplate{
				Spec: konvoy_mesh.ProxyTemplateSpec{
					Sources: []konvoy_mesh.ProxyTemplateSource{
						{
							Profile: &konvoy_mesh.ProxyTemplateProfileSource{
								Name: "transparent-inbound-proxy",
							},
						},
						{
							Profile: &konvoy_mesh.ProxyTemplateProfileSource{
								Name: "transparent-outbound-proxy",
							},
						},
					},
				},
			},
		}

		type testCase struct {
			proxy    *generator.Proxy
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
				proxy: &generator.Proxy{
					Id: generator.ProxyId{Name: "side-car", Namespace: "default"},
					Workload: generator.Workload{
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
			Entry("should support Nodes with IP(s) but without Port(s)", testCase{
				proxy: &generator.Proxy{
					Id: generator.ProxyId{Name: "side-car", Namespace: "default"},
					Workload: generator.Workload{
						Version:   "2",
						Addresses: []string{"192.168.0.1"},
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
                  version: "2"
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
                  version: "2"
`,
			}),
			Entry("should support Nodes with 1 IP and 1 Port", testCase{
				proxy: &generator.Proxy{
					Id: generator.ProxyId{Name: "side-car", Namespace: "default"},
					Workload: generator.Workload{
						Version:   "3",
						Addresses: []string{"192.168.0.1"},
						Ports:     []uint32{8080},
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
				proxy: &generator.Proxy{
					Id: generator.ProxyId{Name: "side-car", Namespace: "default"},
					Workload: generator.Workload{
						Version:   "4",
						Addresses: []string{"192.168.0.1"},
						Ports:     []uint32{8080, 8443},
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
				proxy: &generator.Proxy{
					Id: generator.ProxyId{Name: "side-car", Namespace: "default"},
					Workload: generator.Workload{
						Version:   "5",
						Addresses: []string{"192.168.0.1", "192.168.0.2"},
						Ports:     []uint32{8080, 8443},
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
