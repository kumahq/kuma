package v1alpha1

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"
	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/xds"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	gateway_plugin "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	clusters_builder "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	"github.com/kumahq/kuma/pkg/xds/envoy/endpoints/v3"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

type dummyCLACache struct {
	outboundTargets core_xds.EndpointMap
}

func (d *dummyCLACache) GetCLA(ctx context.Context, meshName, meshHash string, cluster envoy_common.Cluster, apiVersion core_xds.APIVersion, endpointMap core_xds.EndpointMap) (proto.Message, error) {
	return endpoints.CreateClusterLoadAssignment(cluster.Service(), d.outboundTargets[cluster.Service()]), nil
}

var _ envoy_common.CLACache = &dummyCLACache{}

var _ = Describe("MeshTimeout", func() {
	type sidecarTestCase struct {
		resources        []core_xds.Resource
		toRules          core_xds.ToRules
		fromRules        core_xds.FromRules
		expectedListener string
		expectedCluster  string
	}
	DescribeTable("should generate proper Envoy config",
		func(given sidecarTestCase) {
			resourceSet := core_xds.NewResourceSet()
			for _, res := range given.resources {
				r := res
				resourceSet.Add(&r)
			}

			context := createSimpleMeshContextWith(xds_context.NewResources())

			proxy := xds.Proxy{
				Dataplane: builders.Dataplane().
					WithName("backend").
					WithMesh("default").
					WithAddress("127.0.0.1").
					AddOutboundsToServices("other-service", "second-service").
					WithInboundOfTags(mesh_proto.ServiceTag, "backend", mesh_proto.ProtocolTag, "http").
					Build(),
				Policies: xds.MatchedPolicies{
					Dynamic: map[core_model.ResourceType]xds.TypedMatchingPolicies{
						api.MeshTimeoutType: {
							Type:      api.MeshTimeoutType,
							ToRules:   given.toRules,
							FromRules: given.fromRules,
						},
					},
				},
				Routing: core_xds.Routing{
					OutboundTargets: core_xds.EndpointMap{
						"other-service": []core_xds.Endpoint{{
							Tags: map[string]string{
								"kuma.io/protocol": "http",
							},
						}},
						"second-service": []core_xds.Endpoint{{
							Tags: map[string]string{
								"kuma.io/protocol": "tcp",
							},
						}},
					},
				},
			}
			plugin := NewPlugin().(core_plugins.PolicyPlugin)

			Expect(plugin.Apply(resourceSet, context, &proxy)).To(Succeed())
			Expect(getResourceYaml(resourceSet.ListOf(envoy_resource.ListenerType))).To(matchers.MatchGoldenYAML(filepath.Join("..", "testdata", given.expectedListener)))
			Expect(getResourceYaml(resourceSet.ListOf(envoy_resource.ClusterType))).To(matchers.MatchGoldenYAML(filepath.Join("..", "testdata", given.expectedCluster)))
		},
		Entry("basic outbound route", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:     "outbound",
				Origin:   generator.OriginOutbound,
				Resource: httpOutboundListenerWith(10001),
			},
				{
					Name:     "outbound",
					Origin:   generator.OriginOutbound,
					Resource: clusterWithName("other-service"),
				}},
			toRules: core_xds.ToRules{
				Rules: []*core_xds.Rule{
					{
						Subset: core_xds.Subset{},
						Conf: api.Conf{
							ConnectionTimeout: parseDuration("10s"),
							IdleTimeout:       parseDuration("1h"),
							Http: &api.Http{
								RequestTimeout:        parseDuration("5s"),
								StreamIdleTimeout:     parseDuration("1s"),
								MaxStreamDuration:     parseDuration("10m"),
								MaxConnectionDuration: parseDuration("10m"),
							},
						},
					},
				},
			},
			expectedListener: "basic_outbound_listener.golden.yaml",
			expectedCluster:  "basic_outbound_cluster.golden.yaml",
		}),
		Entry("tcp outbound route", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:   "outbound",
				Origin: generator.OriginOutbound,
				Resource: NewListenerBuilder(envoy_common.APIV3).
					Configure(OutboundListener("outbound:127.0.0.1:10002", "127.0.0.1", 10002, core_xds.SocketAddressProtocolTCP)).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
						Configure(TcpProxy(
							"127.0.0.1:10002",
							envoy_common.NewCluster(
								envoy_common.WithService("backend"),
								envoy_common.WithWeight(100),
							),
						)),
					)).
					MustBuild(),
			},
				{
					Name:     "outbound",
					Origin:   generator.OriginOutbound,
					Resource: clusterWithName("second-service"),
				}},
			toRules: core_xds.ToRules{
				Rules: []*core_xds.Rule{
					{
						Subset: core_xds.Subset{core_xds.Tag{
							Key:   mesh_proto.ServiceTag,
							Value: "second-service",
						}},
						Conf: api.Conf{
							ConnectionTimeout: parseDuration("10s"),
							IdleTimeout:       parseDuration("30s"),
						},
					},
				},
			},
			expectedCluster:  "basic_tcp_cluster.golden.yaml",
			expectedListener: "basic_tcp_listener.golden.yaml",
		}),
		Entry("basic inbound route", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:     "inbound",
				Origin:   generator.OriginInbound,
				Resource: httpInboundListenerWith(80),
			},
				{
					Name:     "inbound",
					Origin:   generator.OriginInbound,
					Resource: clusterWithName("localhost:80"),
				}},
			fromRules: core_xds.FromRules{
				Rules: map[core_xds.InboundListener]core_xds.Rules{
					{
						Address: "127.0.0.1",
						Port:    80,
					}: []*core_xds.Rule{
						{
							Subset: core_xds.Subset{},
							Conf: api.Conf{
								ConnectionTimeout: parseDuration("10s"),
								IdleTimeout:       parseDuration("1h"),
								Http: &api.Http{
									RequestTimeout:        parseDuration("5s"),
									StreamIdleTimeout:     parseDuration("1s"),
									MaxStreamDuration:     parseDuration("10m"),
									MaxConnectionDuration: parseDuration("10m"),
								},
							},
						},
					}},
			},
			expectedCluster:  "basic_inbound_cluster.golden.yaml",
			expectedListener: "basic_inbound_listener.golden.yaml",
		}),
		Entry("outbound with defaults when http conf missing", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:     "outbound",
				Origin:   generator.OriginOutbound,
				Resource: httpOutboundListenerWith(10001),
			},
				{
					Name:     "outbound",
					Origin:   generator.OriginOutbound,
					Resource: clusterWithName("other-service"),
				}},
			toRules: core_xds.ToRules{
				Rules: []*core_xds.Rule{
					{
						Subset: core_xds.Subset{
							{
								Key:   mesh_proto.ServiceTag,
								Value: "other-service",
							},
						},
						Conf: api.Conf{
							ConnectionTimeout: parseDuration("10s"),
							IdleTimeout:       parseDuration("1h"),
						},
					},
				},
			},
			expectedCluster:  "outbound_with_defaults_cluster.golden.yaml",
			expectedListener: "outbound_with_defaults_listener.golden.yaml",
		}),
	)

	It("should generate proper Envoy config for MeshGateway Dataplanes",
		func() {
			toRules := core_xds.ToRules{
				Rules: []*core_xds.Rule{
					{
						Subset: core_xds.Subset{},
						Conf: api.Conf{
							ConnectionTimeout: parseDuration("10s"),
							IdleTimeout:       parseDuration("1h"),
							Http: &api.Http{
								RequestTimeout:        parseDuration("5s"),
								StreamIdleTimeout:     parseDuration("1s"),
								MaxStreamDuration:     parseDuration("10m"),
								MaxConnectionDuration: parseDuration("10m"),
							},
						},
					},
				}}

			resources := xds_context.NewResources()
			resources.MeshLocalResources[core_mesh.MeshGatewayType] = GatewayResources()
			resources.MeshLocalResources[core_mesh.MeshGatewayRouteType] = GatewayRoutes()

			context := createSimpleMeshContextWith(resources)
			proxy := xds.Proxy{
				APIVersion: "v3",
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Mesh: "default",
						Name: "gateway",
					},
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Address: "127.0.0.1",
							Gateway: &mesh_proto.Dataplane_Networking_Gateway{
								Tags: map[string]string{
									mesh_proto.ServiceTag: "gateway",
								},
								Type: mesh_proto.Dataplane_Networking_Gateway_BUILTIN,
							},
						},
					},
				},
				Policies: xds.MatchedPolicies{
					Dynamic: map[core_model.ResourceType]xds.TypedMatchingPolicies{
						api.MeshTimeoutType: {
							Type:    api.MeshTimeoutType,
							ToRules: toRules,
						},
					},
				},
			}
			gatewayGenerator := gatewayGenerator()
			generatedResources, err := gatewayGenerator.Generate(context, &proxy)
			if err != nil {
				return
			}
			plugin := NewPlugin().(core_plugins.PolicyPlugin)

			Expect(plugin.Apply(generatedResources, context, &proxy)).To(Succeed())
			Expect(getResourceYaml(generatedResources.ListOf(envoy_resource.ListenerType))).To(matchers.MatchGoldenYAML(filepath.Join("..", "testdata", "gateway_listener.golden.yaml")))
			Expect(getResourceYaml(generatedResources.ListOf(envoy_resource.ClusterType))).To(matchers.MatchGoldenYAML(filepath.Join("..", "testdata", "gateway_cluster.golden.yaml")))
			Expect(getResourceYaml(generatedResources.ListOf(envoy_resource.RouteType))).To(matchers.MatchGoldenYAML(filepath.Join("..", "testdata", "gateway_route.golden.yaml")))
		})
})

