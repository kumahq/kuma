package v1alpha1_test

import (
	"context"
	"fmt"
	"os"
	"path"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/intstr"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshhttproute_plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/plugin/v1alpha1"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtls/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshtls/plugin/v1alpha1"
	gateway_plugin "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/pkg/test/xds/builders"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	"github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

func getResource(
	resourceSet *core_xds.ResourceSet,
	typ envoy_resource.Type,
) []byte {
	resources, err := resourceSet.ListOf(typ).ToDeltaDiscoveryResponse()
	Expect(err).ToNot(HaveOccurred())
	actual, err := util_proto.ToYAML(resources)
	Expect(err).ToNot(HaveOccurred())

	return actual
}

var _ = Describe("MeshTLS", func() {
	type testCase struct {
		caseName    string
		meshBuilder *builders.MeshBuilder
		meshService bool
	}
	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// given
			mesh := given.meshBuilder
			context := *xds_builders.Context().
				WithMeshBuilder(mesh).
				Build()
			resourceSet := core_xds.NewResourceSet()
			secretsTracker := envoy_common.NewSecretsTracker("default", nil)
			if given.meshService {
				resourceSet.Add(getMeshServiceResources(secretsTracker, mesh)...)
			} else {
				resourceSet.Add(getResources(secretsTracker, mesh)...)
			}

			policy := getPolicy(given.caseName)

			proxy := xds_builders.Proxy().
				WithSecretsTracker(secretsTracker).
				WithApiVersion(envoy_common.APIV3).
				WithOutbounds(xds_types.Outbounds{&xds_types.Outbound{
					LegacyOutbound: builders.Outbound().
						WithService("outgoing").
						WithAddress("127.0.0.1").
						WithPort(27777).Build(),
				}}).
				WithDataplane(
					builders.Dataplane().
						WithName("test").
						WithMesh("default").
						WithAddress("127.0.0.1").
						WithTransparentProxying(15006, 15001, "ipv4").
						AddOutbound(
							builders.Outbound().
								WithAddress("127.0.0.1").
								WithPort(27777).
								WithService("outgoing"),
						).
						AddInbound(
							builders.Inbound().
								WithAddress("127.0.0.1").
								WithPort(17777).
								WithService("backend"),
						).
						AddInbound(
							builders.Inbound().
								WithAddress("127.0.0.1").
								WithPort(17778).
								WithService("frontend"),
						),
				).
				WithPolicies(xds_builders.MatchedPolicies().WithFromPolicy(api.MeshTLSType, getFromRules(policy.Spec.From))).
				Build()

			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)

			// when
			Expect(plugin.Apply(resourceSet, context, proxy)).To(Succeed())

			// then
			Expect(getResource(resourceSet, envoy_resource.ListenerType)).
				To(matchers.MatchGoldenYAML(fmt.Sprintf("testdata/%s.listeners.golden.yaml", given.caseName)))
			Expect(getResource(resourceSet, envoy_resource.ClusterType)).
				To(matchers.MatchGoldenYAML(fmt.Sprintf("testdata/%s.clusters.golden.yaml", given.caseName)))
		},
		Entry("strict with no mTLS on the mesh", testCase{
			caseName:    "strict-no-mtls",
			meshBuilder: samples.MeshDefaultBuilder(),
		}),
		Entry("permissive with no mTLS on the mesh", testCase{
			caseName:    "permissive-no-mtls",
			meshBuilder: samples.MeshDefaultBuilder(),
		}),
		Entry("strict with permissive mTLS on the mesh", testCase{
			caseName:    "strict-with-permissive-mtls",
			meshBuilder: samples.MeshMTLSBuilder().WithPermissiveMTLSBackends(),
		}),
		Entry("permissive with permissive mTLS on the mesh", testCase{
			caseName:    "permissive-with-permissive-mtls",
			meshBuilder: samples.MeshMTLSBuilder().WithPermissiveMTLSBackends(),
		}),
		Entry("strict with permissive mTLS on the mesh for MeshService", testCase{
			caseName:    "strict-with-permissive-mtls-meshservice",
			meshBuilder: samples.MeshMTLSBuilder().WithPermissiveMTLSBackends(),
			meshService: true,
		}),
	)

	DescribeTable("should generate proper Envoy config for builtin Gateway",
		func(given testCase) {
			secretsTracker := envoy_common.NewSecretsTracker("default", nil)
			resources := xds_context.NewResources()
			resources.MeshLocalResources[core_mesh.MeshGatewayType] = &core_mesh.MeshGatewayResourceList{
				Items: []*core_mesh.MeshGatewayResource{samples.GatewayResource()},
			}
			backendRefOriginIndex := map[common_api.MatchesHash]int{}
			var backendRef common_api.BackendRef
			if given.meshService {
				backendRefOriginIndex = map[common_api.MatchesHash]int{
					meshhttproute_api.HashMatches([]meshhttproute_api.Match{
						{Path: &meshhttproute_api.PathMatch{Type: meshhttproute_api.Exact, Value: "/"}},
					}): 0,
				}
				meshSvc := meshservice_api.MeshServiceResource{
					Meta: &test_model.ResourceMeta{Name: "backend", Mesh: "default"},
					Spec: &meshservice_api.MeshService{
						Selector: meshservice_api.Selector{},
						Ports: []meshservice_api.Port{{
							Name:        "test-port",
							Port:        80,
							TargetPort:  intstr.FromInt(8084),
							AppProtocol: core_mesh.ProtocolHTTP,
						}},
						Identities: []meshservice_api.MeshServiceIdentity{
							{
								Type:  meshservice_api.MeshServiceIdentityServiceTagType,
								Value: "backend",
							},
						},
					},
					Status: &meshservice_api.MeshServiceStatus{
						VIPs: []meshservice_api.VIP{{
							IP: "10.0.0.1",
						}},
					},
				}
				resources.MeshLocalResources[meshservice_api.MeshServiceType] = &meshservice_api.MeshServiceResourceList{
					Items: []*meshservice_api.MeshServiceResource{&meshSvc},
				}
				backendRef = common_api.BackendRef{
					TargetRef: builders.TargetRefMeshService("backend", "", ""),
					Port:      pointer.To[uint32](80),
					Weight:    pointer.To(uint(100)),
				}
			} else {
				backendRef = common_api.BackendRef{
					TargetRef: builders.TargetRefMeshService("backend", "", ""),
					Weight:    pointer.To(uint(100)),
				}
			}

			policy := getPolicy(given.caseName)

			xdsCtx := *xds_builders.Context().
				WithMeshBuilder(samples.MeshMTLSBuilder()).
				WithResources(resources).
				Build()
			proxy := xds_builders.Proxy().
				WithDataplane(samples.GatewayDataplaneBuilder()).
				WithSecretsTracker(secretsTracker).
				WithPolicies(xds_builders.MatchedPolicies().
					WithGatewayPolicy(api.MeshTLSType, getGatewayRules(policy.Spec.From)).
					WithGatewayPolicy(meshhttproute_api.MeshHTTPRouteType, core_rules.GatewayRules{
						ToRules: core_rules.GatewayToRules{
							ByListenerAndHostname: map[core_rules.InboundListenerHostname]core_rules.ToRules{
								core_rules.NewInboundListenerHostname("192.168.0.1", 8080, "*"): {
									Rules: core_rules.Rules{
										{
											BackendRefOriginIndex: backendRefOriginIndex,
											Origin: []core_model.ResourceMeta{
												&test_model.ResourceMeta{Mesh: "default", Name: "http-route"},
											},
											Subset: core_rules.MeshSubset(),
											Conf: meshhttproute_api.PolicyDefault{
												Rules: []meshhttproute_api.Rule{
													{
														Matches: []meshhttproute_api.Match{{
															Path: &meshhttproute_api.PathMatch{
																Type:  meshhttproute_api.Exact,
																Value: "/",
															},
														}},
														Default: meshhttproute_api.RuleConf{
															BackendRefs: &[]common_api.BackendRef{backendRef},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					})).
				Build()
			for n, p := range core_plugins.Plugins().ProxyPlugins() {
				Expect(p.Apply(context.Background(), xdsCtx.Mesh, proxy)).To(Succeed(), n)
			}
			gatewayGenerator := gateway_plugin.NewGenerator("test-zone")
			generatedResources, err := gatewayGenerator.Generate(context.Background(), nil, xdsCtx, proxy)
			Expect(err).NotTo(HaveOccurred())

			httpRoutePlugin := meshhttproute_plugin.NewPlugin().(core_plugins.PolicyPlugin)
			Expect(httpRoutePlugin.Apply(generatedResources, xdsCtx, proxy)).To(Succeed())

			// when
			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)
			Expect(plugin.Apply(generatedResources, xdsCtx, proxy)).To(Succeed())

			// then
			Expect(getResource(generatedResources, envoy_resource.ListenerType)).
				To(matchers.MatchGoldenYAML(fmt.Sprintf("testdata/%s.listeners.golden.yaml", given.caseName)))
			Expect(getResource(generatedResources, envoy_resource.ClusterType)).
				To(matchers.MatchGoldenYAML(fmt.Sprintf("testdata/%s.clusters.golden.yaml", given.caseName)))
			Expect(getResource(generatedResources, envoy_resource.RouteType)).
				To(matchers.MatchGoldenYAML(fmt.Sprintf("testdata/%s.routes.golden.yaml", given.caseName)))
		},
		Entry("tls version and cypher on gateway", testCase{
			caseName: "gateway-tls-version-and-cipher",
		}),
		Entry("tls version and cypher on gateway with MeshService", testCase{
			caseName:    "gateway-tls-version-and-cipher-meshservice",
			meshService: true,
		}),
	)
})

func getMeshServiceResources(secretsTracker core_xds.SecretsTracker, mesh *builders.MeshBuilder) []*core_xds.Resource {
	return []*core_xds.Resource{
		{
			Name:   "inbound:127.0.0.1:17777",
			Origin: generator.OriginInbound,
			Resource: listeners.NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17777, core_xds.SocketAddressProtocolTCP).
				Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
					Configure(listeners.HttpConnectionManager("127.0.0.1:17777", false)).
					Configure(
						listeners.HttpInboundRoutes(
							"backend",
							envoy_common.Routes{
								{
									Clusters: []envoy_common.Cluster{envoy_common.NewCluster(
										envoy_common.WithService("backend"),
										envoy_common.WithWeight(100),
									)},
								},
							},
						),
					),
				)).MustBuild(),
		},
		{
			Name:   "inbound:127.0.0.1:17778",
			Origin: generator.OriginInbound,
			Resource: listeners.NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17778, core_xds.SocketAddressProtocolTCP).
				Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
					Configure(listeners.TcpProxyDeprecated("127.0.0.1:17778", envoy_common.NewCluster(envoy_common.WithName("frontend")))),
				)).MustBuild(),
		},
		{
			Name:   "outbound",
			Origin: generator.OriginOutbound,
			Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "outgoing").
				Configure(clusters.ClientSideMTLS(secretsTracker, mesh.Build(), "outgoing", true, nil)).
				MustBuild(),
			Protocol: core_mesh.ProtocolHTTP,
			ResourceOrigin: &core_model.TypedResourceIdentifier{
				ResourceIdentifier: core_model.ResourceIdentifier{
					Name:      "backend",
					Mesh:      "default",
					Namespace: "backend-ns",
					Zone:      "zone-1",
				},
				ResourceType: "MeshService",
				SectionName:  "",
			},
		},
	}
}

