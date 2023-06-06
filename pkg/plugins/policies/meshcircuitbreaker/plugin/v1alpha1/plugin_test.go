package v1alpha1_test

import (
	"fmt"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	"path/filepath"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/intstr"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/xds"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshcircuitbreaker/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshcircuitbreaker/plugin/v1alpha1"
	gateway_plugin "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/pkg/test/matchers"
	test_matchers "github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	test_xds "github.com/kumahq/kuma/pkg/test/xds"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ = Describe("MeshCircuitBreaker", func() {
	type sidecarTestCase struct {
		resources       []*core_xds.Resource
		toRules         core_xds.ToRules
		fromRules       core_xds.FromRules
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

			context := xds_context.Context{
				Mesh: xds_context.MeshContext{
					Resource: &core_mesh.MeshResource{
						Meta: &test_model.ResourceMeta{
							Name: "default",
						},
					},
				},
			}

			proxy := core_xds.Proxy{
				Dataplane: builders.Dataplane().
					WithName("backend").
					WithMesh("default").
					WithAddress("127.0.0.1").
					AddOutboundsToServices("other-service", "second-service").
					WithInboundOfTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, "http").
					Build(),
				Policies: core_xds.MatchedPolicies{
					Dynamic: map[core_model.ResourceType]core_xds.TypedMatchingPolicies{
						api.MeshCircuitBreakerType: {
							Type:      api.MeshCircuitBreakerType,
							FromRules: given.fromRules,
							ToRules:   given.toRules,
						},
					},
				},
			}

			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)

			Expect(plugin.Apply(resourceSet, context, &proxy)).To(Succeed())

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
			toRules: core_xds.ToRules{
				Rules: []*rules.Rule{
					{
						Subset: rules.Subset{rules.Tag{
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
			toRules: core_xds.ToRules{
				Rules: []*rules.Rule{
					{
						Subset: rules.Subset{rules.Tag{
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
			toRules: core_xds.ToRules{
				Rules: []*rules.Rule{
					{
						Subset: rules.Subset{rules.Tag{
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
			toRules: core_xds.ToRules{
				Rules: []*rules.Rule{
					{
						Subset: rules.Subset{rules.Tag{
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
			toRules: core_xds.ToRules{
				Rules: []*rules.Rule{
					{
						Subset: rules.Subset{rules.Tag{
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
			fromRules: core_xds.FromRules{
				Rules: map[core_xds.InboundListener]rules.Rules{
					{
						Address: "127.0.0.1",
						Port:    builders.FirstInboundPort,
					}: []*rules.Rule{
						{
							Subset: rules.Subset{},
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
			fromRules: core_xds.FromRules{
				Rules: map[core_xds.InboundListener]rules.Rules{
					{
						Address: "127.0.0.1",
						Port:    builders.FirstInboundPort,
					}: []*rules.Rule{
						{
							Subset: rules.Subset{},
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
			fromRules: core_xds.FromRules{
				Rules: map[core_xds.InboundListener]rules.Rules{
					{
						Address: "127.0.0.1",
						Port:    builders.FirstInboundPort,
					}: []*rules.Rule{
						{
							Subset: rules.Subset{},
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
			fromRules: core_xds.FromRules{
				Rules: map[core_xds.InboundListener]rules.Rules{
					{
						Address: "127.0.0.1",
						Port:    builders.FirstInboundPort,
					}: []*rules.Rule{
						{
							Subset: rules.Subset{},
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
			fromRules: core_xds.FromRules{
				Rules: map[core_xds.InboundListener]rules.Rules{
					{
						Address: "127.0.0.1",
						Port:    builders.FirstInboundPort,
					}: []*rules.Rule{
						{
							Subset: rules.Subset{},
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
					Name:     "other-service-_0_",
					Origin:   generator.OriginOutbound,
					Resource: test_xds.ClusterWithName("other-service-_0_"),
				},
			},
			toRules: core_xds.ToRules{
				Rules: []*rules.Rule{
					{
						Subset: rules.Subset{rules.Tag{
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
	)

	type gatewayTestCase struct {
		name    string
		toRules core_xds.ToRules
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

			context := test_xds.CreateSampleMeshContextWith(resources)
			proxy := xds.Proxy{
				APIVersion: "v3",
				Dataplane:  samples.GatewayDataplane(),
				Policies: xds.MatchedPolicies{
					Dynamic: map[core_model.ResourceType]xds.TypedMatchingPolicies{
						api.MeshCircuitBreakerType: {
							Type:    api.MeshCircuitBreakerType,
							ToRules: given.toRules,
						},
					},
				},
			}
			gatewayGenerator := gateway_plugin.NewGenerator("test-zone")
			generatedResources, err := gatewayGenerator.Generate(context, &proxy)
			Expect(err).NotTo(HaveOccurred())

			// when
			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)
			Expect(plugin.Apply(generatedResources, context, &proxy)).To(Succeed())

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
			toRules: core_xds.ToRules{
				Rules: []*rules.Rule{
					{
						Subset: rules.Subset{rules.Tag{
							Key:   mesh_proto.ServiceTag,
							Value: "backend",
						}},
						Conf: api.Conf{
							ConnectionLimits: genConnectionLimits(),
						},
					},
				},
			},
		}),
	)
})