func parseDuration(duration string) *k8s.Duration {
	d, _ := time.ParseDuration(duration)
	return &k8s.Duration{Duration: d}
}

func clusterWithName(name string) envoy_common.NamedResource {
	return clusters.NewClusterBuilder(envoy_common.APIV3).
		Configure(WithName(name)).
		MustBuild()
}

type NameConfigurer struct {
	Name string
}

func (n *NameConfigurer) Configure(c *envoy_cluster.Cluster) error {
	c.Name = n.Name
	return nil
}

func WithName(name string) clusters_builder.ClusterBuilderOpt {
	return clusters_builder.ClusterBuilderOptFunc(func(config *clusters_builder.ClusterBuilderConfig) {
		config.AddV3(&NameConfigurer{Name: name})
	})
}

func getResourceYaml(list core_xds.ResourceList) []byte {
	actualListener, err := util_proto.ToYAML(list[0].Resource)
	Expect(err).ToNot(HaveOccurred())
	return actualListener
}

func GatewayResources() *core_mesh.MeshGatewayResourceList {
	return &core_mesh.MeshGatewayResourceList{
		Items: []*core_mesh.MeshGatewayResource{{
			Meta: &test_model.ResourceMeta{Name: "gateway", Mesh: "default"},
			Spec: &mesh_proto.MeshGateway{
				Selectors: []*mesh_proto.Selector{{
					Match: map[string]string{
						mesh_proto.ServiceTag: "gateway",
					}},
				},
				Conf: &mesh_proto.MeshGateway_Conf{
					Listeners: []*mesh_proto.MeshGateway_Listener{
						{
							Protocol: mesh_proto.MeshGateway_Listener_HTTP,
							Port:     8080,
						},
					},
				},
			},
		}},
	}
}