func getResources(secretsTracker core_xds.SecretsTracker, mesh *builders.MeshBuilder) []*core_xds.Resource {
	return []*core_xds.Resource{
		{
			Name:   "inbound:127.0.0.1:17777",
			Origin: generator.OriginInbound,
			Resource: listeners.NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17777, core_xds.SocketAddressProtocolTCP).
				Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
					Configure(listeners.HttpConnectionManager("127.0.0.1:17777", false)).
					Configure(
						listeners.HttpInboundRoutes(
							"backend",
							envoy_common.Routes{
								{
									Clusters: []envoy_common.Cluster{envoy_common.NewCluster(
										envoy_common.WithService("backend"),
										envoy_common.WithWeight(100),
									)},
								},
							},
						),
					),
				)).MustBuild(),
		},
		{
			Name:   "inbound:127.0.0.1:17778",
			Origin: generator.OriginInbound,
			Resource: listeners.NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17778, core_xds.SocketAddressProtocolTCP).
				Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
					Configure(listeners.TcpProxyDeprecated("127.0.0.1:17778", envoy_common.NewCluster(envoy_common.WithName("frontend")))),
				)).MustBuild(),
		},
		{
			Name:   "outgoing",
			Origin: generator.OriginOutbound,
			Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "outgoing").
				Configure(clusters.ClientSideMTLS(secretsTracker, mesh.Build(), "outgoing", true, nil)).
				MustBuild(),
		},
	}
}

