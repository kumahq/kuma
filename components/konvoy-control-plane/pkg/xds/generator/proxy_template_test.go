package generator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	konvoy_mesh "github.com/Kong/konvoy/components/konvoy-control-plane/model/api/v1alpha1"
	util_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/proto"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/generator"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/model"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/template"
)

var _ = Describe("Generator", func() {
	Describe("TransparentInboundProxyProfile", func() {

		type testCase struct {
			proxy    *model.Proxy
			expected string
		}

		DescribeTable("Generate Envoy xDS resources",
			func(given testCase) {
				// setup
				gen := &generator.TransparentInboundProxyProfile{}

				// when
				rs, err := gen.Generate(given.proxy)

				// then
				Expect(err).ToNot(HaveOccurred())

				// then
				resp := generator.ResourceList(rs).ToDeltaDiscoveryResponse()
				actual, err := util_proto.ToYAML(resp)

				Expect(err).ToNot(HaveOccurred())
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("should support Nodes without IP addresses and Ports", testCase{
				proxy: &model.Proxy{
					Id: model.ProxyId{Name: "side-car", Namespace: "default"},
					Workload: model.Workload{
						Version: "v1",
					},
				},
				expected: `{}`,
			}),
			Entry("should support Nodes with IP addresses but without Ports", testCase{
				proxy: &model.Proxy{
					Id: model.ProxyId{Name: "side-car", Namespace: "default"},
					Workload: model.Workload{
						Version:   "v1",
						Addresses: []string{"192.168.0.1"},
					},
				},
				expected: `{}`,
			}),
			Entry("should support Nodes with 1 IP address and 1 Port", testCase{
				proxy: &model.Proxy{
					Id: model.ProxyId{Name: "side-car", Namespace: "default"},
					Workload: model.Workload{
						Version:   "v1",
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
          version: v1
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
          version: v1
`,
			}),
			Entry("should support Nodes with 1 IP address and 2 Ports", testCase{
				proxy: &model.Proxy{
					Id: model.ProxyId{Name: "side-car", Namespace: "default"},
					Workload: model.Workload{
						Version:   "v1",
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
            version: v1
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
            version: v1
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
            version: v1
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
            version: v1
`,
			}),
			Entry("should support Nodes with 2 IP addresses and 2 Ports", testCase{
				proxy: &model.Proxy{
					Id: model.ProxyId{Name: "side-car", Namespace: "default"},
					Workload: model.Workload{
						Version:   "v1",
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
          version: v1
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
          version: v1
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
          version: v1
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
          version: v1
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
          version: v1
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
          version: v1
`,
			}),
		)
	})

	Describe("TransparentOutboundProxyProfile", func() {

		type testCase struct {
			proxy    *model.Proxy
			expected string
		}

		DescribeTable("Generate Envoy xDS resources",
			func(given testCase) {
				// setup
				gen := &generator.TransparentOutboundProxyProfile{}

				// when
				rs, err := gen.Generate(given.proxy)

				// then
				Expect(err).ToNot(HaveOccurred())

				// then
				resp := generator.ResourceList(rs).ToDeltaDiscoveryResponse()
				actual, err := util_proto.ToYAML(resp)

				Expect(err).ToNot(HaveOccurred())
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("should support Nodes without IP addresses and ports", testCase{
				proxy: &model.Proxy{
					Id: model.ProxyId{Name: "side-car", Namespace: "default"},
					Workload: model.Workload{
						Version: "v1",
					},
				},
				expected: `
        resources:
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
          version: v1
        - name: pass_through
          resource:
            '@type': type.googleapis.com/envoy.api.v2.Cluster
            connectTimeout: 5s
            lbPolicy: ORIGINAL_DST_LB
            name: pass_through
            type: ORIGINAL_DST
          version: v1
`,
			}),
			Entry("should support Nodes with 1 IP address and 1 Port", testCase{
				proxy: &model.Proxy{
					Id: model.ProxyId{Name: "side-car", Namespace: "default"},
					Workload: model.Workload{
						Version:   "v1",
						Addresses: []string{"192.168.0.1"},
						Ports:     []uint32{8080},
					},
				},
				expected: `
        resources:
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
          version: v1
        - name: pass_through
          resource:
            '@type': type.googleapis.com/envoy.api.v2.Cluster
            connectTimeout: 5s
            lbPolicy: ORIGINAL_DST_LB
            name: pass_through
            type: ORIGINAL_DST
          version: v1
`,
			}),
			Entry("should support Nodes with 2 IP addresses and 2 Ports", testCase{
				proxy: &model.Proxy{
					Id: model.ProxyId{Name: "side-car", Namespace: "default"},
					Workload: model.Workload{
						Version:   "v1",
						Addresses: []string{"192.168.0.1", "192.168.0.2"},
						Ports:     []uint32{8080, 8443},
					},
				},
				expected: `
        resources:
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
          version: v1
        - name: pass_through
          resource:
            '@type': type.googleapis.com/envoy.api.v2.Cluster
            connectTimeout: 5s
            lbPolicy: ORIGINAL_DST_LB
            name: pass_through
            type: ORIGINAL_DST
          version: v1
`,
			}),
		)
	})

	Describe("ProxyTemplateProfileSource", func() {

		type testCase struct {
			proxy    *model.Proxy
			profile  *konvoy_mesh.ProxyTemplateProfileSource
			expected string
		}

		DescribeTable("Generate Envoy xDS resources",
			func(given testCase) {
				// setup
				gen := &generator.ProxyTemplateProfileSource{
					Profile: given.profile,
				}

				// when
				rs, err := gen.Generate(given.proxy)

				// then
				Expect(err).ToNot(HaveOccurred())

				// then
				resp := generator.ResourceList(rs).ToDeltaDiscoveryResponse()
				actual, err := util_proto.ToYAML(resp)

				Expect(err).ToNot(HaveOccurred())
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("should support pre-defined `transparent-inbound-proxy` profile", testCase{
				proxy: &model.Proxy{
					Id: model.ProxyId{Name: "side-car", Namespace: "default"},
					Workload: model.Workload{
						Version:   "v1",
						Addresses: []string{"192.168.0.1"},
						Ports:     []uint32{8080},
					},
				},
				profile: &konvoy_mesh.ProxyTemplateProfileSource{
					Name: template.ProfileTransparentInboundProxy,
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
            version: v1
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
            version: v1
`,
			}),
			Entry("should support pre-defined `transparent-outbound-proxy` profile", testCase{
				proxy: &model.Proxy{
					Id: model.ProxyId{Name: "side-car", Namespace: "default"},
					Workload: model.Workload{
						Version:   "v1",
						Addresses: []string{"192.168.0.1"},
						Ports:     []uint32{8080},
					},
				},
				profile: &konvoy_mesh.ProxyTemplateProfileSource{
					Name: "transparent-outbound-proxy",
				},
				expected: `
        resources:
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
            version: v1
          - name: pass_through
            resource:
              '@type': type.googleapis.com/envoy.api.v2.Cluster
              connectTimeout: 5s
              lbPolicy: ORIGINAL_DST_LB
              name: pass_through
              type: ORIGINAL_DST
            version: v1
`,
			}),
		)
	})

	Describe("ProxyTemplateRawSource", func() {

		Context("Manually-defined xDS resources are not valid", func() {

			type testCase struct {
				proxy *model.Proxy
				raw   *konvoy_mesh.ProxyTemplateRawSource
				err   interface{}
			}

			DescribeTable("Avoid producing invalid Envoy xDS resources",
				func(given testCase) {
					// setup
					gen := &generator.ProxyTemplateRawSource{
						Raw: given.raw,
					}

					// when
					rs, err := gen.Generate(given.proxy)

					// then
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError(given.err))
					Expect(rs).To(BeNil())
				},
				Entry("should fail when `resource` field is empty", testCase{
					proxy: &model.Proxy{
						Id: model.ProxyId{Name: "side-car", Namespace: "default"},
						Workload: model.Workload{
							Version:   "v1",
							Addresses: []string{"192.168.0.1"},
							Ports:     []uint32{8080},
						},
					},
					raw: &konvoy_mesh.ProxyTemplateRawSource{
						Resources: []konvoy_mesh.ProxyTemplateRawResource{{
							Name:    "raw-name",
							Version: "raw-version",
							Resource: `
`,
						}},
					},
					err: "raw.resources[0]{name=\"raw-name\"}.resource: Any JSON doesn't have '@type'",
				}),
				Entry("should fail when `resource` field is neither a YAML nor a JSON", testCase{
					proxy: &model.Proxy{
						Id: model.ProxyId{Name: "side-car", Namespace: "default"},
						Workload: model.Workload{
							Version:   "v1",
							Addresses: []string{"192.168.0.1"},
							Ports:     []uint32{8080},
						},
					},
					raw: &konvoy_mesh.ProxyTemplateRawSource{
						Resources: []konvoy_mesh.ProxyTemplateRawResource{{
							Name:     "raw-name",
							Version:  "raw-version",
							Resource: `{`,
						}},
					},
					err: "raw.resources[0]{name=\"raw-name\"}.resource: unexpected EOF",
				}),
				Entry("should fail when `resource` field has unknown @type", testCase{
					proxy: &model.Proxy{
						Id: model.ProxyId{Name: "side-car", Namespace: "default"},
						Workload: model.Workload{
							Version:   "v1",
							Addresses: []string{"192.168.0.1"},
							Ports:     []uint32{8080},
						},
					},
					raw: &konvoy_mesh.ProxyTemplateRawSource{
						Resources: []konvoy_mesh.ProxyTemplateRawResource{{
							Name:    "raw-name",
							Version: "raw-version",
							Resource: `
              '@type': type.googleapis.com/unknown.Resource
`,
						}},
					},
					err: "raw.resources[0]{name=\"raw-name\"}.resource: unknown message type \"unknown.Resource\"",
				}),
				Entry("should fail when `resource` field is a YAML without '@type' field", testCase{
					proxy: &model.Proxy{
						Id: model.ProxyId{Name: "side-car", Namespace: "default"},
						Workload: model.Workload{
							Version:   "v1",
							Addresses: []string{"192.168.0.1"},
							Ports:     []uint32{8080},
						},
					},
					raw: &konvoy_mesh.ProxyTemplateRawSource{
						Resources: []konvoy_mesh.ProxyTemplateRawResource{{
							Name:    "raw-name",
							Version: "raw-version",
							Resource: `
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
`,
						}},
					},
					err: "raw.resources[0]{name=\"raw-name\"}.resource: Any JSON doesn't have '@type'",
				}),
				Entry("should fail when `resource` field is an invalid xDS resource", testCase{
					proxy: &model.Proxy{
						Id: model.ProxyId{Name: "side-car", Namespace: "default"},
						Workload: model.Workload{
							Version:   "v1",
							Addresses: []string{"192.168.0.1"},
							Ports:     []uint32{8080},
						},
					},
					raw: &konvoy_mesh.ProxyTemplateRawSource{
						Resources: []konvoy_mesh.ProxyTemplateRawResource{{
							Name:    "raw-name",
							Version: "raw-version",
							Resource: `
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
              type: STATIC
`,
						}},
					},
					err: "raw.resources[0]{name=\"raw-name\"}.resource: invalid Cluster.Name: value length must be at least 1 bytes",
				}),
			)
		})

		Context("Manually-defined xDS resources are valid", func() {

			type testCase struct {
				proxy    *model.Proxy
				raw      *konvoy_mesh.ProxyTemplateRawSource
				expected string
			}

			DescribeTable("Generate Envoy xDS resources", func(given testCase) {
				// setup
				gen := &generator.ProxyTemplateRawSource{
					Raw: given.raw,
				}

				// when
				rs, err := gen.Generate(given.proxy)

				// then
				Expect(err).ToNot(HaveOccurred())

				// then
				resp := generator.ResourceList(rs).ToDeltaDiscoveryResponse()
				actual, err := util_proto.ToYAML(resp)

				Expect(err).ToNot(HaveOccurred())
				Expect(actual).To(MatchYAML(given.expected))
			},
				Entry("should support empty resource list", testCase{
					proxy: &model.Proxy{
						Id: model.ProxyId{Name: "side-car", Namespace: "default"},
						Workload: model.Workload{
							Version:   "v1",
							Addresses: []string{"192.168.0.1"},
							Ports:     []uint32{8080},
						},
					},
					raw: &konvoy_mesh.ProxyTemplateRawSource{
						Resources: nil,
					},
					expected: `{}`,
				}),
				Entry("should support Listener resource as YAML", testCase{
					proxy: &model.Proxy{
						Id: model.ProxyId{Name: "side-car", Namespace: "default"},
						Workload: model.Workload{
							Version:   "v1",
							Addresses: []string{"192.168.0.1"},
							Ports:     []uint32{8080},
						},
					},
					raw: &konvoy_mesh.ProxyTemplateRawSource{
						Resources: []konvoy_mesh.ProxyTemplateRawResource{{
							Name:    "raw-name",
							Version: "raw-version",
							Resource: `
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
`,
						}},
					},
					expected: `
          resources:
            - name: raw-name
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
              version: raw-version
`,
				}),
				Entry("should support Cluster resource as YAML", testCase{
					proxy: &model.Proxy{
						Id: model.ProxyId{Name: "side-car", Namespace: "default"},
						Workload: model.Workload{
							Version:   "v1",
							Addresses: []string{"192.168.0.1"},
							Ports:     []uint32{8080},
						},
					},
					raw: &konvoy_mesh.ProxyTemplateRawSource{
						Resources: []konvoy_mesh.ProxyTemplateRawResource{{
							Name:    "raw-name",
							Version: "raw-version",
							Resource: `
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
`,
						}},
					},
					expected: `
          resources:
            - name: raw-name
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
              version: raw-version
`,
				}),
				Entry("should support Cluster resource as JSON", testCase{
					proxy: &model.Proxy{
						Id: model.ProxyId{Name: "side-car", Namespace: "default"},
						Workload: model.Workload{
							Version:   "v1",
							Addresses: []string{"192.168.0.1"},
							Ports:     []uint32{8080},
						},
					},
					raw: &konvoy_mesh.ProxyTemplateRawSource{
						Resources: []konvoy_mesh.ProxyTemplateRawResource{{
							Name:    "raw-name",
							Version: "raw-version",
							Resource: `
              {
                "@type": "type.googleapis.com/envoy.api.v2.Cluster",
                "connectTimeout": "5s",
                "loadAssignment": {
                  "clusterName": "localhost:8080",
                  "endpoints": [
                    {
                      "lbEndpoints": [
                        {
                          "endpoint": {
                            "address": {
                              "socketAddress": {
                                "address": "127.0.0.1",
                                "portValue": 8080
                              }
                            }
                          }
                        }
                      ]
                    }
                  ]
                },
                "name": "localhost:8080",
                "type": "STATIC"
              }
`,
						}},
					},
					expected: `
          resources:
            - name: raw-name
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
              version: raw-version
`,
				}),
			)
		})
	})

	Describe("TemplateProxyGenerator", func() {
		Context("Error case", func() {
			type testCase struct {
				proxy    *model.Proxy
				template *konvoy_mesh.ProxyTemplate
				err      interface{}
			}

			DescribeTable("Avoid producing invalid Envoy xDS resources",
				func(given testCase) {
					// setup
					gen := &generator.TemplateProxyGenerator{
						ProxyTemplate: given.template,
					}

					// when
					rs, err := gen.Generate(given.proxy)

					// then
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError(given.err))
					Expect(rs).To(BeNil())
				},
				Entry("should fail when raw xDS resource is not valid", testCase{
					proxy: &model.Proxy{
						Id: model.ProxyId{Name: "side-car", Namespace: "default"},
						Workload: model.Workload{
							Version:   "v1",
							Addresses: []string{"192.168.0.1"},
							Ports:     []uint32{8080},
						},
					},
					template: &konvoy_mesh.ProxyTemplate{
						Spec: konvoy_mesh.ProxyTemplateSpec{
							Sources: []konvoy_mesh.ProxyTemplateSource{
								{
									Profile: &konvoy_mesh.ProxyTemplateProfileSource{
										Name: template.ProfileTransparentOutboundProxy,
									},
								},
								{
									Raw: &konvoy_mesh.ProxyTemplateRawSource{
										Resources: []konvoy_mesh.ProxyTemplateRawResource{{
											Name:     "raw-name",
											Version:  "raw-version",
											Resource: `{`,
										}},
									},
								},
							},
						},
					},
					err: "sources[1]{name=\"\"}: raw.resources[0]{name=\"raw-name\"}.resource: unexpected EOF",
				}),
			)
		})

		Context("Happy case", func() {

			type testCase struct {
				proxy    *model.Proxy
				template *konvoy_mesh.ProxyTemplate
				expected string
			}

			DescribeTable("Generate Envoy xDS resources",
				func(given testCase) {
					// setup
					gen := &generator.TemplateProxyGenerator{
						ProxyTemplate: given.template,
					}

					// when
					rs, err := gen.Generate(given.proxy)

					// then
					Expect(err).ToNot(HaveOccurred())

					// then
					resp := generator.ResourceList(rs).ToDeltaDiscoveryResponse()
					actual, err := util_proto.ToYAML(resp)

					Expect(err).ToNot(HaveOccurred())
					Expect(actual).To(MatchYAML(given.expected))
				},
				Entry("should support a combination of pre-defined profiles and raw xDS resources", testCase{
					proxy: &model.Proxy{
						Id: model.ProxyId{Name: "side-car", Namespace: "default"},
						Workload: model.Workload{
							Version:   "v1",
							Addresses: []string{"192.168.0.1"},
							Ports:     []uint32{8080},
						},
					},
					template: &konvoy_mesh.ProxyTemplate{
						Spec: konvoy_mesh.ProxyTemplateSpec{
							Sources: []konvoy_mesh.ProxyTemplateSource{
								{
									Profile: &konvoy_mesh.ProxyTemplateProfileSource{
										Name: template.ProfileTransparentOutboundProxy,
									},
								},
								{
									Raw: &konvoy_mesh.ProxyTemplateRawSource{
										Resources: []konvoy_mesh.ProxyTemplateRawResource{{
											Name:    "raw-name",
											Version: "raw-version",
											Resource: `
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
`,
										}},
									},
								},
							},
						},
					},
					expected: `
          resources:
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
              version: v1
            - name: pass_through
              resource:
                '@type': type.googleapis.com/envoy.api.v2.Cluster
                connectTimeout: 5s
                lbPolicy: ORIGINAL_DST_LB
                name: pass_through
                type: ORIGINAL_DST
              version: v1
            - name: raw-name
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
              version: raw-version
`,
				}),
			)
		})
	})
})
