package v1alpha1_test

import (
	"fmt"
	"path/filepath"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshcircuitbreaker/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshcircuitbreaker/plugin/v1alpha1"
	"github.com/kumahq/kuma/pkg/test"
	test_matchers "github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	test_xds "github.com/kumahq/kuma/pkg/test/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ = Describe("MeshCircuitBreaker", func() {
	type sidecarTestCase struct {
		resources       []*core_xds.Resource
		toRules         core_xds.ToRules
		fromRules       core_xds.FromRules
		expectedCluster string
	}

	genConnectionLimits := func() *api.ConnectionLimits {
		return &api.ConnectionLimits{
			MaxConnectionPools: test.PointerUint32(1111),
			MaxConnections:     test.PointerUint32(2222),
			MaxPendingRequests: test.PointerUint32(3333),
			MaxRequests:        test.PointerUint32(4444),
			MaxRetries:         test.PointerUint32(5555),
		}
	}

	genOutlierDetection := func(disabled bool) *api.OutlierDetection {
		return &api.OutlierDetection{
			Disabled:                    test.PointerBool(disabled),
			Interval:                    test.ParseDuration("10s"),
			BaseEjectionTime:            test.ParseDuration("8s"),
			MaxEjectionPercent:          test.PointerUint32(88),
			SplitExternalAndLocalErrors: test.PointerBool(true),
			Detectors: &api.Detectors{
				TotalFailures: &api.DetectorTotalFailures{
					Consecutive: test.PointerUint32(12),
				},
				GatewayFailures: &api.DetectorGatewayFailures{
					Consecutive: test.PointerUint32(91),
				},
				LocalOriginFailures: &api.DetectorLocalOriginFailures{
					Consecutive: test.PointerUint32(3),
				},
				SuccessRate: &api.DetectorSuccessRateFailures{
					MinimumHosts:            test.PointerUint32(33),
					RequestVolume:           test.PointerUint32(99),
					StandardDeviationFactor: test.PointerUint32(1900),
				},
				FailurePercentage: &api.DetectorFailurePercentageFailures{
					MinimumHosts:  test.PointerUint32(32),
					RequestVolume: test.PointerUint32(182),
					Threshold:     test.PointerUint32(80),
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

			Expect(util_proto.ToYAML(resourceSet.ListOf(envoy_resource.ClusterType)[0].Resource)).
				To(test_matchers.MatchGoldenYAML(filepath.Join("testdata", given.expectedCluster)))
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
				Rules: []*core_xds.Rule{
					{
						Subset: core_xds.Subset{core_xds.Tag{
							Key:   mesh_proto.ServiceTag,
							Value: "other-service",
						}},
						Conf: api.Conf{
							ConnectionLimits: genConnectionLimits(),
						},
					},
				},
			},
			expectedCluster: "outbound_cluster_connection_limits.golden.yaml",
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
				Rules: []*core_xds.Rule{
					{
						Subset: core_xds.Subset{core_xds.Tag{
							Key:   mesh_proto.ServiceTag,
							Value: "second-service",
						}},
						Conf: api.Conf{
							OutlierDetection: genOutlierDetection(false),
						},
					},
				},
			},
			expectedCluster: "outbound_cluster_outlier_detection.golden.yaml",
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
				Rules: []*core_xds.Rule{
					{
						Subset: core_xds.Subset{core_xds.Tag{
							Key:   mesh_proto.ServiceTag,
							Value: "second-service",
						}},
						Conf: api.Conf{
							OutlierDetection: genOutlierDetection(true),
						},
					},
				},
			},
			expectedCluster: "outbound_cluster_outlier_detection_disabled.golden.yaml",
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
				Rules: []*core_xds.Rule{
					{
						Subset: core_xds.Subset{core_xds.Tag{
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
			expectedCluster: "outbound_cluster_connection_limits_outlier_detection.golden.yaml",
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
				Rules: []*core_xds.Rule{
					{
						Subset: core_xds.Subset{core_xds.Tag{
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
			expectedCluster: "outbound_cluster_connection_limits_outlier_detection_disabled.golden.yaml",
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
				Rules: map[core_xds.InboundListener]core_xds.Rules{
					{
						Address: "127.0.0.1",
						Port:    builders.FirstInboundPort,
					}: []*core_xds.Rule{
						{
							Subset: core_xds.Subset{},
							Conf: api.Conf{
								ConnectionLimits: genConnectionLimits(),
							},
						},
					},
				},
			},
			expectedCluster: "inbound_cluster_connection_limits.golden.yaml",
		}),
		Entry("basic inbound cluster with outlier detection", sidecarTestCase{
			resources: []*core_xds.Resource{
				{
					Name:     "inbound",
					Origin:   generator.OriginInbound,
					Resource: test_xds.ClusterWithName(fmt.Sprintf("localhost:%d", builders.FirstInboundPort)),
				},
			},
			fromRules: core_xds.FromRules{
				Rules: map[core_xds.InboundListener]core_xds.Rules{
					{
						Address: "127.0.0.1",
						Port:    builders.FirstInboundPort,
					}: []*core_xds.Rule{
						{
							Subset: core_xds.Subset{},
							Conf: api.Conf{
								OutlierDetection: genOutlierDetection(false),
							},
						},
					},
				},
			},
			expectedCluster: "inbound_cluster_outlier_detection.golden.yaml",
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
				Rules: map[core_xds.InboundListener]core_xds.Rules{
					{
						Address: "127.0.0.1",
						Port:    builders.FirstInboundPort,
					}: []*core_xds.Rule{
						{
							Subset: core_xds.Subset{},
							Conf: api.Conf{
								OutlierDetection: genOutlierDetection(true),
							},
						},
					},
				},
			},
			expectedCluster: "inbound_cluster_outlier_detection_disabled.golden.yaml",
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
				Rules: map[core_xds.InboundListener]core_xds.Rules{
					{
						Address: "127.0.0.1",
						Port:    builders.FirstInboundPort,
					}: []*core_xds.Rule{
						{
							Subset: core_xds.Subset{},
							Conf: api.Conf{
								ConnectionLimits: genConnectionLimits(),
								OutlierDetection: genOutlierDetection(false),
							},
						},
					},
				},
			},
			expectedCluster: "inbound_cluster_connection_limits_outlier_detection.golden.yaml",
		}),
		Entry("basic outbound cluster with connection limits, outlier detection and disabled=true", sidecarTestCase{
			resources: []*core_xds.Resource{
				{
					Name:     "inbound",
					Origin:   generator.OriginInbound,
					Resource: test_xds.ClusterWithName(fmt.Sprintf("localhost:%d", builders.FirstInboundPort)),
				},
			},
			fromRules: core_xds.FromRules{
				Rules: map[core_xds.InboundListener]core_xds.Rules{
					{
						Address: "127.0.0.1",
						Port:    builders.FirstInboundPort,
					}: []*core_xds.Rule{
						{
							Subset: core_xds.Subset{},
							Conf: api.Conf{
								ConnectionLimits: genConnectionLimits(),
								OutlierDetection: genOutlierDetection(true),
							},
						},
					},
				},
			},
			expectedCluster: "inbound_cluster_connection_limits_outlier_detection_disabled.golden.yaml",
		}),
	)
})
