package v1alpha1

import (
	"time"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/xds"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/xds"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	clusters_builder "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ = Describe("MeshTimeout", func() {
	type sidecarTestCase struct {
		resources         []core_xds.Resource
		toRules           core_xds.ToRules
		fromRules         core_xds.FromRules
		expectedListeners []string
		expectedClusters  []string
	}
	DescribeTable("should generate proper Envoy config",
		func(given sidecarTestCase) {
			resourceSet := core_xds.NewResourceSet()
			for _, res := range given.resources {
				r := res
				resourceSet.Add(&r)
			}

			context := xds_context.Context{
				Mesh: xds_context.MeshContext{
					Resource: &core_mesh.MeshResource{
						Meta: &test_model.ResourceMeta{
							Name: "default",
						},
					},
				},
			}

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
			policies_xds.ResourceArrayShouldEqual(resourceSet.ListOf(envoy_resource.ListenerType), given.expectedListeners)
			policies_xds.ResourceArrayShouldEqual(resourceSet.ListOf(envoy_resource.ClusterType), given.expectedClusters)
		},
		Entry("basic outbound route", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:   "outbound",
				Origin: generator.OriginOutbound,
				Resource: NewListenerBuilder(envoy_common.APIV3).
					Configure(OutboundListener("outbound:127.0.0.1:10001", "127.0.0.1", 10001, core_xds.SocketAddressProtocolTCP)).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
						Configure(HttpConnectionManager("127.0.0.1:10001", false)).
						Configure(
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
						),
					)).MustBuild(),
			},
				{
					Name:     "outbound",
					Origin:   generator.OriginOutbound,
					Resource: buildClusterWithName("other-service"),
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
			expectedClusters: []string{`
        connectTimeout: 10s
        name: other-service
        typedExtensionProtocolOptions:
           envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
               '@type': type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
               commonHttpProtocolOptions:
                   idleTimeout: 3600s
                   maxConnectionDuration: 600s
                   maxStreamDuration: 600s`},
			expectedListeners: []string{`
address:
  socketAddress:
      address: 127.0.0.1
      portValue: 10001
filterChains:
  - filters:
      - name: envoy.filters.network.http_connection_manager
        typedConfig:
          '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
          httpFilters:
              - name: envoy.filters.http.router
                typedConfig:
                  '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
          routeConfig:
              name: outbound:backend
              requestHeadersToAdd:
                  - header:
                      key: x-kuma-tags
                      value: '&kuma.io/service=web&'
              validateClusters: false
              virtualHosts:
                  - domains:
                      - '*'
                    name: backend
                    routes:
                      - match:
                          prefix: /
                        route:
                          cluster: backend
                          timeout: 5s
          statPrefix: "127_0_0_1_10001"
          streamIdleTimeout: 1s
name: outbound:127.0.0.1:10001
trafficDirection: OUTBOUND`,
			}}),
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
					Resource: buildClusterWithName("second-service"),
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
			expectedClusters: []string{`
connectTimeout: 10s
name: second-service`},
			expectedListeners: []string{`
address:
  socketAddress:
      address: 127.0.0.1
      portValue: 10002
filterChains:
  - filters:
      - name: envoy.filters.network.tcp_proxy
        typedConfig:
          '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
          cluster: backend
          idleTimeout: 30s
          statPrefix: "127_0_0_1_10002"
name: outbound:127.0.0.1:10002
trafficDirection: OUTBOUND`,
			}}),
		Entry("basic inbound route", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:   "inbound",
				Origin: generator.OriginInbound,
				Resource: NewListenerBuilder(envoy_common.APIV3).
					Configure(InboundListener("inbound:127.0.0.1:80", "127.0.0.1", 80, core_xds.SocketAddressProtocolTCP)).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
						Configure(HttpConnectionManager("127.0.0.1:80", false)).
						Configure(
							HttpInboundRoutes(
								"backend",
								envoy_common.Routes{{
									Clusters: []envoy_common.Cluster{envoy_common.NewCluster(
										envoy_common.WithService("backend"),
										envoy_common.WithWeight(100),
									)},
								}},
							),
						),
					)).MustBuild(),
			},
				{
					Name:     "inbound",
					Origin:   generator.OriginInbound,
					Resource: buildClusterWithName("localhost:80"),
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
			expectedClusters: []string{`
        connectTimeout: 10s
        name: localhost:80
        typedExtensionProtocolOptions:
          envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
              '@type': type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
              commonHttpProtocolOptions:
                  idleTimeout: 3600s
                  maxConnectionDuration: 600s
                  maxStreamDuration: 600s`},
			expectedListeners: []string{`
        address:
          socketAddress:
              address: 127.0.0.1
              portValue: 80
        enableReusePort: false
        filterChains:
          - filters:
              - name: envoy.filters.network.http_connection_manager
                typedConfig:
                  '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                  httpFilters:
                      - name: envoy.filters.http.router
                        typedConfig:
                          '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                  routeConfig:
                      name: inbound:backend
                      requestHeadersToRemove:
                          - x-kuma-tags
                      validateClusters: false
                      virtualHosts:
                          - domains:
                              - '*'
                            name: backend
                            routes:
                              - match:
                                  prefix: /
                                route:
                                  cluster: backend
                                  timeout: 5s
                  statPrefix: "127_0_0_1_80"
                  streamIdleTimeout: 1s
        name: inbound:127.0.0.1:80
        trafficDirection: INBOUND`,
			}}),
		Entry("outbound with defaults when http conf missing", sidecarTestCase{
			resources: []core_xds.Resource{{
				Name:   "outbound",
				Origin: generator.OriginOutbound,
				Resource: NewListenerBuilder(envoy_common.APIV3).
					Configure(OutboundListener("outbound:127.0.0.1:10001", "127.0.0.1", 10001, core_xds.SocketAddressProtocolTCP)).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
						Configure(HttpConnectionManager("127.0.0.1:10001", false)).
						Configure(
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
						),
					)).MustBuild(),
			},
				{
					Name:     "outbound",
					Origin:   generator.OriginOutbound,
					Resource: buildClusterWithName("other-service"),
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
			expectedClusters: []string{`
        connectTimeout: 10s
        name: other-service
        typedExtensionProtocolOptions:
           envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
               '@type': type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
               commonHttpProtocolOptions:
                   idleTimeout: 3600s
                   maxConnectionDuration: 0s
                   maxStreamDuration: 0s`},
			expectedListeners: []string{`
address:
  socketAddress:
      address: 127.0.0.1
      portValue: 10001
filterChains:
  - filters:
      - name: envoy.filters.network.http_connection_manager
        typedConfig:
          '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
          httpFilters:
              - name: envoy.filters.http.router
                typedConfig:
                  '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
          routeConfig:
              name: outbound:backend
              requestHeadersToAdd:
                  - header:
                      key: x-kuma-tags
                      value: '&kuma.io/service=web&'
              validateClusters: false
              virtualHosts:
                  - domains:
                      - '*'
                    name: backend
                    routes:
                      - match:
                          prefix: /
                        route:
                          cluster: backend
                          timeout: 15s
          statPrefix: "127_0_0_1_10001"
          streamIdleTimeout: 1800s
name: outbound:127.0.0.1:10001
trafficDirection: OUTBOUND`,
			}}),
	)
})

func parseDuration(duration string) *k8s.Duration {
	d, _ := time.ParseDuration(duration)
	return &k8s.Duration{Duration: d}
}

func buildClusterWithName(name string) envoy_common.NamedResource {
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
