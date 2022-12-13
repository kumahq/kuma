package v1alpha1_test

import (
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/xds"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhealthcheck/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshhealthcheck/plugin/v1alpha1"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/xds"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/generator"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MeshHealthCheck", func() {
	echoServiceTag := "echo"
	type testCase struct {
		resources         []core_xds.Resource
		toRules           core_xds.ToRules
		expectedClusters  []string
	}
	outbounds := []core_xds.Resource{
		{
			Name:   "outbound",
			Origin: generator.OriginOutbound,
			Resource: NewListenerBuilder(envoy_common.APIV3).
				Configure(OutboundListener("outbound:127.0.0.1:27777", "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP)).
				Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
					Configure(HttpConnectionManager("127.0.0.1:27777", false)),
				)).MustBuild(),
		},
		{
			Name: "cluster",
			Origin: generator.OriginOutbound,
			Resource: clusters.NewClusterBuilder(envoy_common.APIV3).
				Configure(policies_xds.WithName(echoServiceTag)).
				MustBuild(),
		},
	}
	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			resources := core_xds.NewResourceSet()
			for _, res := range given.resources {
				r := res
				resources.Add(&r)
			}

			context := xds_context.Context{}
			proxy := xds.Proxy{
				APIVersion: envoy_common.APIV3,
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Mesh: "default",
						Name: "backend",
					},
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{
									Address: "127.0.0.1",
									Port:    27777,
									Tags: map[string]string{
										mesh_proto.ServiceTag: echoServiceTag,
										mesh_proto.ProtocolTag: "http",
									},
								},
							},
						},
					},
				},
				Policies: xds.MatchedPolicies{
					Dynamic: map[core_model.ResourceType]xds.TypedMatchingPolicies{
						api.MeshHealthCheckType: {
							Type:    api.MeshHealthCheckType,
							ToRules: given.toRules,
						},
					},
				},
			}
			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)

			Expect(plugin.Apply(resources, context, &proxy)).To(Succeed())
			policies_xds.ResourceArrayShouldEqual(resources.ListOf(envoy_resource.ClusterType), given.expectedClusters)
		},
		Entry("http hc", testCase{
			resources: outbounds,
			toRules: core_xds.ToRules{
				Rules: []*core_xds.Rule{
					{
						Subset: core_xds.Subset{},
						Conf: api.Conf{
							Interval: *policies_xds.ParseDuration("10s"),
							Timeout: *policies_xds.ParseDuration("2s"),
							UnhealthyThreshold: 3,
							HealthyThreshold: 1,
							Http: &api.HttpHealthCheck{
								Path: "/health",
							},
						},
					},
				}},
			expectedClusters: []string{`
name: echo
healthChecks:
- healthyThreshold: 1
  httpHealthCheck:
    path: /health
  interval: 10s
  timeout: 2s
  unhealthyThreshold: 3
`},
		}),
	)
})
