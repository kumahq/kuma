package server

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/gogo/protobuf/types"

	util_cache "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/cache"
	util_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/proto"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
)

var _ = Describe("Reconcile", func() {
	Describe("basicSnapshotGenerator", func() {

		generator := basicSnapshotGenerator{}

		type testCase struct {
			node     *core.Node
			expected string
		}

		DescribeTable("Generate Snapshot per Envoy Node",
			func(given testCase) {
				// when
				s := generator.GenerateSnapshot(given.node)

				// then
				resp := util_cache.ToDeltaDiscoveryResponse(s)
				actual, err := util_proto.ToYAML(resp)

				Expect(err).ToNot(HaveOccurred())
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("should support Nodes without metadata", testCase{
				node: &core.Node{
					Id:      "side-car",
					Cluster: "example",
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
                  version: bcb116e377ade95222292def4228a8427bf9505e
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
                  version: bcb116e377ade95222292def4228a8427bf9505e
`,
			}),
			Entry("should support Nodes with IP(s) but without Port(s)", testCase{
				node: &core.Node{
					Id:      "side-car",
					Cluster: "example",
					Metadata: &types.Struct{
						Fields: map[string]*types.Value{
							"IPS": &types.Value{
								Kind: &types.Value_StringValue{
									StringValue: "192.168.0.1",
								},
							},
						},
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
                  version: da242e6d17542a6994df846e46d7911a9dd7347e
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
                  version: da242e6d17542a6994df846e46d7911a9dd7347e
`,
			}),
			Entry("should support Nodes with 1 IP and 1 Port", testCase{
				node: &core.Node{
					Id:      "side-car",
					Cluster: "example",
					Metadata: &types.Struct{
						Fields: map[string]*types.Value{
							"IPS": &types.Value{
								Kind: &types.Value_StringValue{
									StringValue: "192.168.0.1",
								},
							},
							"PORTS": &types.Value{
								Kind: &types.Value_StringValue{
									StringValue: "8080",
								},
							},
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
                  version: 5eca733d9d49e8b9d8bb2f40ad900b9e16f7715c
                - name: pass_through
                  resource:
                    '@type': type.googleapis.com/envoy.api.v2.Cluster
                    connectTimeout: 5s
                    lbPolicy: ORIGINAL_DST_LB
                    name: pass_through
                    type: ORIGINAL_DST
                  version: 5eca733d9d49e8b9d8bb2f40ad900b9e16f7715c
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
                  version: 5eca733d9d49e8b9d8bb2f40ad900b9e16f7715c
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
                  version: 5eca733d9d49e8b9d8bb2f40ad900b9e16f7715c
`,
			}),
			Entry("should support Nodes with 1 IP and N Port(s)", testCase{
				node: &core.Node{
					Id:      "side-car",
					Cluster: "example",
					Metadata: &types.Struct{
						Fields: map[string]*types.Value{
							"IPS": &types.Value{
								Kind: &types.Value_StringValue{
									StringValue: "192.168.0.1",
								},
							},
							"PORTS": &types.Value{
								Kind: &types.Value_StringValue{
									StringValue: "8080,8443",
								},
							},
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
                  version: a578863a36c4751c656c33fab277cbe0a1cc8937
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
                  version: a578863a36c4751c656c33fab277cbe0a1cc8937
                - name: pass_through
                  resource:
                    '@type': type.googleapis.com/envoy.api.v2.Cluster
                    connectTimeout: 5s
                    lbPolicy: ORIGINAL_DST_LB
                    name: pass_through
                    type: ORIGINAL_DST
                  version: a578863a36c4751c656c33fab277cbe0a1cc8937
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
                  version: a578863a36c4751c656c33fab277cbe0a1cc8937
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
                  version: a578863a36c4751c656c33fab277cbe0a1cc8937
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
                  version: a578863a36c4751c656c33fab277cbe0a1cc8937
`,
			}),
			Entry("should support Nodes with M IP(s) and N Port(s)", testCase{
				node: &core.Node{
					Id:      "side-car",
					Cluster: "example",
					Metadata: &types.Struct{
						Fields: map[string]*types.Value{
							"IPS": &types.Value{
								Kind: &types.Value_StringValue{
									StringValue: "192.168.0.1,192.168.0.2",
								},
							},
							"PORTS": &types.Value{
								Kind: &types.Value_StringValue{
									StringValue: "8080,8443",
								},
							},
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
                  version: c79ceb9cc3a57ca2ee19275af9efb31f468e543f
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
                  version: c79ceb9cc3a57ca2ee19275af9efb31f468e543f
                - name: pass_through
                  resource:
                    '@type': type.googleapis.com/envoy.api.v2.Cluster
                    connectTimeout: 5s
                    lbPolicy: ORIGINAL_DST_LB
                    name: pass_through
                    type: ORIGINAL_DST
                  version: c79ceb9cc3a57ca2ee19275af9efb31f468e543f
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
                  version: c79ceb9cc3a57ca2ee19275af9efb31f468e543f
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
                  version: c79ceb9cc3a57ca2ee19275af9efb31f468e543f
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
                  version: c79ceb9cc3a57ca2ee19275af9efb31f468e543f
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
                  version: c79ceb9cc3a57ca2ee19275af9efb31f468e543f
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
                  version: c79ceb9cc3a57ca2ee19275af9efb31f468e543f
`,
			}),
			Entry("should tolerate Node metadata with invalid IP(s)", testCase{
				node: &core.Node{
					Id:      "side-car",
					Cluster: "example",
					Metadata: &types.Struct{
						Fields: map[string]*types.Value{
							"IPS": &types.Value{
								Kind: &types.Value_StringValue{
									StringValue: ",192.168.0.1,,",
								},
							},
							"PORTS": &types.Value{
								Kind: &types.Value_StringValue{
									StringValue: "8080",
								},
							},
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
                  version: 6edf23d326642535319b127e8e21ec39773a6305
                - name: pass_through
                  resource:
                    '@type': type.googleapis.com/envoy.api.v2.Cluster
                    connectTimeout: 5s
                    lbPolicy: ORIGINAL_DST_LB
                    name: pass_through
                    type: ORIGINAL_DST
                  version: 6edf23d326642535319b127e8e21ec39773a6305
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
                  version: 6edf23d326642535319b127e8e21ec39773a6305
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
                  version: 6edf23d326642535319b127e8e21ec39773a6305
`,
			}),
			Entry("should tolerate Node metadata with invalid Port(s)", testCase{
				node: &core.Node{
					Id:      "side-car",
					Cluster: "example",
					Metadata: &types.Struct{
						Fields: map[string]*types.Value{
							"IPS": &types.Value{
								Kind: &types.Value_StringValue{
									StringValue: "192.168.0.1",
								},
							},
							"PORTS": &types.Value{
								Kind: &types.Value_StringValue{
									StringValue: ",8080,b",
								},
							},
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
                  version: 3ead480b0dec2d246e0e639f647219bbfc737b08
                - name: pass_through
                  resource:
                    '@type': type.googleapis.com/envoy.api.v2.Cluster
                    connectTimeout: 5s
                    lbPolicy: ORIGINAL_DST_LB
                    name: pass_through
                    type: ORIGINAL_DST
                  version: 3ead480b0dec2d246e0e639f647219bbfc737b08
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
                  version: 3ead480b0dec2d246e0e639f647219bbfc737b08
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
                  version: 3ead480b0dec2d246e0e639f647219bbfc737b08
`,
			}),
		)
	})
})
