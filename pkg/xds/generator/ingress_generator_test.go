package generator_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ = Describe("IngressGenerator", func() {
	type testCase struct {
		dataplane       string
		expected        string
		outboundTargets core_xds.EndpointMap
		trafficRoutes   *core_mesh.TrafficRouteResourceList
	}

	DescribeTable("should generate Envoy xDS resources",
		func(given testCase) {
			gen := generator.IngressGenerator{}

			zoneIngress := &mesh_proto.ZoneIngress{}
			Expect(util_proto.FromYAML([]byte(given.dataplane), zoneIngress)).To(Succeed())

			zoneIngressRes := &core_mesh.ZoneIngressResource{
				Meta: &test_model.ResourceMeta{
					Version: "1",
				},
				Spec: zoneIngress,
			}
			proxy := &core_xds.Proxy{
				Id:          *core_xds.BuildProxyId("default", "ingress"),
				ZoneIngress: zoneIngressRes,
				APIVersion:  envoy_common.APIV3,
				Routing: core_xds.Routing{
					OutboundTargets: given.outboundTargets,
				},
				ZoneIngressProxy: &core_xds.ZoneIngressProxy{
					TrafficRouteList: given.trafficRoutes,
					GatewayRoutes:    &core_mesh.MeshGatewayRouteResourceList{},
				},
			}

			// when
			rs, err := gen.Generate(xds_context.Context{}, proxy)
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

			// and output matches golden files
			Expect(actual).To(MatchGoldenYAML(filepath.Join("testdata", "ingress", given.expected)))
		},
		Entry("01. default trafficroute, single mesh", testCase{
			dataplane: `
            networking:
              address: 10.0.0.1
              port: 10001
            availableServices:
              - mesh: mesh1
                tags:
                  kuma.io/service: backend
                  version: v1
                  region: eu
              - mesh: mesh1
                tags:
                  kuma.io/service: backend
                  version: v2
                  region: us
`,
			expected: "01.envoy.golden.yaml",
			outboundTargets: map[core_xds.ServiceName][]core_xds.Endpoint{
				"backend": {
					{
						Target: "192.168.0.1",
						Port:   2521,
						Tags: map[string]string{
							"kuma.io/service": "backend",
							"version":         "v1",
							"region":          "eu",
							"mesh":            "mesh1",
						},
						Weight: 1,
					},
					{
						Target: "192.168.0.2",
						Port:   2521,
						Tags: map[string]string{
							"kuma.io/service": "backend",
							"version":         "v2",
							"region":          "us",
							"mesh":            "mesh1",
						},
						Weight: 1,
					},
				},
			},
			trafficRoutes: &core_mesh.TrafficRouteResourceList{
				Items: []*core_mesh.TrafficRouteResource{
					{
						Spec: &mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{{
								Match: mesh_proto.MatchAnyService(),
							}},
							Destinations: []*mesh_proto.Selector{{
								Match: mesh_proto.MatchAnyService(),
							}},
							Conf: &mesh_proto.TrafficRoute_Conf{
								Destination: mesh_proto.MatchAnyService(),
							},
						},
					},
				},
			},
		}),
		Entry("02. default trafficroute, empty ingress", testCase{
			dataplane: `
            networking:
              address: 10.0.0.1
              port: 10001
`,
			expected:        "02.envoy.golden.yaml",
			outboundTargets: map[core_xds.ServiceName][]core_xds.Endpoint{},
			trafficRoutes: &core_mesh.TrafficRouteResourceList{
				Items: []*core_mesh.TrafficRouteResource{
					{
						Spec: &mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{{
								Match: mesh_proto.MatchAnyService(),
							}},
							Destinations: []*mesh_proto.Selector{{
								Match: mesh_proto.MatchAnyService(),
							}},
							Conf: &mesh_proto.TrafficRoute_Conf{
								Destination: mesh_proto.MatchAnyService(),
							},
						},
					},
				},
			},
		}),
		Entry("03. trafficroute with many destinations", testCase{
			dataplane: `
            networking:
              address: 10.0.0.1
              port: 10001
            availableServices:
              - mesh: mesh1
                tags:
                  kuma.io/service: backend
                  version: v1
                  region: eu
              - mesh: mesh1
                tags:
                  kuma.io/service: backend
                  version: v2
                  region: us
`,
			expected: "03.envoy.golden.yaml",
			outboundTargets: map[core_xds.ServiceName][]core_xds.Endpoint{
				"backend": {
					{
						Target: "192.168.0.1",
						Port:   2521,
						Tags: map[string]string{
							"kuma.io/service": "backend",
							"version":         "v1",
							"region":          "eu",
							"mesh":            "mesh1",
						},
						Weight: 1,
					},
					{
						Target: "192.168.0.2",
						Port:   2521,
						Tags: map[string]string{
							"kuma.io/service": "backend",
							"version":         "v2",
							"region":          "us",
							"mesh":            "mesh1",
						},
						Weight: 1,
					},
				},
			},
			trafficRoutes: &core_mesh.TrafficRouteResourceList{
				Items: []*core_mesh.TrafficRouteResource{
					{
						Spec: &mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{{
								Match: mesh_proto.MatchAnyService(),
							}},
							Destinations: []*mesh_proto.Selector{{
								Match: mesh_proto.MatchAnyService(),
							}},
							Conf: &mesh_proto.TrafficRoute_Conf{
								Destination: mesh_proto.MatchAnyService(),
							},
						},
					},
					{
						Spec: &mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{{
								Match: mesh_proto.MatchAnyService(),
							}},
							Destinations: []*mesh_proto.Selector{{
								Match: mesh_proto.MatchAnyService(),
							}},
							Conf: &mesh_proto.TrafficRoute_Conf{
								Split: []*mesh_proto.TrafficRoute_Split{
									{
										Weight: util_proto.UInt32(10),
										Destination: map[string]string{
											mesh_proto.ServiceTag: "backend",
											"version":             "v2",
										},
									},
									{
										Weight: util_proto.UInt32(90),
										Destination: map[string]string{
											mesh_proto.ServiceTag: "backend",
											"region":              "eu",
										},
									},
								},
							},
						},
					},
				},
			},
		}),
		Entry("04. several services in several meshes", testCase{
			dataplane: `
            networking:
              address: 10.0.0.1
              port: 10001
            availableServices:
              - mesh: mesh1
                tags:
                  kuma.io/service: backend
                  version: v1
                  region: eu
              - mesh: mesh1
                tags:
                  kuma.io/service: backend
                  version: v2
                  region: us
              - mesh: mesh2
                tags:
                  kuma.io/service: backend
                  cloud: eks
                  arch: ARM
                  os: ubuntu
                  region: asia
                  version: v3
              - mesh: mesh2
                tags:
                  kuma.io/service: frontend
                  cloud: gke
                  arch: x86
                  os: debian
                  region: eu
                  version: v1
              - mesh: mesh2
                tags:
                  kuma.io/service: frontend
                  cloud: aks
                  version: v2
`,
			expected: "04.envoy.golden.yaml",
			outboundTargets: map[core_xds.ServiceName][]core_xds.Endpoint{
				"backend": {
					{
						Target: "192.168.0.1",
						Port:   2521,
						Tags: map[string]string{
							"kuma.io/service": "backend",
							"version":         "v1",
							"region":          "eu",
							"mesh":            "mesh1",
						},
						Weight: 1,
					},
					{
						Target: "192.168.0.2",
						Port:   2521,
						Tags: map[string]string{
							"kuma.io/service": "backend",
							"version":         "v2",
							"region":          "us",
							"mesh":            "mesh1",
						},
						Weight: 1,
					},
					{
						Target: "192.168.0.3",
						Port:   2521,
						Tags: map[string]string{
							"kuma.io/service": "backend",
							"cloud":           "eks",
							"arch":            "ARM",
							"os":              "ubuntu",
							"region":          "asia",
							"version":         "v3",
							"mesh":            "mesh2",
						},
						Weight: 1,
					},
					{
						Target: "192.168.0.4",
						Port:   2521,
						Tags: map[string]string{
							"kuma.io/service": "frontend",
							"cloud":           "gke",
							"arch":            "x86",
							"os":              "debian",
							"region":          "eu",
							"version":         "v1",
							"mesh":            "mesh2",
						},
						Weight: 1,
					},
					{
						Target: "192.168.0.5",
						Port:   2521,
						Tags: map[string]string{
							"kuma.io/service": "frontend",
							"cloud":           "aks",
							"version":         "v2",
							"mesh":            "mesh2",
						},
						Weight: 1,
					},
				},
			},
			trafficRoutes: &core_mesh.TrafficRouteResourceList{
				Items: []*core_mesh.TrafficRouteResource{
					{
						Spec: &mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{{
								Match: mesh_proto.MatchAnyService(),
							}},
							Destinations: []*mesh_proto.Selector{{
								Match: mesh_proto.MatchAnyService(),
							}},
							Conf: &mesh_proto.TrafficRoute_Conf{
								Destination: mesh_proto.MatchAnyService(),
							},
						},
					},
					{
						Spec: &mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{{
								Match: mesh_proto.MatchAnyService(),
							}},
							Destinations: []*mesh_proto.Selector{{
								Match: mesh_proto.MatchAnyService(),
							}},
							Conf: &mesh_proto.TrafficRoute_Conf{
								Destination: map[string]string{
									mesh_proto.ServiceTag: "*",
									"version":             "v2",
								},
							},
						},
					},
					{
						Spec: &mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{{
								Match: mesh_proto.MatchAnyService(),
							}},
							Destinations: []*mesh_proto.Selector{{
								Match: mesh_proto.MatchAnyService(),
							}},
							Conf: &mesh_proto.TrafficRoute_Conf{
								Split: []*mesh_proto.TrafficRoute_Split{
									{
										Weight: util_proto.UInt32(10),
										Destination: map[string]string{
											mesh_proto.ServiceTag: "backend",
											"version":             "v2",
										},
									},
									{
										Weight: util_proto.UInt32(90),
										Destination: map[string]string{
											mesh_proto.ServiceTag: "backend",
											"region":              "eu",
										},
									},
								},
							},
						},
					},
					{
						Spec: &mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{{
								Match: mesh_proto.MatchAnyService(),
							}},
							Destinations: []*mesh_proto.Selector{{
								Match: mesh_proto.MatchAnyService(),
							}},
							Conf: &mesh_proto.TrafficRoute_Conf{
								Split: []*mesh_proto.TrafficRoute_Split{
									{
										Weight: util_proto.UInt32(10),
										Destination: map[string]string{
											mesh_proto.ServiceTag: "frontend",
											"region":              "eu",
											"cloud":               "gke",
										},
									},
									{
										Weight: util_proto.UInt32(90),
										Destination: map[string]string{
											mesh_proto.ServiceTag: "frontend",
											"cloud":               "aks",
										},
									},
								},
							},
						},
					},
				},
			},
		}),
		Entry("05. trafficroute repeated", testCase{
			dataplane: `
            networking:
              address: 10.0.0.1
              port: 10001
            availableServices:
              - mesh: mesh1
                tags:
                  kuma.io/service: backend
                  version: v1
                  region: eu
              - mesh: mesh1
                tags:
                  kuma.io/service: backend
                  version: v2
                  region: us
`,
			expected: "05.envoy.golden.yaml",
			outboundTargets: map[core_xds.ServiceName][]core_xds.Endpoint{
				"backend": {},
			},
			trafficRoutes: &core_mesh.TrafficRouteResourceList{
				Items: []*core_mesh.TrafficRouteResource{
					{
						Spec: &mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{{
								Match: mesh_proto.MatchAnyService(),
							}},
							Destinations: []*mesh_proto.Selector{{
								Match: mesh_proto.MatchAnyService(),
							}},
							Conf: &mesh_proto.TrafficRoute_Conf{
								Destination: mesh_proto.MatchAnyService(),
							},
						},
					},
					{
						Spec: &mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{{
								Match: mesh_proto.MatchService("foo"),
							}},
							Destinations: []*mesh_proto.Selector{{
								Match: mesh_proto.MatchAnyService(),
							}},
							Conf: &mesh_proto.TrafficRoute_Conf{
								Destination: mesh_proto.MatchAnyService(),
							},
						},
					},
				},
			},
		}),
	)
})