func getPolicy(caseName string) *api.MeshTLSResource {
	// setup
	meshTLS := api.NewMeshTLSResource()

	// when
	contents, err := os.ReadFile(path.Join("testdata", caseName+".policy.yaml"))
	Expect(err).ToNot(HaveOccurred())
	err = core_model.FromYAML(contents, &meshTLS.Spec)
	Expect(err).ToNot(HaveOccurred())

	meshTLS.SetMeta(&test_model.ResourceMeta{
		Name: "name",
		Mesh: core_model.DefaultMesh,
	})
	// and
	verr := meshTLS.Validate()
	Expect(verr).ToNot(HaveOccurred())

	return meshTLS
}

func getFromRules(froms []api.From) core_rules.FromRules {
	var rules []*core_rules.Rule

	for _, from := range froms {
		rules = append(rules, &core_rules.Rule{
			Subset: core_rules.Subset{},
			Conf:   from.Default,
		})
	}

	return core_rules.FromRules{
		Rules: map[core_rules.InboundListener]core_rules.Rules{
			{
				Address: "127.0.0.1",
				Port:    17777,
			}: rules,
			{
				Address: "127.0.0.1",
				Port:    17778,
			}: rules,
		},
	}
}

func getGatewayRules(froms []api.From) core_rules.GatewayRules {
	var rules []*core_rules.Rule

	for _, from := range froms {
		rules = append(rules, &core_rules.Rule{
			Subset: core_rules.Subset{},
			Conf:   from.Default,
		})
	}

	return core_rules.GatewayRules{
		FromRules: map[core_rules.InboundListener]core_rules.Rules{
			{
				Address: "127.0.0.1",
				Port:    17777,
			}: rules,
			{
				Address: "127.0.0.1",
				Port:    17778,
			}: rules,
		},
	}
}
