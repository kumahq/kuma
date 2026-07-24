//nolint:staticcheck // SA1019 Test file: tests backward compatibility with deprecated core_rules.Rule
package v1alpha1_test

import (
	"path/filepath"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/intstr"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	"github.com/kumahq/kuma/v3/pkg/core/naming"
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
	core_rules "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/inbound"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/outbound"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/subsetutils"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshcircuitbreaker/api/v1alpha1"
	plugin "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshcircuitbreaker/plugin/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/test"
	test_matchers "github.com/kumahq/kuma/v3/pkg/test/matchers"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	"github.com/kumahq/kuma/v3/pkg/test/resources/samples"
	test_xds "github.com/kumahq/kuma/v3/pkg/test/xds"
	xds_builders "github.com/kumahq/kuma/v3/pkg/test/xds/builders"
	xds_samples "github.com/kumahq/kuma/v3/pkg/test/xds/samples"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	envoy_names "github.com/kumahq/kuma/v3/pkg/xds/envoy/names"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/metadata"
)

var _ = Describe("MeshCircuitBreaker", func() {
	backendMeshServiceIdentifier := kri.Identifier{
		ResourceType: "MeshService",
		Mesh:         "default",
		Name:         "backend",
		SectionName:  "",
	}

	type sidecarTestCase struct {
		resources       []*core_xds.Resource
		toRules         core_rules.ToRules
		fromRules       core_rules.FromRules
		withoutPolicy   bool
		unifiedNaming   bool
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
			HealthyPanicThreshold: pointer.To(intstr.FromString("30.2")),
		}
	}

	clusterWithExistingCircuitBreakers := func(name string) *envoy_cluster.Cluster {
		cluster := test_xds.ClusterWithName(name).(*envoy_cluster.Cluster)
		cluster.CircuitBreakers = &envoy_cluster.CircuitBreakers{
			Thresholds: []*envoy_cluster.CircuitBreakers_Thresholds{{
				Priority:       envoy_config_core_v3.RoutingPriority_DEFAULT,
				MaxConnections: util_proto.UInt32(66),
			}},
		}
		return cluster
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
				})
			if given.unifiedNaming {
				proxy = proxy.WithMetadata(&core_xds.DataplaneMetadata{
					Features: xds_types.Features{xds_types.FeatureUnifiedResourceNaming: true},
				})
			}
			if !given.withoutPolicy {
				proxy = proxy.WithPolicies(
					xds_builders.MatchedPolicies().WithPolicy(api.MeshCircuitBreakerType, given.toRules, given.fromRules),
				)
			}
			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)
			Expect(plugin.Apply(resourceSet, context, proxy.Build())).To(Succeed())

			for idx, expected := range given.expectedCluster {
				Expect(util_proto.ToYAML(resourceSet.ListOf(envoy_resource.ClusterType)[idx].Resource)).
					To(test_matchers.MatchGoldenYAML(filepath.Join("testdata", expected)))
			}
		},
		Entry("basic outbound cluster with connection limits", sidecarTestCase{
			resources: []*core_xds.Resource{
				{
					Name:     "outbound",
					Origin:   metadata.OriginOutbound,
					Resource: test_xds.ClusterWithName("other-service"),
				},
			},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: subsetutils.Subset{subsetutils.Tag{
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
		Entry("basic outbound cluster without MeshCircuitBreaker", sidecarTestCase{
			resources: []*core_xds.Resource{
				{
					Name:     "outbound",
					Origin:   metadata.OriginOutbound,
					Resource: test_xds.ClusterWithName("other-service"),
				},
			},
			withoutPolicy:   true,
			expectedCluster: []string{"outbound_cluster_track_remaining_only.golden.yaml"},
		}),
		Entry("basic outbound cluster preserves existing circuit breakers without MeshCircuitBreaker", sidecarTestCase{
			resources: []*core_xds.Resource{
				{
					Name:     "outbound",
					Origin:   metadata.OriginOutbound,
					Resource: clusterWithExistingCircuitBreakers("other-service"),
				},
			},
			withoutPolicy:   true,
			expectedCluster: []string{"outbound_cluster_existing_circuit_breakers.golden.yaml"},
		}),
		Entry("basic outbound cluster with outlier detection", sidecarTestCase{
			resources: []*core_xds.Resource{
				{
					Name:     "outbound",
					Origin:   metadata.OriginOutbound,
					Resource: test_xds.ClusterWithName("second-service"),
				},
			},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: subsetutils.Subset{subsetutils.Tag{
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
					Origin:   metadata.OriginOutbound,
					Resource: test_xds.ClusterWithName("second-service"),
				},
			},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: subsetutils.Subset{subsetutils.Tag{
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
					Origin:   metadata.OriginOutbound,
					Resource: test_xds.ClusterWithName("second-service"),
				},
			},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: subsetutils.Subset{subsetutils.Tag{
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
					Origin:   metadata.OriginOutbound,
					Resource: test_xds.ClusterWithName("second-service"),
				},
			},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: subsetutils.Subset{subsetutils.Tag{
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
					Origin:   metadata.OriginInbound,
					Resource: test_xds.ClusterWithName(envoy_names.GetInboundClusterName(builders.FirstInboundServicePort, builders.FirstInboundPort)),
				},
			},
			fromRules: core_rules.FromRules{
				Rules: map[core_rules.InboundListener]core_rules.Rules{
					{Address: "127.0.0.1", Port: builders.FirstInboundPort}: []*core_rules.Rule{
						{
							Subset: subsetutils.Subset{},
							Conf: api.Conf{
								ConnectionLimits: genConnectionLimits(),
							},
						},
					},
				},
				InboundRules: map[core_rules.InboundListener][]*inbound.Rule{
					{Address: "127.0.0.1", Port: builders.FirstInboundPort}: {{
						Conf: api.Conf{ConnectionLimits: genConnectionLimits()},
					}},
				},
			},
			expectedCluster: []string{"inbound_cluster_connection_limits.golden.yaml"},
		}),
		Entry("basic inbound cluster with connection limits (unified naming)", sidecarTestCase{
			unifiedNaming: true,
			resources: []*core_xds.Resource{
				{
					Name:     "inbound",
					Origin:   metadata.OriginInbound,
					Resource: test_xds.ClusterWithName(naming.MustContextualInboundName(core_mesh.NewDataplaneResource(), builders.FirstInboundPort)),
				},
			},
			fromRules: core_rules.FromRules{
				Rules: map[core_rules.InboundListener]core_rules.Rules{
					{Address: "127.0.0.1", Port: builders.FirstInboundPort}: []*core_rules.Rule{
						{
							Subset: subsetutils.Subset{},
							Conf: api.Conf{
								ConnectionLimits: genConnectionLimits(),
							},
						},
					},
				},
				InboundRules: map[core_rules.InboundListener][]*inbound.Rule{
					{Address: "127.0.0.1", Port: builders.FirstInboundPort}: {{
						Conf: api.Conf{ConnectionLimits: genConnectionLimits()},
					}},
				},
			},
			expectedCluster: []string{"inbound_cluster_connection_limits_unified_naming.golden.yaml"},
		}),
		Entry("basic inbound cluster with outlier detection", sidecarTestCase{
			resources: []*core_xds.Resource{
				{
					Name:     "inbound",
					Origin:   metadata.OriginInbound,
					Resource: test_xds.ClusterWithName(envoy_names.GetInboundClusterName(builders.FirstInboundServicePort, builders.FirstInboundPort)),
				},
			},
			fromRules: core_rules.FromRules{
				Rules: map[core_rules.InboundListener]core_rules.Rules{
					{Address: "127.0.0.1", Port: builders.FirstInboundPort}: []*core_rules.Rule{
						{
							Subset: subsetutils.Subset{},
							Conf:   api.Conf{OutlierDetection: genOutlierDetection(false)},
						},
					},
				},
				InboundRules: map[core_rules.InboundListener][]*inbound.Rule{
					{Address: "127.0.0.1", Port: builders.FirstInboundPort}: {{
						Conf: api.Conf{OutlierDetection: genOutlierDetection(false)},
					}},
				},
			},
			expectedCluster: []string{"inbound_cluster_outlier_detection.golden.yaml"},
		}),
		Entry("basic inbound cluster with outlier detection and disabled=true", sidecarTestCase{
			resources: []*core_xds.Resource{
				{
					Name:     "inbound",
					Origin:   metadata.OriginInbound,
					Resource: test_xds.ClusterWithName(envoy_names.GetInboundClusterName(builders.FirstInboundServicePort, builders.FirstInboundPort)),
				},
			},
			fromRules: core_rules.FromRules{
				Rules: map[core_rules.InboundListener]core_rules.Rules{
					{Address: "127.0.0.1", Port: builders.FirstInboundPort}: []*core_rules.Rule{
						{
							Subset: subsetutils.Subset{},
							Conf: api.Conf{
								OutlierDetection: genOutlierDetection(true),
							},
						},
					},
				},
				InboundRules: map[core_rules.InboundListener][]*inbound.Rule{
					{Address: "127.0.0.1", Port: builders.FirstInboundPort}: {{
						Conf: api.Conf{
							OutlierDetection: genOutlierDetection(true),
						},
					}},
				},
			},
			expectedCluster: []string{"inbound_cluster_outlier_detection_disabled.golden.yaml"},
		}),
		Entry("basic inbound cluster with connection limits and outlier detection", sidecarTestCase{
			resources: []*core_xds.Resource{
				{
					Name:     "inbound",
					Origin:   metadata.OriginInbound,
					Resource: test_xds.ClusterWithName(envoy_names.GetInboundClusterName(builders.FirstInboundServicePort, builders.FirstInboundPort)),
				},
			},
			fromRules: core_rules.FromRules{
				Rules: map[core_rules.InboundListener]core_rules.Rules{
					{Address: "127.0.0.1", Port: builders.FirstInboundPort}: []*core_rules.Rule{
						{
							Subset: subsetutils.Subset{},
							Conf: api.Conf{
								ConnectionLimits: genConnectionLimits(),
								OutlierDetection: genOutlierDetection(false),
							},
						},
					},
				},
				InboundRules: map[core_rules.InboundListener][]*inbound.Rule{
					{Address: "127.0.0.1", Port: builders.FirstInboundPort}: {{
						Conf: api.Conf{
							ConnectionLimits: genConnectionLimits(),
							OutlierDetection: genOutlierDetection(false),
						},
					}},
				},
			},
			expectedCluster: []string{"inbound_cluster_connection_limits_outlier_detection.golden.yaml"},
		}),
		Entry("basic inbound cluster with connection limits, outlier detection and disabled=true", sidecarTestCase{
			resources: []*core_xds.Resource{
				{
					Name:     "inbound",
					Origin:   metadata.OriginInbound,
					Resource: test_xds.ClusterWithName(envoy_names.GetInboundClusterName(builders.FirstInboundServicePort, builders.FirstInboundPort)),
				},
			},
			fromRules: core_rules.FromRules{
				Rules: map[core_rules.InboundListener]core_rules.Rules{
					{Address: "127.0.0.1", Port: builders.FirstInboundPort}: []*core_rules.Rule{
						{
							Subset: subsetutils.Subset{},
							Conf: api.Conf{
								ConnectionLimits: genConnectionLimits(),
								OutlierDetection: genOutlierDetection(true),
							},
						},
					},
				},
				InboundRules: map[core_rules.InboundListener][]*inbound.Rule{
					{Address: "127.0.0.1", Port: builders.FirstInboundPort}: {{
						Conf: api.Conf{
							ConnectionLimits: genConnectionLimits(),
							OutlierDetection: genOutlierDetection(true),
						},
					}},
				},
			},
			expectedCluster: []string{"inbound_cluster_connection_limits_outlier_detection_disabled.golden.yaml"},
		}),
		Entry("split outbound cluster with connection limits and outlier detection", sidecarTestCase{
			resources: []*core_xds.Resource{
				{
					Name:     "other-service",
					Origin:   metadata.OriginOutbound,
					Resource: test_xds.ClusterWithName("other-service"),
				},
				{
					Name:     "other-service-5ab6003f460fabce",
					Origin:   metadata.OriginOutbound,
					Resource: test_xds.ClusterWithName("other-service-5ab6003f460fabce"),
				},
			},
			toRules: core_rules.ToRules{
				Rules: []*core_rules.Rule{
					{
						Subset: subsetutils.Subset{subsetutils.Tag{
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
					Origin:         metadata.OriginOutbound,
					Resource:       test_xds.ClusterWithName("backend"),
					ResourceOrigin: backendMeshServiceIdentifier,
				},
			},
			toRules: core_rules.ToRules{
				ResourceRules: map[kri.Identifier]outbound.ResourceRule{
					backendMeshServiceIdentifier: {
						Conf: []any{
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

	type zoneProxyTestCase struct {
		resources       []*core_xds.Resource
		toRules         core_rules.ToRules
		withoutPolicy   bool
		expectedCluster string
	}

	mesKRI := kri.Identifier{
		ResourceType: meshexternalservice_api.MeshExternalServiceType,
		Mesh:         "default",
		Name:         "example",
		SectionName:  "9000",
	}

	DescribeTable("should generate proper Envoy config for mesh-scoped zone proxy",
		func(given zoneProxyTestCase) {
			resourceSet := core_xds.NewResourceSet()
			resourceSet.Add(given.resources...)

			mes := builders.MeshExternalService().Build()

			xdsCtx := *xds_builders.Context().
				WithMeshBuilder(samples.MeshDefaultBuilder()).
				WithMeshLocalResources([]core_model.Resource{mes}).
				Build()

			proxy := xds_builders.Proxy().
				With(func(p *core_xds.Proxy) {
					p.Dataplane = &core_mesh.DataplaneResource{
						Meta: &test_model.ResourceMeta{Name: "zone-proxy", Mesh: "default"},
						Spec: &mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "10.0.0.1",
								Listeners: []*mesh_proto.Dataplane_Networking_Listener{{
									Type:    mesh_proto.Dataplane_Networking_Listener_ZoneEgress,
									Address: "10.0.0.1",
									Port:    10002,
									Name:    "ze-port",
									State:   mesh_proto.Dataplane_Networking_Listener_Ready,
								}},
							},
						},
					}
				})
			if !given.withoutPolicy {
				proxy = proxy.WithPolicies(
					xds_builders.MatchedPolicies().WithPolicy(api.MeshCircuitBreakerType, given.toRules, core_rules.FromRules{}),
				)
			}

			p := plugin.NewPlugin().(core_plugins.PolicyPlugin)
			Expect(p.Apply(resourceSet, xdsCtx, proxy.Build())).To(Succeed())

			Expect(util_proto.ToYAML(resourceSet.ListOf(envoy_resource.ClusterType)[0].Resource)).
				To(test_matchers.MatchGoldenYAML(filepath.Join("testdata", given.expectedCluster)))
		},
		Entry("MeshExternalService cluster with connection limits on zone proxy", zoneProxyTestCase{
			resources: []*core_xds.Resource{{
				Name:           mesKRI.String(),
				Origin:         metadata.OriginEgress,
				Resource:       test_xds.ClusterWithName(mesKRI.String()),
				ResourceOrigin: mesKRI,
			}},
			toRules: core_rules.ToRules{
				ResourceRules: outbound.ResourceRules{
					mesKRI: {
						Conf: []any{api.Conf{ConnectionLimits: genConnectionLimits()}},
					},
				},
			},
			expectedCluster: "basic-meshexternalservice.zone_proxy_cluster.golden.yaml",
		}),
		Entry("MeshExternalService cluster without MeshCircuitBreaker on zone proxy", zoneProxyTestCase{
			resources: []*core_xds.Resource{{
				Name:           mesKRI.String(),
				Origin:         metadata.OriginEgress,
				Resource:       test_xds.ClusterWithName(mesKRI.String()),
				ResourceOrigin: mesKRI,
			}},
			withoutPolicy:   true,
			expectedCluster: "basic-meshexternalservice-track-remaining.zone_proxy_cluster.golden.yaml",
		}),
		Entry("MeshExternalService cluster with outlier detection on zone proxy", zoneProxyTestCase{
			resources: []*core_xds.Resource{{
				Name:           mesKRI.String(),
				Origin:         metadata.OriginEgress,
				Resource:       test_xds.ClusterWithName(mesKRI.String()),
				ResourceOrigin: mesKRI,
			}},
			toRules: core_rules.ToRules{
				ResourceRules: outbound.ResourceRules{
					mesKRI: {
						Conf: []any{api.Conf{OutlierDetection: genOutlierDetection(false)}},
					},
				},
			},
			expectedCluster: "basic-meshexternalservice-outlier.zone_proxy_cluster.golden.yaml",
		}),
	)

	It("should enable track remaining on clusters regardless of origin", func() {
		resourceSet := core_xds.NewResourceSet()
		resourceSet.Add(&core_xds.Resource{
			Name:     "prometheus",
			Origin:   metadata.OriginPrometheus,
			Resource: test_xds.ClusterWithName("prometheus"),
		})

		p := plugin.NewPlugin().(core_plugins.PolicyPlugin)
		Expect(p.Apply(resourceSet, xds_samples.SampleContext(), xds_builders.Proxy().Build())).To(Succeed())

		cluster := resourceSet.ListOf(envoy_resource.ClusterType)[0].Resource.(*envoy_cluster.Cluster)
		Expect(cluster.GetCircuitBreakers().GetThresholds()).To(HaveLen(1))
		Expect(cluster.GetCircuitBreakers().GetThresholds()[0].GetPriority()).To(Equal(envoy_config_core_v3.RoutingPriority_DEFAULT))
		Expect(cluster.GetCircuitBreakers().GetThresholds()[0].TrackRemaining).To(BeTrue())
	})

	It("should enable track remaining for ZoneEgressProxy without MeshCircuitBreaker", func() {
		resourceSet := core_xds.NewResourceSet()
		resourceSet.Add(&core_xds.Resource{
			Name:           mesKRI.String(),
			Origin:         metadata.OriginEgress,
			Resource:       test_xds.ClusterWithName(mesKRI.String()),
			ResourceOrigin: mesKRI,
		})

		proxy := xds_builders.Proxy().
			With(func(p *core_xds.Proxy) {
				p.ZoneEgressProxy = &core_xds.ZoneEgressProxy{
					ZoneEgressResource: &core_mesh.ZoneEgressResource{
						Meta: &test_model.ResourceMeta{Name: "zone-egress", Mesh: "default"},
						Spec: &mesh_proto.ZoneEgress{
							Networking: &mesh_proto.ZoneEgress_Networking{
								Address: "10.0.0.1",
								Port:    10002,
							},
						},
					},
				}
			}).
			Build()

		p := plugin.NewPlugin().(core_plugins.PolicyPlugin)
		Expect(p.Apply(resourceSet, xds_samples.SampleContext(), proxy)).To(Succeed())

		cluster := resourceSet.ListOf(envoy_resource.ClusterType)[0].Resource.(*envoy_cluster.Cluster)
		Expect(cluster.GetCircuitBreakers().GetThresholds()).To(HaveLen(1))
		Expect(cluster.GetCircuitBreakers().GetThresholds()[0].GetPriority()).To(Equal(envoy_config_core_v3.RoutingPriority_DEFAULT))
		Expect(cluster.GetCircuitBreakers().GetThresholds()[0].TrackRemaining).To(BeTrue())
	})
})
