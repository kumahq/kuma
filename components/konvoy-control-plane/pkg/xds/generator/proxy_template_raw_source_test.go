package generator_test

import (
	xds_context "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/xds"
	util_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/proto"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/generator"

	test_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/test/resources/model"
)

var _ = Describe("ProxyTemplateRawSource", func() {

	Context("Manually-defined xDS resources are not valid", func() {

		type testCase struct {
			proxy *model.Proxy
			raw   *mesh_proto.ProxyTemplateRawSource
			err   interface{}
		}

		DescribeTable("Avoid producing invalid Envoy xDS resources",
			func(given testCase) {
				// setup
				gen := &generator.ProxyTemplateRawSource{
					Raw: given.raw,
				}
				ctx := xds_context.Context{}

				// when
				rs, err := gen.Generate(ctx, given.proxy)

				// then
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(given.err))
				Expect(rs).To(BeNil())
			},
			Entry("should fail when `resource` field is empty", testCase{
				proxy: &model.Proxy{
					Id: model.ProxyId{Name: "side-car", Namespace: "default"},
					Dataplane: &mesh_core.DataplaneResource{
						Meta: &test_model.ResourceMeta{
							Version: "v1",
						},
						Spec: mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{Interface: "192.168.0.1:80:8080"},
								},
							},
						},
					},
				},
				raw: &mesh_proto.ProxyTemplateRawSource{
					Resources: []*mesh_proto.ProxyTemplateRawResource{{
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
					Dataplane: &mesh_core.DataplaneResource{
						Meta: &test_model.ResourceMeta{
							Version: "v1",
						},
						Spec: mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{Interface: "192.168.0.1:80:8080"},
								},
							},
						},
					},
				},
				raw: &mesh_proto.ProxyTemplateRawSource{
					Resources: []*mesh_proto.ProxyTemplateRawResource{{
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
					Dataplane: &mesh_core.DataplaneResource{
						Meta: &test_model.ResourceMeta{
							Version: "v1",
						},
						Spec: mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{Interface: "192.168.0.1:80:8080"},
								},
							},
						},
					},
				},
				raw: &mesh_proto.ProxyTemplateRawSource{
					Resources: []*mesh_proto.ProxyTemplateRawResource{{
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
					Dataplane: &mesh_core.DataplaneResource{
						Meta: &test_model.ResourceMeta{
							Version: "v1",
						},
						Spec: mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{Interface: "192.168.0.1:80:8080"},
								},
							},
						},
					},
				},
				raw: &mesh_proto.ProxyTemplateRawSource{
					Resources: []*mesh_proto.ProxyTemplateRawResource{{
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
					Dataplane: &mesh_core.DataplaneResource{
						Meta: &test_model.ResourceMeta{
							Version: "v1",
						},
						Spec: mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{Interface: "192.168.0.1:80:8080"},
								},
							},
						},
					},
				},
				raw: &mesh_proto.ProxyTemplateRawSource{
					Resources: []*mesh_proto.ProxyTemplateRawResource{{
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
			raw      *mesh_proto.ProxyTemplateRawSource
			expected string
		}

		DescribeTable("Generate Envoy xDS resources", func(given testCase) {
			// setup
			gen := &generator.ProxyTemplateRawSource{
				Raw: given.raw,
			}
			ctx := xds_context.Context{}

			// when
			rs, err := gen.Generate(ctx, given.proxy)

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
					Dataplane: &mesh_core.DataplaneResource{
						Meta: &test_model.ResourceMeta{
							Version: "v1",
						},
						Spec: mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{Interface: "192.168.0.1:80:8080"},
								},
							},
						},
					},
				},
				raw: &mesh_proto.ProxyTemplateRawSource{
					Resources: nil,
				},
				expected: `{}`,
			}),
			Entry("should support Listener resource as YAML", testCase{
				proxy: &model.Proxy{
					Id: model.ProxyId{Name: "side-car", Namespace: "default"},
					Dataplane: &mesh_core.DataplaneResource{
						Meta: &test_model.ResourceMeta{
							Version: "v1",
						},
						Spec: mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{Interface: "192.168.0.1:80:8080"},
								},
							},
						},
					},
				},
				raw: &mesh_proto.ProxyTemplateRawSource{
					Resources: []*mesh_proto.ProxyTemplateRawResource{{
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
					Dataplane: &mesh_core.DataplaneResource{
						Meta: &test_model.ResourceMeta{
							Version: "v1",
						},
						Spec: mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{Interface: "192.168.0.1:80:8080"},
								},
							},
						},
					},
				},
				raw: &mesh_proto.ProxyTemplateRawSource{
					Resources: []*mesh_proto.ProxyTemplateRawResource{{
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
					Dataplane: &mesh_core.DataplaneResource{
						Meta: &test_model.ResourceMeta{
							Version: "v1",
						},
						Spec: mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{Interface: "192.168.0.1:80:8080"},
								},
							},
						},
					},
				},
				raw: &mesh_proto.ProxyTemplateRawSource{
					Resources: []*mesh_proto.ProxyTemplateRawResource{{
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
