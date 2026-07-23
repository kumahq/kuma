package generator_test

import (
	"context"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
	meshhttproute_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshtcproute_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtcproute/api/v1alpha1"
	. "github.com/kumahq/kuma/v3/pkg/test/matchers"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	"github.com/kumahq/kuma/v3/pkg/test/resources/samples"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
	"github.com/kumahq/kuma/v3/pkg/xds/generator"
)

var _ = Describe("IngressGenerator", func() {
	type testCase struct {
		ingress          string
		expected         string
		proxyZone        string
		meshResourceList []*core_xds.MeshProxyResources
		unifiedNaming    bool
	}

	DescribeTable("should generate Envoy xDS resources",
		func(given testCase) {
			gen := generator.IngressGenerator{}

			zoneIngress := &mesh_proto.ZoneIngress{}
			Expect(util_proto.FromYAML([]byte(given.ingress), zoneIngress)).To(Succeed())

			zoneIngressRes := &core_mesh.ZoneIngressResource{
				Meta: &test_model.ResourceMeta{
					Version: "1",
				},
				Spec: zoneIngress,
			}
			proxy := &core_xds.Proxy{
				Id:         *core_xds.BuildProxyId("default", "ingress"),
				Zone:       given.proxyZone,
				APIVersion: envoy_common.APIV3,
				ZoneIngressProxy: &core_xds.ZoneIngressProxy{
					ZoneIngressResource: zoneIngressRes,
					MeshResourceList:    given.meshResourceList,
				},
				InternalAddresses: DummyInternalAddresses,
				Metadata: &core_xds.DataplaneMetadata{
					Features: map[string]bool{xds_types.FeatureUnifiedResourceNaming: given.unifiedNaming},
				},
			}

			// when
			rs, err := gen.Generate(context.Background(), nil, xds_context.Context{
				ControlPlane: &xds_context.ControlPlaneContext{
					Zone: "east",
				},
			}, proxy)
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
			ingress: `
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
			meshResourceList: []*core_xds.MeshProxyResources{
				{
					Mesh: builders.Mesh().WithName("mesh1").Build(),
					EndpointMap: map[core_xds.ServiceName][]core_xds.Endpoint{
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
				},
			},
		}),
		Entry("01. unified naming, single mesh", testCase{
			ingress: `
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
			expected:      "01.unified-naming.envoy.golden.yaml",
			unifiedNaming: true,
			meshResourceList: []*core_xds.MeshProxyResources{
				{
					Mesh: builders.Mesh().WithName("mesh1").Build(),
					EndpointMap: map[core_xds.ServiceName][]core_xds.Endpoint{
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
				},
			},
		}),
		Entry("02. default trafficroute, empty ingress", testCase{
			ingress: `
            networking:
              address: 10.0.0.1
              port: 10001
`,
			proxyZone:        "zone-main",
			expected:         "02.envoy.golden.yaml",
			meshResourceList: []*core_xds.MeshProxyResources{},
		}),
		Entry("03. trafficroute with many destinations", testCase{
			ingress: `
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
			meshResourceList: []*core_xds.MeshProxyResources{
				{
					Mesh: builders.Mesh().WithName("mesh1").Build(),
					EndpointMap: map[core_xds.ServiceName][]core_xds.Endpoint{
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
				},
			},
		}),
		Entry("04. several services in several meshes", testCase{
			ingress: `
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
			meshResourceList: []*core_xds.MeshProxyResources{
				{
					Mesh: builders.Mesh().WithName("mesh1").Build(),
					EndpointMap: map[core_xds.ServiceName][]core_xds.Endpoint{
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
				},
				{
					Mesh: builders.Mesh().WithName("mesh2").Build(),
					EndpointMap: map[core_xds.ServiceName][]core_xds.Endpoint{
						"frontend": {
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
						"backend": {
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
						},
					},
				},
			},
		}),
		Entry("05. trafficroute repeated", testCase{
			ingress: `
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
			meshResourceList: []*core_xds.MeshProxyResources{
				{
					Mesh: builders.Mesh().WithName("mesh1").Build(),
					EndpointMap: map[core_xds.ServiceName][]core_xds.Endpoint{
						"backend": {},
					},
				},
			},
		}),
		Entry("with MeshHTTPRoute", testCase{
			ingress: `
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
			expected: "with-meshhttproute.envoy.golden.yaml",
			meshResourceList: []*core_xds.MeshProxyResources{
				{
					Mesh: builders.Mesh().WithName("mesh1").Build(),
					EndpointMap: map[core_xds.ServiceName][]core_xds.Endpoint{
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
					Resources: map[core_model.ResourceType]core_model.ResourceList{
						meshhttproute_api.MeshHTTPRouteType: &meshhttproute_api.MeshHTTPRouteResourceList{
							Items: []*meshhttproute_api.MeshHTTPRouteResource{
								{
									Spec: &meshhttproute_api.MeshHTTPRoute{
										TargetRef: &common_api.TopLevelTargetRef{
											Kind: common_api.MeshService,
											Name: pointer.To("frontend"),
										},
										To: &[]meshhttproute_api.To{{
											TargetRef: common_api.OutboundTargetRef{
												Kind: common_api.MeshService,
												Name: pointer.To("backend"),
											},
											Rules: []meshhttproute_api.Rule{{
												Matches: []meshhttproute_api.Match{{
													Path: &meshhttproute_api.PathMatch{
														Type:  meshhttproute_api.PathPrefix,
														Value: "/v1",
													},
												}},
												Default: meshhttproute_api.RuleConf{
													BackendRefs: &[]common_api.BackendRef{{
														TargetRef: common_api.TargetRef{
															Kind: common_api.LegacyMeshServiceSubsetKind(),
															Name: pointer.To("backend"),
															Tags: &map[string]string{
																"version": "v1",
																"region":  "eu",
															},
														},
													}},
												},
											}},
										}},
									},
								},
							},
						},
					},
				},
			},
		}),
		Entry("with MeshHTTPRoute and subsets", testCase{
			ingress: `
            networking:
              address: 10.0.0.1
              port: 10001
            availableServices:
              - mesh: mesh1
                tags:
                  kuma.io/service: backend
                  version: v1
                  region: eu
                  kuma.io/zone: zone
              - mesh: mesh1
                tags:
                  kuma.io/service: backend
                  version: v2
                  region: us
                  kuma.io/zone: zone
`,
			expected: "with-meshhttproute-subset.envoy.golden.yaml",
			meshResourceList: []*core_xds.MeshProxyResources{
				{
					Mesh: builders.Mesh().WithName("mesh1").Build(),
					EndpointMap: map[core_xds.ServiceName][]core_xds.Endpoint{
						"backend": {
							{
								Target: "192.168.0.1",
								Port:   2521,
								Tags: map[string]string{
									"kuma.io/service": "backend",
									"version":         "v1",
									"region":          "eu",
									"mesh":            "mesh1",
									"kuma.io/zone":    "zone",
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
									"kuma.io/zone":    "zone",
								},
								Weight: 1,
							},
						},
					},
					Resources: map[core_model.ResourceType]core_model.ResourceList{
						meshhttproute_api.MeshHTTPRouteType: &meshhttproute_api.MeshHTTPRouteResourceList{
							Items: []*meshhttproute_api.MeshHTTPRouteResource{
								{
									Spec: &meshhttproute_api.MeshHTTPRoute{
										TargetRef: &common_api.TopLevelTargetRef{
											Kind: common_api.MeshService,
											Name: pointer.To("frontend"),
										},
										To: &[]meshhttproute_api.To{{
											TargetRef: common_api.OutboundTargetRef{
												Kind: common_api.MeshService,
												Name: pointer.To("backend"),
											},
											Rules: []meshhttproute_api.Rule{{
												Matches: []meshhttproute_api.Match{{
													Path: &meshhttproute_api.PathMatch{
														Type:  meshhttproute_api.PathPrefix,
														Value: "/v1",
													},
												}},
												Default: meshhttproute_api.RuleConf{
													BackendRefs: &[]common_api.BackendRef{{
														TargetRef: common_api.TargetRef{
															Kind: common_api.LegacyMeshServiceSubsetKind(),
															Name: pointer.To("backend"),
															Tags: &map[string]string{
																"version":      "v1",
																"region":       "eu",
																"kuma.io/zone": "zone",
															},
														},
													}},
												},
											}},
										}},
									},
								},
							},
						},
					},
				},
			},
		}),
		Entry("with MeshTCPRoute and subsets", testCase{
			ingress: `
            networking:
              address: 10.0.0.1
              port: 10001
            availableServices:
              - mesh: mesh1
                tags:
                  kuma.io/service: backend
                  version: v1
                  region: eu
                  kuma.io/zone: zone
              - mesh: mesh1
                tags:
                  kuma.io/service: backend
                  version: v2
                  region: us
                  kuma.io/zone: zone
`,
			expected: "with-meshtcproute-subset.envoy.golden.yaml",
			meshResourceList: []*core_xds.MeshProxyResources{
				{
					Mesh: builders.Mesh().WithName("mesh1").Build(),
					EndpointMap: map[core_xds.ServiceName][]core_xds.Endpoint{
						"backend": {
							{
								Target: "192.168.0.1",
								Port:   2521,
								Tags: map[string]string{
									"kuma.io/service": "backend",
									"version":         "v1",
									"region":          "eu",
									"mesh":            "mesh1",
									"kuma.io/zone":    "zone",
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
									"kuma.io/zone":    "zone",
								},
								Weight: 1,
							},
						},
					},
					Resources: map[core_model.ResourceType]core_model.ResourceList{
						meshtcproute_api.MeshTCPRouteType: &meshtcproute_api.MeshTCPRouteResourceList{
							Items: []*meshtcproute_api.MeshTCPRouteResource{
								{
									Spec: &meshtcproute_api.MeshTCPRoute{
										TargetRef: &common_api.TopLevelTargetRef{
											Kind: common_api.MeshService,
											Name: pointer.To("frontend"),
										},
										To: &[]meshtcproute_api.To{{
											TargetRef: common_api.OutboundTargetRef{
												Kind: common_api.MeshService,
												Name: pointer.To("backend"),
											},
											Rules: []meshtcproute_api.Rule{{
												Default: meshtcproute_api.RuleConf{
													BackendRefs: &[]common_api.BackendRef{{
														TargetRef: common_api.TargetRef{
															Kind: common_api.LegacyMeshServiceSubsetKind(),
															Name: pointer.To("backend"),
															Tags: &map[string]string{
																"version":      "v1",
																"region":       "eu",
																"kuma.io/zone": "zone",
															},
														},
													}},
												},
											}},
										}},
									},
								},
							},
						},
					},
				},
			},
		}),
		Entry("with instance tags", testCase{
			ingress: `
            networking:
              address: 10.0.0.1
              port: 10001
            availableServices:
              - mesh: mesh1
                tags:
                  kuma.io/service: backend
                  kuma.io/instance: ins-0
              - mesh: mesh1
                tags:
                  kuma.io/service: backend
                  kuma.io/instance: ins-1
`,
			expected: "with-instance-tags.envoy.golden.yaml",
			meshResourceList: []*core_xds.MeshProxyResources{
				{
					Mesh: builders.Mesh().WithName("mesh1").Build(),
					EndpointMap: map[core_xds.ServiceName][]core_xds.Endpoint{
						"backend": {
							{
								Target: "192.168.0.1",
								Port:   2521,
								Tags: map[string]string{
									"kuma.io/service":  "backend",
									"kuma.io/instance": "ins-0",
									"mesh":             "mesh1",
								},
								Weight: 1,
							},
							{
								Target: "192.168.0.2",
								Port:   2521,
								Tags: map[string]string{
									"kuma.io/service":  "backend",
									"kuma.io/instance": "ins-1",
									"mesh":             "mesh1",
								},
								Weight: 1,
							},
						},
					},
				},
			},
		}),
		Entry("with MeshService", testCase{
			ingress: `
            networking:
              address: 10.0.0.1
              port: 10001
`,
			expected: "mesh-service.envoy.golden.yaml",
			meshResourceList: []*core_xds.MeshProxyResources{
				{
					Mesh: builders.Mesh().WithName("mesh1").Build(),
					EndpointMap: map[core_xds.ServiceName][]core_xds.Endpoint{
						"backend": {
							{
								Target: "192.168.0.1",
								Port:   2521,
								Tags: map[string]string{
									"kuma.io/service": "backend_svc_80",
								},
								Weight: 1,
							},
						},
						"backend2": {
							{
								Target: "192.168.0.2",
								Port:   2521,
								Tags: map[string]string{
									"kuma.io/service": "backend2_svc_80",
								},
								Weight: 1,
							},
						},
					},
					Resources: map[core_model.ResourceType]core_model.ResourceList{
						meshservice_api.MeshServiceType: &meshservice_api.MeshServiceResourceList{
							Items: []*meshservice_api.MeshServiceResource{
								// should be included because of zone label = local zone
								samples.MeshServiceBackendBuilder().
									WithLabels(map[string]string{mesh_proto.ZoneTag: "east"}).
									Build(),
								// should be included because zone label is missing
								samples.MeshServiceBackendBuilder().
									WithName("backend2").
									Build(),
								// should be ignored because it's non-local MeshService
								samples.MeshServiceBackendBuilder().
									WithName("xyz").
									WithLabels(map[string]string{mesh_proto.ZoneTag: "west"}).
									Build(),
							},
						},
					},
				},
			},
		}),
	)
})
