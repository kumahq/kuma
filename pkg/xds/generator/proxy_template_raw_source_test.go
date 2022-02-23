package generator_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	model "github.com/kumahq/kuma/pkg/core/xds"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ = Describe("ProxyTemplateRawSource", func() {

	Context("Manually-defined xDS resources are not valid", func() {

		type testCase struct {
			proxy *model.Proxy
			raw   []*mesh_proto.ProxyTemplateRawResource
			err   string
		}

		DescribeTable("Avoid producing invalid Envoy xDS resources",
			func(given testCase) {
				// setup
				gen := &generator.ProxyTemplateRawSource{
					Resources: given.raw,
				}
				ctx := xds_context.Context{}

				// when
				rs, err := gen.Generate(ctx, given.proxy)

				// then
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(given.err))
				Expect(rs).To(BeNil())
			},
			Entry("should fail when `resource` field is empty", testCase{
				proxy: &model.Proxy{
					Id: *model.BuildProxyId("", "side-car"),
					Dataplane: &core_mesh.DataplaneResource{
						Meta: &test_model.ResourceMeta{
							Version: "v1",
						},
						Spec: &mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Port:        80,
										ServicePort: 8080,
									},
								},
							},
						},
					},
					APIVersion: envoy_common.APIV3,
				},
				raw: []*mesh_proto.ProxyTemplateRawResource{{
					Name:     "raw-name",
					Version:  "raw-version",
					Resource: ``,
				}},
				err: "invalid empty type URL",
			}),
			Entry("should fail when `resource` field is neither a YAML nor a JSON", testCase{
				proxy: &model.Proxy{
					Id: *model.BuildProxyId("", "side-car"),
					Dataplane: &core_mesh.DataplaneResource{
						Meta: &test_model.ResourceMeta{
							Version: "v1",
						},
						Spec: &mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Port:        80,
										ServicePort: 8080,
									},
								},
							},
						},
					},
					APIVersion: envoy_common.APIV3,
				},
				raw: []*mesh_proto.ProxyTemplateRawResource{{
					Name:     "raw-name",
					Version:  "raw-version",
					Resource: `{`,
				}},
				err: "raw.resources[0]{name=\"raw-name\"}.resource: unexpected EOF",
			}),
			Entry("should fail when `resource` field has unknown @type", testCase{
				proxy: &model.Proxy{
					Id: *model.BuildProxyId("", "side-car"),
					Dataplane: &core_mesh.DataplaneResource{
						Meta: &test_model.ResourceMeta{
							Version: "v1",
						},
						Spec: &mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Port:        80,
										ServicePort: 8080,
									},
								},
							},
						},
					},
					APIVersion: envoy_common.APIV3,
				},
				raw: []*mesh_proto.ProxyTemplateRawResource{{
					Name:    "raw-name",
					Version: "raw-version",
					Resource: `
                    '@type': type.googleapis.com/unknown.Resource
`,
				}},
				err: "raw.resources[0]{name=\"raw-name\"}.resource: could not resolve Any message type: type.googleapis.com/unknown.Resource",
			}),
			Entry("should fail when `resource` field is a YAML without '@type' field", testCase{
				proxy: &model.Proxy{
					Id: *model.BuildProxyId("", "side-car"),
					Dataplane: &core_mesh.DataplaneResource{
						Meta: &test_model.ResourceMeta{
							Version: "v1",
						},
						Spec: &mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Port:        80,
										ServicePort: 8080,
									},
								},
							},
						},
					},
					APIVersion: envoy_common.APIV3,
				},
				raw: []*mesh_proto.ProxyTemplateRawResource{{
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
				err: "raw.resources[0]{name=\"raw-name\"}.resource: Any JSON doesn't have '@type'",
			}),
			Entry("should fail when `resource` field is an invalid xDS resource", testCase{
				proxy: &model.Proxy{
					Id: *model.BuildProxyId("", "side-car"),
					Dataplane: &core_mesh.DataplaneResource{
						Meta: &test_model.ResourceMeta{
							Version: "v1",
						},
						Spec: &mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Port:        80,
										ServicePort: 8080,
									},
								},
							},
						},
					},
					APIVersion: envoy_common.APIV3,
				},
				raw: []*mesh_proto.ProxyTemplateRawResource{{
					Name:    "raw-name",
					Version: "raw-version",
					Resource: `
                    '@type': type.googleapis.com/envoy.config.cluster.v3.Cluster
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
				err: "raw.resources[0]{name=\"raw-name\"}.resource: invalid Cluster.Name: value length must be at least 1 runes",
			}),
		)
	})

	Context("Manually-defined xDS resources are valid", func() {

		type testCase struct {
			proxy    *model.Proxy
			raw      []*mesh_proto.ProxyTemplateRawResource
			expected string
		}

		DescribeTable("Generate Envoy xDS resources", func(given testCase) {
			// setup
			gen := &generator.ProxyTemplateRawSource{
				Resources: given.raw,
			}
			ctx := xds_context.Context{}

			// when
			rs, err := gen.Generate(ctx, given.proxy)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			resp, err := rs.List().ToDeltaDiscoveryResponse()
			// then
			Expect(err).ToNot(HaveOccurred())
			// when
			actual, err := util_proto.ToYAML(resp)
			// then
			Expect(err).ToNot(HaveOccurred())

			Expect(actual).To(MatchYAML(given.expected))
		},
			Entry("should support empty resource list", testCase{
				proxy: &model.Proxy{
					Id: *model.BuildProxyId("", "side-car"),
					Dataplane: &core_mesh.DataplaneResource{
						Meta: &test_model.ResourceMeta{
							Version: "v1",
						},
						Spec: &mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Port:        80,
										ServicePort: 8080,
									},
								},
							},
						},
					},
					APIVersion: envoy_common.APIV3,
				},
				raw:      nil,
				expected: `{}`,
			}),
			Entry("should support Listener resource as YAML", testCase{
				proxy: &model.Proxy{
					Id: *model.BuildProxyId("", "side-car"),
					Dataplane: &core_mesh.DataplaneResource{
						Meta: &test_model.ResourceMeta{
							Version: "v1",
						},
						Spec: &mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Port:        80,
										ServicePort: 8080,
									},
								},
							},
						},
					},
					APIVersion: envoy_common.APIV3,
				},
				raw: []*mesh_proto.ProxyTemplateRawResource{{
					Name:    "raw-name",
					Version: "raw-version",
					Resource: `
          '@type': type.googleapis.com/envoy.config.listener.v3.Listener
          address:
            socketAddress:
              address: 0.0.0.0
              portValue: 15001
          filterChains:
          - filters:
            - name: envoy.filters.network.tcp_proxy
              typedConfig:
                '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                cluster: pass_through
                statPrefix: pass_through
          name: catch_all
          useOriginalDst: true
`,
				}},
				expected: `
          resources:
            - name: raw-name
              resource:
                '@type': type.googleapis.com/envoy.config.listener.v3.Listener
                address:
                  socketAddress:
                    address: 0.0.0.0
                    portValue: 15001
                filterChains:
                - filters:
                  - name: envoy.filters.network.tcp_proxy
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                      cluster: pass_through
                      statPrefix: pass_through
                name: catch_all
                useOriginalDst: true
`,
			}),
			Entry("should support Cluster resource as YAML", testCase{
				proxy: &model.Proxy{
					Id: *model.BuildProxyId("", "side-car"),
					Dataplane: &core_mesh.DataplaneResource{
						Meta: &test_model.ResourceMeta{
							Version: "v1",
						},
						Spec: &mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Port:        80,
										ServicePort: 8080,
									},
								},
							},
						},
					},
					APIVersion: envoy_common.APIV3,
				},
				raw: []*mesh_proto.ProxyTemplateRawResource{{
					Name:    "raw-name",
					Version: "raw-version",
					Resource: `
                    '@type': type.googleapis.com/envoy.config.cluster.v3.Cluster
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
				expected: `
          resources:
            - name: raw-name
              resource:
                '@type': type.googleapis.com/envoy.config.cluster.v3.Cluster
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
			}),
			Entry("should support Cluster resource as JSON", testCase{
				proxy: &model.Proxy{
					Id: *model.BuildProxyId("", "side-car"),
					Dataplane: &core_mesh.DataplaneResource{
						Meta: &test_model.ResourceMeta{
							Version: "v1",
						},
						Spec: &mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Port:        80,
										ServicePort: 8080,
									},
								},
							},
						},
					},
					APIVersion: envoy_common.APIV3,
				},
				raw: []*mesh_proto.ProxyTemplateRawResource{{
					Name:    "raw-name",
					Version: "raw-version",
					Resource: `
              {
                "@type": "type.googleapis.com/envoy.config.cluster.v3.Cluster",
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
				expected: `
          resources:
            - name: raw-name
              resource:
                '@type': type.googleapis.com/envoy.config.cluster.v3.Cluster
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
			}),
		)
	})
})
