package v1alpha1_test

import (
	"context"
	"fmt"
	"path/filepath"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/intstr"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshcircuitbreaker/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshcircuitbreaker/plugin/v1alpha1"
	gateway_plugin "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/pkg/test/matchers"
	test_matchers "github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	test_xds "github.com/kumahq/kuma/pkg/test/xds"
	xds_builders "github.com/kumahq/kuma/pkg/test/xds/builders"
	xds_samples "github.com/kumahq/kuma/pkg/test/xds/samples"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ = Describe("MeshCircuitBreaker", func() {
	backendMeshServiceIdentifier := core_model.TypedResourceIdentifier{
		ResourceIdentifier: core_model.ResourceIdentifier{
			Name:      "backend",
			Mesh:      "default",
			Namespace: "backend-ns",
			Zone:      "zone-1",
		},
		ResourceType: "MeshService",
		SectionName:  "",
	}

	type sidecarTestCase struct {
		resources       []*core_xds.Resource
		toRules         core_rules.ToRules
		fromRules       core_rules.FromRules
		expectedCluster []string
	}

	genConnectionLimits := func() *api.ConnectionLimits {
		return &api.ConnectionLimits{
			MaxConnectionPools: pointer.To(uint32(1111)),
			MaxConnections:     pointer.To(uint32(2222)),
			MaxPendingRequests: pointer.To(uint32(3333)),
			MaxRequests:        pointer.To(uint32(4444)),
			MaxRetries:         pointer.To(uint32(5555)),
		}
	}

	genOutlierDetection := func(disabled bool) *api.OutlierDetection {
		return &api.OutlierDetection{
			Disabled:                    &disabled,
			Interval:                    test.ParseDuration("10s"),
			BaseEjectionTime:            test.ParseDuration("8s"),
			MaxEjectionPercent:          pointer.To(uint32(88)),
			SplitExternalAndLocalErrors: pointer.To(true),
			Detectors: &api.Detectors{
				TotalFailures: &api.DetectorTotalFailures{
					Consecutive: pointer.To(uint32(12)),
				},
				GatewayFailures: &api.DetectorGatewayFailures{
					Consecutive: pointer.To(uint32(91)),
				},
				LocalOriginFailures: &api.DetectorLocalOriginFailures{
					Consecutive: pointer.To(uint32(3)),
				},
				SuccessRate: &api.DetectorSuccessRateFailures{
					MinimumHosts:            pointer.To(uint32(33)),
					RequestVolume:           pointer.To(uint32(99)),
					StandardDeviationFactor: pointer.To(intstr.FromString("1.9")),
				},
				FailurePercentage: &api.DetectorFailurePercentageFailures{
					MinimumHosts:  pointer.To(uint32(32)),
					RequestVolume: pointer.To(uint32(182)),
					Threshold:     pointer.To(uint32(80)),
				},
			},
		}
	}

	DescribeTable("should generate proper Envoy config",
		func(given sidecarTestCase) {
			resourceSet := core_xds.NewResourceSet()
			resourceSet.Add(given.resources...)

			context := xds_samples.SampleContext()

			proxy := xds_builders.Proxy().
				WithDataplane(
					builders.Dataplane().
						WithName("backend").
						WithMesh("default").
						WithAddress("127.0.0.1").
						WithInboundOfTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, "http"),
				).
				WithOutbounds(xds_types.Outbounds{
					{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
						Port: builders.FirstOutboundPort,
						Tags: map[string]string{
							mesh_proto.ServiceTag: "other-service",
						},
					}},
					{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
						Port: builders.FirstOutboundPort + 1,
						Tags: map[string]string{
							mesh_proto.ServiceTag: "second-service",
						},
					}},
				}).
				WithPolicies(
					xds_builders.MatchedPolicies().WithPolicy(api.MeshCircuitBreakerType, given.toRules, given.fromRules),
				).
				Build()
			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)

			Expect(plugin.Apply(resourceSet, context, proxy)).To(Succeed())

			for idx, expected := range given.expectedCluster {
				Expect(util_proto.ToYAML(resourceSet.ListOf(envoy_resource.ClusterType)[idx].Resource)).
					To(test_matchers.MatchGoldenYAML(filepath.Join("testdata", expected)))
			}
		},
		Entry("basic outbound cluster with connection limits", sidecarTestCase{
			resources: []*core_xds.Resource{
				{
					Name:     "outbound",
					Origin:   generator.OriginOutbound,
					Resource: test_xds.ClusterWithName("other-service"),
				},
			},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{core_rules.Tag{
							Key:   mesh_proto.ServiceTag,
							Value: "other-service",
						}},
						Conf: api.Conf{
							ConnectionLimits: genConnectionLimits(),
						},
					},
				},
			},
			expectedCluster: []string{"outbound_cluster_connection_limits.golden.yaml"},
		}),
		Entry("basic outbound cluster with outlier detection", sidecarTestCase{
			resources: []*core_xds.Resource{
				{
					Name:     "outbound",
					Origin:   generator.OriginOutbound,
					Resource: test_xds.ClusterWithName("second-service"),
				},
			},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{core_rules.Tag{
							Key:   mesh_proto.ServiceTag,
							Value: "second-service",
						}},
						Conf: api.Conf{
							OutlierDetection: genOutlierDetection(false),
						},
					},
				},
			},
			expectedCluster: []string{"outbound_cluster_outlier_detection.golden.yaml"},
		}),
		Entry("basic outbound cluster with outlier detection and disabled=true", sidecarTestCase{
			resources: []*core_xds.Resource{
				{
					Name:     "outbound",
					Origin:   generator.OriginOutbound,
					Resource: test_xds.ClusterWithName("second-service"),
				},
			},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{core_rules.Tag{
							Key:   mesh_proto.ServiceTag,
							Value: "second-service",
						}},
						Conf: api.Conf{
							OutlierDetection: genOutlierDetection(true),
						},
					},
				},
			},
			expectedCluster: []string{"outbound_cluster_outlier_detection_disabled.golden.yaml"},
		}),
		Entry("basic outbound cluster with connection limits and outlier detection", sidecarTestCase{
			resources: []*core_xds.Resource{
				{
					Name:     "outbound",
					Origin:   generator.OriginOutbound,
					Resource: test_xds.ClusterWithName("second-service"),
				},
			},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{core_rules.Tag{
							Key:   mesh_proto.ServiceTag,
							Value: "second-service",
						}},
						Conf: api.Conf{
							ConnectionLimits: genConnectionLimits(),
							OutlierDetection: genOutlierDetection(false),
						},
					},
				},
			},
			expectedCluster: []string{"outbound_cluster_connection_limits_outlier_detection.golden.yaml"},
		}),
		Entry("basic outbound cluster with connection limits, outlier detection and disabled=true", sidecarTestCase{
			resources: []*core_xds.Resource{
				{
					Name:     "outbound",
					Origin:   generator.OriginOutbound,
					Resource: test_xds.ClusterWithName("second-service"),
				},
			},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{core_rules.Tag{
							Key:   mesh_proto.ServiceTag,
							Value: "second-service",
						}},
						Conf: api.Conf{
							ConnectionLimits: genConnectionLimits(),
							OutlierDetection: genOutlierDetection(true),
						},
					},
				},
			},
			expectedCluster: []string{"outbound_cluster_connection_limits_outlier_detection_disabled.golden.yaml"},
		}),
		Entry("basic inbound cluster with connection limits", sidecarTestCase{
			resources: []*core_xds.Resource{
				{
					Name:     "inbound",
					Origin:   generator.OriginInbound,
					Resource: test_xds.ClusterWithName(fmt.Sprintf("localhost:%d", builders.FirstInboundPort)),
				},
			},
			fromRules: core_rules.FromRules{
				Rules: map[core_rules.InboundListener]core_rules.Rules{
					{
						Address: "127.0.0.1",
						Port:    builders.FirstInboundPort,
					}: []*core_rules.Rule{
						{
							Subset: core_rules.Subset{},
							Conf: api.Conf{
								ConnectionLimits: genConnectionLimits(),
							},
						},
					},
				},
			},
			expectedCluster: []string{"inbound_cluster_connection_limits.golden.yaml"},
		}),
		Entry("basic inbound cluster with outlier detection", sidecarTestCase{
			resources: []*core_xds.Resource{
				{
					Name:     "inbound",
					Origin:   generator.OriginInbound,
					Resource: test_xds.ClusterWithName(envoy_names.GetLocalClusterName(builders.FirstInboundPort)),
				},
			},
			fromRules: core_rules.FromRules{
				Rules: map[core_rules.InboundListener]core_rules.Rules{
					{
						Address: "127.0.0.1",
						Port:    builders.FirstInboundPort,
					}: []*core_rules.Rule{
						{
							Subset: core_rules.Subset{},
							Conf: api.Conf{
								OutlierDetection: genOutlierDetection(false),
							},
						},
					},
				},
			},
			expectedCluster: []string{"inbound_cluster_outlier_detection.golden.yaml"},
		}),
		Entry("basic inbound cluster with outlier detection and disabled=true", sidecarTestCase{
			resources: []*core_xds.Resource{
				{
					Name:     "inbound",
					Origin:   generator.OriginInbound,
					Resource: test_xds.ClusterWithName(fmt.Sprintf("localhost:%d", builders.FirstInboundPort)),
				},
			},
			fromRules: core_rules.FromRules{
				Rules: map[core_rules.InboundListener]core_rules.Rules{
					{
						Address: "127.0.0.1",
						Port:    builders.FirstInboundPort,
					}: []*core_rules.Rule{
						{
							Subset: core_rules.Subset{},
							Conf: api.Conf{
								OutlierDetection: genOutlierDetection(true),
							},
						},
					},
				},
			},
			expectedCluster: []string{"inbound_cluster_outlier_detection_disabled.golden.yaml"},
		}),
		Entry("basic inbound cluster with connection limits and outlier detection", sidecarTestCase{
			resources: []*core_xds.Resource{
				{
					Name:     "inbound",
					Origin:   generator.OriginInbound,
					Resource: test_xds.ClusterWithName(fmt.Sprintf("localhost:%d", builders.FirstInboundPort)),
				},
			},
			fromRules: core_rules.FromRules{
				Rules: map[core_rules.InboundListener]core_rules.Rules{
					{
						Address: "127.0.0.1",
						Port:    builders.FirstInboundPort,
					}: []*core_rules.Rule{
						{
							Subset: core_rules.Subset{},
							Conf: api.Conf{
								ConnectionLimits: genConnectionLimits(),
								OutlierDetection: genOutlierDetection(false),
							},
						},
					},
				},
			},
			expectedCluster: []string{"inbound_cluster_connection_limits_outlier_detection.golden.yaml"},
		}),
		Entry("basic inbound cluster with connection limits, outlier detection and disabled=true", sidecarTestCase{
			resources: []*core_xds.Resource{
				{
					Name:     "inbound",
					Origin:   generator.OriginInbound,
					Resource: test_xds.ClusterWithName(fmt.Sprintf("localhost:%d", builders.FirstInboundPort)),
				},
			},
			fromRules: core_rules.FromRules{
				Rules: map[core_rules.InboundListener]core_rules.Rules{
					{
						Address: "127.0.0.1",
						Port:    builders.FirstInboundPort,
					}: []*core_rules.Rule{
						{
							Subset: core_rules.Subset{},
							Conf: api.Conf{
								ConnectionLimits: genConnectionLimits(),
								OutlierDetection: genOutlierDetection(true),
							},
						},
					},
				},
			},
			expectedCluster: []string{"inbound_cluster_connection_limits_outlier_detection_disabled.golden.yaml"},
		}),
		Entry("split outbound cluster with connection limits and outlier detection", sidecarTestCase{
			resources: []*core_xds.Resource{
				{
					Name:     "other-service",
					Origin:   generator.OriginOutbound,
					Resource: test_xds.ClusterWithName("other-service"),
				},
				{
					Name:     "other-service-5ab6003f460fabce",
					Origin:   generator.OriginOutbound,
					Resource: test_xds.ClusterWithName("other-service-5ab6003f460fabce"),
				},
			},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: core_rules.Subset{core_rules.Tag{
							Key:   mesh_proto.ServiceTag,
							Value: "other-service",
						}},
						Conf: api.Conf{
							ConnectionLimits: genConnectionLimits(),
							OutlierDetection: genOutlierDetection(false),
						},
					},
				},
			},
			expectedCluster: []string{
				"outbound_split_cluster_connection_limits_outlier_detection.golden.yaml",
				"outbound_split_cluster_0_connection_limits_outlier_detection.golden.yaml",
			},
		}),
		Entry("basic outbound cluster with connection limits targeting real MeshService", sidecarTestCase{
			resources: []*core_xds.Resource{
				{
					Name:           "outbound",
					Origin:         generator.OriginOutbound,
					Resource:       test_xds.ClusterWithName("backend"),
					ResourceOrigin: &backendMeshServiceIdentifier,
				},
			},
			toRules: core_rules.ToRules{
				ResourceRules: map[core_model.TypedResourceIdentifier]core_rules.ResourceRule{
					backendMeshServiceIdentifier: {
						Conf: []interface{}{
							api.Conf{
								ConnectionLimits: genConnectionLimits(),
								OutlierDetection: genOutlierDetection(false),
							},
						},
					},
				},
			},
			expectedCluster: []string{"outbound_cluster_connection_limits_real_resource.golden.yaml"},
		}),
	)

	type gatewayTestCase struct {
		name  string
		rules core_rules.GatewayRules
	}
	DescribeTable("should generate proper Envoy config for MeshGateways",
		func(given gatewayTestCase) {
			Expect(given.name).ToNot(BeEmpty())
			resources := xds_context.NewResources()
			resources.MeshLocalResources[core_mesh.MeshGatewayType] = &core_mesh.MeshGatewayResourceList{
				Items: []*core_mesh.MeshGatewayResource{samples.GatewayResource()},
			}
			resources.MeshLocalResources[core_mesh.MeshGatewayRouteType] = &core_mesh.MeshGatewayRouteResourceList{
				Items: []*core_mesh.MeshGatewayRouteResource{samples.BackendGatewayRoute()},
			}

			xdsCtx := *xds_builders.Context().
				WithMeshBuilder(samples.MeshDefaultBuilder()).
				WithResources(resources).
				AddServiceProtocol("backend", core_mesh.ProtocolHTTP).
				Build()
			proxy := xds_builders.Proxy().
				WithDataplane(samples.GatewayDataplaneBuilder()).
				WithPolicies(xds_builders.MatchedPolicies().WithGatewayPolicy(api.MeshCircuitBreakerType, given.rules)).
				Build()
			for n, p := range core_plugins.Plugins().ProxyPlugins() {
				Expect(p.Apply(context.Background(), xdsCtx.Mesh, proxy)).To(Succeed(), n)
			}
			gatewayGenerator := gateway_plugin.NewGenerator("test-zone")
			generatedResources, err := gatewayGenerator.Generate(context.Background(), nil, xdsCtx, proxy)
			Expect(err).NotTo(HaveOccurred())

			// when
			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)
			Expect(plugin.Apply(generatedResources, xdsCtx, proxy)).To(Succeed())

			getResourceYaml := func(list core_xds.ResourceList) []byte {
				actualResource, err := util_proto.ToYAML(list[0].Resource)
				Expect(err).ToNot(HaveOccurred())
				return actualResource
			}

			// then
			Expect(getResourceYaml(generatedResources.ListOf(envoy_resource.ClusterType))).
				To(matchers.MatchGoldenYAML(filepath.Join("testdata", fmt.Sprintf("%s.gateway_cluster.golden.yaml", given.name))))
		},
		Entry("basic outbound cluster with connection limits", gatewayTestCase{
			name: "basic",
			rules: core_rules.GatewayRules{
				ToRules: core_rules.GatewayToRules{
					ByListener: map[core_rules.InboundListener]core_rules.Rules{
						{Address: "192.168.0.1", Port: 8080}: {{
							Subset: core_rules.Subset{core_rules.Tag{
								Key:   mesh_proto.ServiceTag,
								Value: "backend",
							}},
							Conf: api.Conf{
								ConnectionLimits: genConnectionLimits(),
							},
						}},
					},
				},
			},
		}),
	)
})