func GatewayRoutes() *core_mesh.MeshGatewayRouteResourceList {
	return &core_mesh.MeshGatewayRouteResourceList{
		Items: []*core_mesh.MeshGatewayRouteResource{
			{
				Meta: &test_model.ResourceMeta{Name: "gateway", Mesh: "default"},
				Spec: &mesh_proto.MeshGatewayRoute{
					Selectors: []*mesh_proto.Selector{{
						Match: map[string]string{
							mesh_proto.ServiceTag: "gateway",
						}},
					},
					Conf: &mesh_proto.MeshGatewayRoute_Conf{
						Route: &mesh_proto.MeshGatewayRoute_Conf_Http{
							Http: &mesh_proto.MeshGatewayRoute_HttpRoute{
								Rules: []*mesh_proto.MeshGatewayRoute_HttpRoute_Rule{
									{
										Matches: []*mesh_proto.MeshGatewayRoute_HttpRoute_Match{
											{Path: &mesh_proto.MeshGatewayRoute_HttpRoute_Match_Path{
												Match: 0,
												Value: "/",
											}},
										},
										Backends: []*mesh_proto.MeshGatewayRoute_Backend{
											{
												Weight: 1,
												Destination: map[string]string{
													"kuma.io/service": "some-service",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func createSimpleMeshContextWith(resources xds_context.Resources) xds_context.Context {
	return xds_context.Context{
		Mesh: xds_context.MeshContext{
			Resource: &core_mesh.MeshResource{
				Meta: &test_model.ResourceMeta{
					Name: "default",
				},
			},
			Resources: resources,
			EndpointMap: map[core_xds.ServiceName][]core_xds.Endpoint{
				"some-service": {
					{
						Tags: map[string]string{
							"app": "some-service",
						},
					},
				},
			},
		},
		ControlPlane: &xds_context.ControlPlaneContext{CLACache: &dummyCLACache{}, Zone: "test-zone"},
	}
}

func gatewayGenerator() gateway_plugin.Generator {
	return gateway_plugin.Generator{
		FilterChainGenerators: gateway_plugin.FilterChainGenerators{
			FilterChainGenerators: map[mesh_proto.MeshGateway_Listener_Protocol]gateway_plugin.FilterChainGenerator{
				mesh_proto.MeshGateway_Listener_HTTP:  &gateway_plugin.HTTPFilterChainGenerator{},
				mesh_proto.MeshGateway_Listener_HTTPS: &gateway_plugin.HTTPSFilterChainGenerator{},
				mesh_proto.MeshGateway_Listener_TCP:   &gateway_plugin.TCPFilterChainGenerator{},
			}},
		ClusterGenerator: gateway_plugin.ClusterGenerator{
			Zone: "test-zone",
		},
		Zone: "test-zone",
	}
}

func httpOutboundListenerWith(port uint32) envoy_common.NamedResource {
	return createListener(
		port,
		OutboundListener(fmt.Sprintf("outbound:127.0.0.1:%d", port), "127.0.0.1", port, core_xds.SocketAddressProtocolTCP),
		HttpOutboundRoute(
			"backend",
			envoy_common.Routes{{
				Clusters: []envoy_common.Cluster{envoy_common.NewCluster(
					envoy_common.WithService("backend"),
					envoy_common.WithWeight(100),
				)},
			}},
			map[string]map[string]bool{
				"kuma.io/service": {
					"web": true,
				},
			},
		),
		"outbound",
	)
}

func httpInboundListenerWith(port uint32) envoy_common.NamedResource {
	return createListener(
		port,
		InboundListener(fmt.Sprintf("inbound:127.0.0.1:%d", port), "127.0.0.1", port, core_xds.SocketAddressProtocolTCP),
		HttpInboundRoutes(
			"backend",
			envoy_common.Routes{{
				Clusters: []envoy_common.Cluster{envoy_common.NewCluster(
					envoy_common.WithService("backend"),
					envoy_common.WithWeight(100),
				)},
			}},
		),
		"inbound")
}

func createListener(port uint32, listener ListenerBuilderOpt, route FilterChainBuilderOpt, direction string) envoy_common.NamedResource {
	return NewListenerBuilder(envoy_common.APIV3).
		Configure(listener).
		Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
			Configure(HttpConnectionManager(fmt.Sprintf("%s:127.0.0.1:%d", direction, port), false)).
			Configure(route),
		)).MustBuild()
}
