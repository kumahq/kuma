package api_server_test

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/emicklei/go-restful/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"google.golang.org/protobuf/types/known/durationpb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	api_server "github.com/kumahq/kuma/pkg/api-server"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/kds/samples"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	samples2 "github.com/kumahq/kuma/pkg/test/resources/samples"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

type selectors []*mesh_proto.Selector

func anyService() []*mesh_proto.Selector {
	return []*mesh_proto.Selector{
		{
			Match: map[string]string{
				mesh_proto.ServiceTag: "*",
			},
		},
	}
}

func serviceSelector(name, protocol string) *mesh_proto.Selector {
	if protocol == "" {
		return &mesh_proto.Selector{
			Match: map[string]string{
				mesh_proto.ServiceTag: name,
			},
		}
	} else {
		return &mesh_proto.Selector{
			Match: map[string]string{
				mesh_proto.ServiceTag:  name,
				mesh_proto.ProtocolTag: protocol,
			},
		}
	}
}

var _ = Describe("Inspect WS", func() {
	type testCase struct {
		path        string
		matcher     types.GomegaMatcher
		resources   []core_model.Resource
		global      bool
		contentType string
	}
	AfterEach(func() {
		core.Now = time.Now
	})

	DescribeTable("should return valid response",
		func(given testCase) {
			// setup
			core.Now = func() time.Time { return time.Time{} }

			resourceStore := memory.NewStore()
			rm := manager.NewResourceManager(resourceStore)
			for _, resource := range given.resources {
				err := rm.Create(context.Background(), resource,
					store.CreateBy(core_model.MetaToResourceKey(resource.GetMeta())))
				Expect(err).ToNot(HaveOccurred())
			}

			var apiServer *api_server.ApiServer
			var stop func()
			conf := NewTestApiServerConfigurer().WithStore(resourceStore)
			if given.global {
				conf = conf.WithGlobal()
			} else {
				conf = conf.WithZone("local")
			}
			apiServer, _, stop = StartApiServer(conf)
			defer stop()

			// when
			resp, err := http.Get((&url.URL{
				Scheme: "http",
				Host:   apiServer.Address(),
				Path:   given.path,
			}).String())
			Expect(err).ToNot(HaveOccurred())

			// then
			bytes, err := io.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(given.matcher)

			Expect(resp.Header.Get("content-type")).To(Equal(given.contentType))
		},
		Entry("inspect dataplane", testCase{
			path:    "/meshes/default/dataplanes/backend-1/policies",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_dataplane.json")),
			resources: []core_model.Resource{
				builders.Mesh().Build(),
				builders.Dataplane().
					WithName("backend-1").
					WithHttpServices("backend").
					AddOutboundsToServices("redis", "elastic", "postgres", "web").
					Build(),
				&core_mesh.TrafficPermissionResource{
					Meta: &test_model.ResourceMeta{Name: "tp-1", Mesh: "default"},
					Spec: &mesh_proto.TrafficPermission{
						Sources:      anyService(),
						Destinations: anyService(),
					},
				},
				&core_mesh.FaultInjectionResource{
					Meta: &test_model.ResourceMeta{Name: "fi-1", Mesh: "default"},
					Spec: &mesh_proto.FaultInjection{
						Sources: anyService(),
						Destinations: selectors{
							serviceSelector("backend", "http"),
						},
						Conf: &mesh_proto.FaultInjection_Conf{
							Delay: &mesh_proto.FaultInjection_Conf_Delay{
								Value:      durationpb.New(5 * time.Second),
								Percentage: util_proto.Double(90),
							},
						},
					},
				},
				&core_mesh.FaultInjectionResource{
					Meta: &test_model.ResourceMeta{Name: "fi-2", Mesh: "default"},
					Spec: &mesh_proto.FaultInjection{
						Sources: anyService(),
						Destinations: selectors{
							serviceSelector("backend", "http"),
						},
						Conf: &mesh_proto.FaultInjection_Conf{
							Abort: &mesh_proto.FaultInjection_Conf_Abort{
								HttpStatus: util_proto.UInt32(500),
								Percentage: util_proto.Double(80),
							},
						},
					},
				},
				&core_mesh.TimeoutResource{
					Meta: &test_model.ResourceMeta{Name: "t-1", Mesh: "default"},
					Spec: &mesh_proto.Timeout{
						Sources: anyService(),
						Destinations: selectors{
							serviceSelector("redis", ""),
						},
						Conf: samples.Timeout.Conf,
					},
				},
				&core_mesh.HealthCheckResource{
					Meta: &test_model.ResourceMeta{Name: "hc-1", Mesh: "default"},
					Spec: &mesh_proto.HealthCheck{
						Sources: selectors{
							serviceSelector("backend", ""),
						},
						Destinations: anyService(),
						Conf:         samples.HealthCheck.Conf,
					},
				},
				builders.MeshTrafficPermission().
					WithTargetRef(builders.TargetRefMesh()).
					AddFrom(builders.TargetRefMesh(), v1alpha1.Allow).
					Build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect dataplane, empty response", testCase{
			path:    "/meshes/default/dataplanes/backend-1/policies",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_dataplane_empty-response.json")),
			resources: []core_model.Resource{
				builders.Mesh().Build(),
				builders.Dataplane().
					WithName("backend-1").
					WithServices("backend").
					AddOutboundsToServices("redis", "elastic", "postgres", "web").
					Build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect gateway dataplane", testCase{
			path:    "/meshes/default/dataplanes/gateway-1/policies",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_gateway_dataplane.json")),
			resources: []core_model.Resource{
				builders.Mesh().Build(),
				builders.Dataplane().
					WithName("gateway-1").
					WithBuiltInGateway("elastic").
					Build(),
				&core_mesh.TrafficLogResource{
					Meta: &test_model.ResourceMeta{Name: "tl-1", Mesh: "default"},
					Spec: &mesh_proto.TrafficLog{
						Sources:      anyService(),
						Destinations: anyService(),
					},
				},
				&core_mesh.TrafficPermissionResource{
					Meta: &test_model.ResourceMeta{Name: "tp-1", Mesh: "default"},
					Spec: &mesh_proto.TrafficPermission{
						Sources:      anyService(),
						Destinations: anyService(),
					},
				},
				&core_mesh.MeshGatewayResource{
					Meta: &test_model.ResourceMeta{Name: "elastic", Mesh: "default"},
					Spec: &mesh_proto.MeshGateway{
						Selectors: selectors{
							serviceSelector("elastic", ""),
						},
						Conf: &mesh_proto.MeshGateway_Conf{
							Listeners: []*mesh_proto.MeshGateway_Listener{
								{
									Protocol: mesh_proto.MeshGateway_Listener_HTTP,
									Port:     80,
								},
							},
						},
					},
				},
				&core_mesh.MeshGatewayRouteResource{
					Meta: &test_model.ResourceMeta{Name: "route-1", Mesh: "default"},
					Spec: &mesh_proto.MeshGatewayRoute{
						Selectors: selectors{
							serviceSelector("elastic", ""),
						},
						Conf: &mesh_proto.MeshGatewayRoute_Conf{
							Route: &mesh_proto.MeshGatewayRoute_Conf_Http{
								Http: &mesh_proto.MeshGatewayRoute_HttpRoute{
									Rules: []*mesh_proto.MeshGatewayRoute_HttpRoute_Rule{
										{
											Matches: []*mesh_proto.MeshGatewayRoute_HttpRoute_Match{
												{
													Path: &mesh_proto.MeshGatewayRoute_HttpRoute_Match_Path{
														Match: mesh_proto.MeshGatewayRoute_HttpRoute_Match_Path_EXACT,
														Value: "/redis",
													},
												},
											},
											Backends: []*mesh_proto.MeshGatewayRoute_Backend{
												{
													Destination: serviceSelector("redis", "").Match,
												},
											},
										},
										{
											Matches: []*mesh_proto.MeshGatewayRoute_HttpRoute_Match{
												{
													Path: &mesh_proto.MeshGatewayRoute_HttpRoute_Match_Path{
														Match: mesh_proto.MeshGatewayRoute_HttpRoute_Match_Path_EXACT,
														Value: "/backend",
													},
												},
											},
											Backends: []*mesh_proto.MeshGatewayRoute_Backend{
												{
													Destination: serviceSelector("backend", "").Match,
												},
											},
										},
									},
								},
							},
						},
					},
				},
				&core_mesh.TimeoutResource{
					Meta: &test_model.ResourceMeta{Name: "t-1", Mesh: "default"},
					Spec: &mesh_proto.Timeout{
						Sources: anyService(),
						Destinations: selectors{
							serviceSelector("redis", ""),
						},
						Conf: samples.Timeout.Conf,
					},
				},
				&core_mesh.HealthCheckResource{
					Meta: &test_model.ResourceMeta{Name: "hc-1", Mesh: "default"},
					Spec: &mesh_proto.HealthCheck{
						Sources: selectors{
							serviceSelector("elastic", ""),
						},
						Destinations: selectors{
							serviceSelector("backend", ""),
						},
						Conf: samples.HealthCheck.Conf,
					},
				},
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect meshgateway dataplanes", testCase{
			path:    "/meshes/default/meshgateways/gateway/dataplanes",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_gateway_dataplanes.json")),
			resources: []core_model.Resource{
				builders.Mesh().Build(),
				builders.Dataplane().
					WithName("gateway-1").
					WithBuiltInGateway("gateway").
					Build(),
				builders.Dataplane().
					WithName("othergateway-1").
					WithBuiltInGateway("othergateway").
					Build(),
				builders.Dataplane().
					WithName("redis-1").
					WithServices("redis").
					AddOutboundsToServices("backend", "elastic").
					Build(),
				&core_mesh.MeshGatewayResource{
					Meta: &test_model.ResourceMeta{Name: "gateway", Mesh: "default"},
					Spec: &mesh_proto.MeshGateway{
						Selectors: selectors{
							serviceSelector("gateway", ""),
						},
						Conf: &mesh_proto.MeshGateway_Conf{
							Listeners: []*mesh_proto.MeshGateway_Listener{
								{
									Protocol: mesh_proto.MeshGateway_Listener_HTTP,
									Port:     80,
								},
							},
						},
					},
				},
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect meshgatewayroute dataplanes", testCase{
			path:    "/meshes/default/meshgatewayroutes/gatewayroute/dataplanes",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_gatewayroutes_dataplanes.json")),
			resources: []core_model.Resource{
				builders.Mesh().Build(),
				builders.Dataplane().
					WithName("gateway-1").
					WithBuiltInGateway("elastic").
					Build(),
				builders.Dataplane().
					WithName("othergateway-1").
					WithBuiltInGateway("othergateway").
					Build(),
				builders.Dataplane().
					WithName("redis-1").
					WithServices("redis").
					AddOutboundsToServices("backend", "elastic").
					Build(),
				&core_mesh.MeshGatewayResource{
					Meta: &test_model.ResourceMeta{Name: "elastic", Mesh: "default"},
					Spec: &mesh_proto.MeshGateway{
						Selectors: selectors{
							serviceSelector("elastic", ""),
						},
						Conf: &mesh_proto.MeshGateway_Conf{
							Listeners: []*mesh_proto.MeshGateway_Listener{
								{
									Protocol: mesh_proto.MeshGateway_Listener_HTTP,
									Port:     80,
								},
							},
						},
					},
				},
				&core_mesh.MeshGatewayRouteResource{
					Meta: &test_model.ResourceMeta{Name: "gatewayroute", Mesh: "default"},
					Spec: &mesh_proto.MeshGatewayRoute{
						Selectors: selectors{
							serviceSelector("elastic", ""),
						},
						Conf: &mesh_proto.MeshGatewayRoute_Conf{
							Route: &mesh_proto.MeshGatewayRoute_Conf_Http{
								Http: &mesh_proto.MeshGatewayRoute_HttpRoute{
									Rules: []*mesh_proto.MeshGatewayRoute_HttpRoute_Rule{
										{
											Matches: []*mesh_proto.MeshGatewayRoute_HttpRoute_Match{
												{
													Path: &mesh_proto.MeshGatewayRoute_HttpRoute_Match_Path{
														Match: mesh_proto.MeshGatewayRoute_HttpRoute_Match_Path_EXACT,
														Value: "/redis",
													},
												},
											},
											Backends: []*mesh_proto.MeshGatewayRoute_Backend{
												{
													Destination: serviceSelector("redis", "").Match,
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
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect traffic permission", testCase{
			path:    "/meshes/default/traffic-permissions/tp-1/dataplanes",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_traffic-permission.json")),
			resources: []core_model.Resource{
				builders.Mesh().Build(),
				&core_mesh.TrafficPermissionResource{
					Meta: &test_model.ResourceMeta{Name: "tp-1", Mesh: "default"},
					Spec: &mesh_proto.TrafficPermission{
						Sources: anyService(),
						Destinations: selectors{
							serviceSelector("backend", "http"),
							serviceSelector("redis", "http"),
							serviceSelector("elastic", "http"),
						},
					},
				},
				builders.Dataplane().
					WithName("backend-1").
					WithHttpServices("backend").
					AddOutboundsToServices("redis", "elastic", "postgres").
					Build(),
				builders.Dataplane().
					WithName("redis-1").
					WithHttpServices("redis").
					AddOutboundsToServices("redis", "elastic", "postgres").
					Build(),
				builders.Dataplane().
					WithName("elastic-1").
					WithHttpServices("elastic").
					AddOutboundsToServices("redis", "elastic", "postgres").
					Build(),
				builders.Dataplane().WithName("web-1").WithServices("web").Build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect fault injection", testCase{
			path:    "/meshes/mesh-1/fault-injections/fi-1/dataplanes",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_fault-injection.json")),
			resources: []core_model.Resource{
				builders.Mesh().WithName("mesh-1").Build(),
				&core_mesh.FaultInjectionResource{
					Meta: &test_model.ResourceMeta{Name: "fi-1", Mesh: "mesh-1"},
					Spec: &mesh_proto.FaultInjection{
						Sources: anyService(),
						Destinations: selectors{
							serviceSelector("backend", "http"),
							serviceSelector("redis", "http"),
							serviceSelector("elastic", "http"),
						},
						Conf: samples.FaultInjection.Conf,
					},
				},
				builders.Dataplane().WithName("backend-redis-1").WithMesh("mesh-1").WithHttpServices("backend", "redis").Build(),
				builders.Dataplane().WithName("elastic-1").WithMesh("mesh-1").WithHttpServices("elastic").AddOutboundsToServices("backend", "redis").Build(),
				// not matched by FaultInjection
				builders.Dataplane().WithName("web-1").WithMesh("mesh-1").WithHttpServices("web").Build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect rate limit", testCase{
			path:    "/meshes/mesh-1/rate-limits/rl-1/dataplanes",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_rate-limit.json")),
			resources: []core_model.Resource{
				builders.Mesh().WithName("mesh-1").Build(),
				&core_mesh.RateLimitResource{
					Meta: &test_model.ResourceMeta{Name: "rl-1", Mesh: "mesh-1"},
					Spec: &mesh_proto.RateLimit{
						Sources: anyService(),
						Destinations: selectors{
							serviceSelector("backend", "http"),
							serviceSelector("redis", "http"),
							serviceSelector("elastic", "http"),
							serviceSelector("es", ""),
						},
						Conf: samples.RateLimit.Conf,
					},
				},
				builders.Dataplane().WithName("elastic-1").WithMesh("mesh-1").WithHttpServices("elastic").AddOutboundsToServices("backend", "redis", "es").Build(),
				// not matched by RateLimit
				builders.Dataplane().WithName("web-1").WithMesh("mesh-1").WithHttpServices("web").Build(),
				&core_mesh.ExternalServiceResource{
					Meta: &test_model.ResourceMeta{Name: "es-1", Mesh: "mesh-1"},
					Spec: &mesh_proto.ExternalService{
						Networking: &mesh_proto.ExternalService_Networking{Address: "2.2.2.2:80"},
						Tags: map[string]string{
							mesh_proto.ServiceTag: "es",
						},
					},
				},
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect traffic log", testCase{
			path:    "/meshes/mesh-1/traffic-logs/tl-1/dataplanes",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_traffic-log.json")),
			resources: []core_model.Resource{
				builders.Mesh().WithName("mesh-1").Build(),
				&core_mesh.TrafficLogResource{
					Meta: &test_model.ResourceMeta{Name: "tl-1", Mesh: "mesh-1"},
					Spec: &mesh_proto.TrafficLog{
						Sources: anyService(),
						Destinations: selectors{
							serviceSelector("redis", ""),
							serviceSelector("elastic", ""),
						},
						Conf: samples.TrafficLog.Conf,
					},
				},
				builders.Dataplane().WithName("backend-1").WithMesh("mesh-1").WithServices("backend").AddOutboundsToServices("redis", "elastic", "web").Build(),
				builders.Dataplane().WithName("redis-1").WithMesh("mesh-1").WithServices("redis").AddOutboundsToServices("elastic", "web").Build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect health check", testCase{
			path:    "/meshes/mesh-1/health-checks/hc-1/dataplanes",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_health-check.json")),
			resources: []core_model.Resource{
				builders.Mesh().WithName("mesh-1").Build(),
				&core_mesh.HealthCheckResource{
					Meta: &test_model.ResourceMeta{Name: "hc-1", Mesh: "mesh-1"},
					Spec: &mesh_proto.HealthCheck{
						Sources: anyService(),
						Destinations: selectors{
							serviceSelector("redis", ""),
							serviceSelector("elastic", ""),
						},
						Conf: samples.HealthCheck.Conf,
					},
				},
				builders.Dataplane().WithName("backend-1").WithMesh("mesh-1").WithServices("backend").AddOutboundsToServices("redis", "elastic", "web").Build(),
				builders.Dataplane().WithName("redis-1").WithMesh("mesh-1").WithServices("redis").AddOutboundsToServices("elastic", "web").Build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect circuit breaker", testCase{
			path:    "/meshes/mesh-1/circuit-breakers/cb-1/dataplanes",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_circuit-breaker.json")),
			resources: []core_model.Resource{
				builders.Mesh().WithName("mesh-1").Build(),
				&core_mesh.CircuitBreakerResource{
					Meta: &test_model.ResourceMeta{Name: "cb-1", Mesh: "mesh-1"},
					Spec: &mesh_proto.CircuitBreaker{
						Sources: anyService(),
						Destinations: selectors{
							serviceSelector("redis", ""),
							serviceSelector("elastic", ""),
						},
						Conf: samples.CircuitBreaker.Conf,
					},
				},
				builders.Dataplane().WithName("backend-1").WithMesh("mesh-1").WithServices("backend").AddOutboundsToServices("redis", "elastic", "web").Build(),
				builders.Dataplane().WithName("redis-1").WithMesh("mesh-1").WithServices("redis").AddOutboundsToServices("elastic", "web").Build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect retry", testCase{
			path:    "/meshes/mesh-1/retries/r-1/dataplanes",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_retry.json")),
			resources: []core_model.Resource{
				builders.Mesh().WithName("mesh-1").Build(),
				&core_mesh.MeshGatewayResource{
					Meta: &test_model.ResourceMeta{Name: "gateway", Mesh: "mesh-1"},
					Spec: &mesh_proto.MeshGateway{
						Selectors: selectors{
							serviceSelector("meshgateway", ""),
						},
						Conf: &mesh_proto.MeshGateway_Conf{
							Listeners: []*mesh_proto.MeshGateway_Listener{
								{
									Protocol: mesh_proto.MeshGateway_Listener_HTTP,
									Port:     80,
								},
							},
						},
					},
				},
				&core_mesh.MeshGatewayRouteResource{
					Meta: &test_model.ResourceMeta{Name: "route-1", Mesh: "mesh-1"},
					Spec: &mesh_proto.MeshGatewayRoute{
						Selectors: selectors{
							serviceSelector("meshgateway", ""),
						},
						Conf: &mesh_proto.MeshGatewayRoute_Conf{
							Route: &mesh_proto.MeshGatewayRoute_Conf_Http{
								Http: &mesh_proto.MeshGatewayRoute_HttpRoute{
									Rules: []*mesh_proto.MeshGatewayRoute_HttpRoute_Rule{
										{
											Matches: []*mesh_proto.MeshGatewayRoute_HttpRoute_Match{
												{
													Path: &mesh_proto.MeshGatewayRoute_HttpRoute_Match_Path{
														Match: mesh_proto.MeshGatewayRoute_HttpRoute_Match_Path_EXACT,
														Value: "/redis",
													},
												},
											},
											Backends: []*mesh_proto.MeshGatewayRoute_Backend{
												{
													Destination: serviceSelector("redis", "").Match,
												},
											},
										},
										{
											Matches: []*mesh_proto.MeshGatewayRoute_HttpRoute_Match{
												{
													Path: &mesh_proto.MeshGatewayRoute_HttpRoute_Match_Path{
														Match: mesh_proto.MeshGatewayRoute_HttpRoute_Match_Path_EXACT,
														Value: "/backend",
													},
												},
											},
											Backends: []*mesh_proto.MeshGatewayRoute_Backend{
												{
													Destination: serviceSelector("backend", "").Match,
												},
											},
										},
									},
								},
							},
						},
					},
				},
				&core_mesh.RetryResource{
					Meta: &test_model.ResourceMeta{Name: "r-1", Mesh: "mesh-1"},
					Spec: &mesh_proto.Retry{
						Sources: anyService(),
						Destinations: selectors{
							serviceSelector("redis", ""),
							serviceSelector("elastic", ""),
						},
						Conf: samples.Retry.Conf,
					},
				},
				builders.Dataplane().WithName("meshgateway-1").WithMesh("mesh-1").WithBuiltInGateway("meshgateway").Build(),
				builders.Dataplane().WithName("backend-1").WithMesh("mesh-1").WithServices("backend").AddOutboundsToServices("redis", "elastic", "web").Build(),
				builders.Dataplane().WithName("redis-1").WithMesh("mesh-1").WithServices("redis").AddOutboundsToServices("elastic", "web").Build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect timeout", testCase{
			path:    "/meshes/mesh-1/timeouts/t-1/dataplanes",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_timeout.json")),
			resources: []core_model.Resource{
				builders.Mesh().WithName("mesh-1").Build(),
				&core_mesh.TimeoutResource{
					Meta: &test_model.ResourceMeta{Name: "t-1", Mesh: "mesh-1"},
					Spec: &mesh_proto.Timeout{
						Sources: anyService(),
						Destinations: selectors{
							serviceSelector("redis", ""),
							serviceSelector("elastic", ""),
						},
						Conf: samples.Timeout.Conf,
					},
				},
				builders.Dataplane().WithName("backend-1").WithMesh("mesh-1").WithServices("backend").AddOutboundsToServices("redis", "elastic", "web").Build(),
				builders.Dataplane().WithName("redis-1").WithMesh("mesh-1").WithServices("redis").AddOutboundsToServices("elastic", "web").Build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect traffic route", testCase{
			path:    "/meshes/mesh-1/traffic-routes/t-1/dataplanes",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_traffic-route.json")),
			resources: []core_model.Resource{
				builders.Mesh().WithName("mesh-1").Build(),
				&core_mesh.TrafficRouteResource{
					Meta: &test_model.ResourceMeta{Name: "t-1", Mesh: "mesh-1"},
					Spec: &mesh_proto.TrafficRoute{
						Sources:      anyService(),
						Destinations: anyService(),
						Conf:         samples.TrafficRoute.Conf,
					},
				},
				builders.Dataplane().WithName("backend-1").WithMesh("mesh-1").WithServices("backend").AddOutboundsToServices("redis", "web").Build(),
				builders.Dataplane().WithName("redis-1").WithMesh("mesh-1").WithServices("redis").AddOutboundsToServices("elastic", "web").Build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect traffic trace", testCase{
			path:    "/meshes/mesh-1/traffic-traces/tt-1/dataplanes",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_traffic-trace.json")),
			resources: []core_model.Resource{
				builders.Mesh().WithName("mesh-1").Build(),
				&core_mesh.TrafficTraceResource{
					Meta: &test_model.ResourceMeta{Name: "tt-1", Mesh: "mesh-1"},
					Spec: &mesh_proto.TrafficTrace{
						Selectors: anyService(),
						Conf:      samples.TrafficTrace.Conf,
					},
				},
				&core_mesh.MeshGatewayResource{
					Meta: &test_model.ResourceMeta{Name: "meshgateway", Mesh: "mesh-1"},
					Spec: &mesh_proto.MeshGateway{
						Selectors: selectors{
							serviceSelector("meshgateway", ""),
						},
						Conf: &mesh_proto.MeshGateway_Conf{
							Listeners: []*mesh_proto.MeshGateway_Listener{
								{
									Protocol: mesh_proto.MeshGateway_Listener_HTTP,
									Port:     80,
								},
							},
						},
					},
				},
				&core_mesh.MeshGatewayRouteResource{
					Meta: &test_model.ResourceMeta{Name: "route-1", Mesh: "mesh-1"},
					Spec: &mesh_proto.MeshGatewayRoute{
						Selectors: selectors{
							serviceSelector("meshgateway", ""),
						},
						Conf: &mesh_proto.MeshGatewayRoute_Conf{
							Route: &mesh_proto.MeshGatewayRoute_Conf_Http{
								Http: &mesh_proto.MeshGatewayRoute_HttpRoute{
									Rules: []*mesh_proto.MeshGatewayRoute_HttpRoute_Rule{
										{
											Matches: []*mesh_proto.MeshGatewayRoute_HttpRoute_Match{
												{
													Path: &mesh_proto.MeshGatewayRoute_HttpRoute_Match_Path{
														Match: mesh_proto.MeshGatewayRoute_HttpRoute_Match_Path_EXACT,
														Value: "/redis",
													},
												},
											},
											Backends: []*mesh_proto.MeshGatewayRoute_Backend{
												{
													Destination: serviceSelector("redis", "").Match,
												},
											},
										},
									},
								},
							},
						},
					},
				},
				builders.Dataplane().WithName("meshgateway-1").WithMesh("mesh-1").WithBuiltInGateway("meshgateway").Build(),
				builders.Dataplane().WithName("backend-1").WithMesh("mesh-1").WithServices("backend").AddOutboundsToServices("redis", "elastic", "web").Build(),
				builders.Dataplane().WithName("redis-1").WithMesh("mesh-1").WithServices("redis").AddOutboundsToServices("redis", "elastic", "web").Build(),
				builders.Dataplane().WithName("web-1").WithMesh("mesh-1").WithServices("web").AddOutboundsToServices("elastic", "backend").Build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect proxytemplate", testCase{
			path:    "/meshes/mesh-1/proxytemplates/tt-1/dataplanes",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_proxytemplate.json")),
			resources: []core_model.Resource{
				builders.Mesh().WithName("mesh-1").Build(),
				&core_mesh.ProxyTemplateResource{
					Meta: &test_model.ResourceMeta{Name: "tt-1", Mesh: "mesh-1"},
					Spec: &mesh_proto.ProxyTemplate{
						Selectors: anyService(),
						Conf:      samples.ProxyTemplate.Conf,
					},
				},
				builders.Dataplane().WithName("backend-1").WithMesh("mesh-1").WithServices("backend").AddOutboundsToServices("redis", "elastic", "web").Build(),
				builders.Dataplane().WithName("redis-1").WithMesh("mesh-1").WithServices("redis").AddOutboundsToServices("redis", "elastic", "web").Build(),
				builders.Dataplane().WithName("web-1").WithMesh("mesh-1").WithServices("web").AddOutboundsToServices("elastic", "backend").Build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect traffic trace, empty response", testCase{
			path:    "/meshes/mesh-1/traffic-traces/tt-1/dataplanes",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_traffic-trace_empty-response.json")),
			resources: []core_model.Resource{
				builders.Mesh().WithName("mesh-1").Build(),
				&core_mesh.TrafficTraceResource{
					Meta: &test_model.ResourceMeta{Name: "tt-1", Mesh: "mesh-1"},
					Spec: &mesh_proto.TrafficTrace{
						Selectors: anyService(),
						Conf:      samples.TrafficTrace.Conf,
					},
				},
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect meshtrafficpermission", testCase{
			path:    "/meshes/mesh-1/meshtrafficpermissions/mtp-1/dataplanes",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_meshtrafficpermission.json")),
			resources: []core_model.Resource{
				builders.Mesh().WithName("mesh-1").Build(),
				builders.Dataplane().WithName("backend-1").WithMesh("mesh-1").WithServices("backend").Build(),
				builders.MeshTrafficPermission().
					WithMesh("mesh-1").
					WithTargetRef(builders.TargetRefService("backend")).
					AddFrom(builders.TargetRefMesh(), v1alpha1.Allow).
					Build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect xds for dataplane", testCase{
			path:    "/meshes/mesh-1/dataplanes/backend-1/xds",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_xds_dataplane.json")),
			resources: []core_model.Resource{
				builders.Mesh().WithName("mesh-1").Build(),
				builders.Dataplane().WithName("backend-1").WithAddress("1.1.1.1").WithMesh("mesh-1").WithAdminPort(3301).WithServices("backend").AddOutboundsToServices("redis", "elastic", "web").Build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect xds for local zone ingress", testCase{
			path:    "/zoneingresses/zi-1/xds",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_xds_local_zoneingress.json")),
			resources: []core_model.Resource{
				builders.ZoneIngress().
					WithName("zi-1").
					WithZone("").
					WithAdminPort(2201).
					WithAddress("2.2.2.2").
					WithPort(8080).
					WithAdvertisedAddress("3.3.3.3").
					WithAdvertisedPort(80).
					Build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect xds for zone ingress from another zone", testCase{
			path:    "/zoneingresses/zi-1/xds",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_xds_remote_zoneingress.json")),
			resources: []core_model.Resource{
				builders.ZoneIngress().
					WithName("zi-1").
					WithZone("not-local-zone").
					WithAdminPort(2201).
					WithAddress("2.2.2.2").
					WithPort(8080).
					WithAdvertisedAddress("3.3.3.3").
					WithAdvertisedPort(80).
					Build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect xds for zone ingress on global", testCase{
			global:  true,
			path:    "/zoneingresses/zi-1/xds",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_xds_local_zoneingress.json")),
			resources: []core_model.Resource{
				builders.ZoneIngress().
					WithName("zi-1").
					WithZone(""). // local zone ingress has empty "zone" field
					WithAdminPort(2201).
					WithAddress("2.2.2.2").
					WithPort(8080).
					WithAdvertisedAddress("3.3.3.3").
					WithAdvertisedPort(80).
					Build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect xds for dataplane on global", testCase{
			global:  true,
			path:    "/meshes/mesh-1/dataplanes/backend-1/xds",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_xds_dataplane.json")),
			resources: []core_model.Resource{
				builders.Mesh().WithName("mesh-1").Build(),
				builders.Dataplane().WithName("backend-1").WithMesh("mesh-1").WithAdminPort(3301).WithAddress("1.1.1.1").WithServices("backend").Build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect xds for zone egress", testCase{
			path:    "/zoneegresses/ze-1/xds",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_xds_zoneegress.json")),
			resources: []core_model.Resource{
				builders.ZoneEgress().WithName("ze-1").WithAddress("4.4.4.4").WithPort(8080).WithAdminPort(4321).Build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect stats for dataplane", testCase{
			path:    "/meshes/mesh-1/dataplanes/backend-1/stats",
			matcher: matchers.MatchGoldenEqual(path.Join("testdata", "inspect_stats_dataplane.out")),
			resources: []core_model.Resource{
				builders.Mesh().WithName("mesh-1").Build(),
				builders.Dataplane().WithName("backend-1").WithMesh("mesh-1").WithAdminPort(3301).WithServices("backend").AddOutboundsToServices("redis", "elastic", "web").Build(),
			},
			contentType: "text/plain",
		}),
		Entry("inspect clusters for dataplane", testCase{
			path:    "/meshes/mesh-1/dataplanes/backend-1/clusters",
			matcher: matchers.MatchGoldenEqual(path.Join("testdata", "inspect_clusters_dataplane.out")),
			resources: []core_model.Resource{
				builders.Mesh().WithName("mesh-1").Build(),
				builders.Dataplane().WithName("backend-1").WithMesh("mesh-1").WithAdminPort(3301).WithServices("backend").AddOutboundsToServices("redis", "elastic", "web").Build(),
			},
			contentType: "text/plain",
		}),
		Entry("inspect rules empty", testCase{
			path:    "/meshes/default/dataplanes/web-01/rules",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_dataplane_rules_empty.golden.json")),
			resources: []core_model.Resource{
				samples2.MeshDefault(),
				samples2.DataplaneWeb(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect rules basic", testCase{
			path:    "/meshes/default/dataplanes/web-01/rules",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_dataplane_rules.golden.json")),
			resources: []core_model.Resource{
				samples2.MeshDefault(),
				samples2.DataplaneWeb(),
				builders.MeshTrafficPermission().
					WithTargetRef(builders.TargetRefService("web")).
					AddFrom(builders.TargetRefServiceSubset("client", "kuma.io/zone", "east"), v1alpha1.Deny).
					Build(),
				builders.MeshAccessLog().
					WithTargetRef(builders.TargetRefService("web")).
					AddTo(builders.TargetRefMesh(), samples2.MeshAccessLogFileConf()).
					Build(),
				builders.MeshTrace().
					WithTargetRef(builders.TargetRefService("web")).
					WithZipkinBackend(samples2.ZipkinBackend()).
					Build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect rules subset", testCase{
			path:    "/meshes/default/dataplanes/dp-1/rules",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_dataplane_rules_subset.golden.json")),
			resources: []core_model.Resource{
				builders.Mesh().Build(),
				builders.Dataplane().
					WithName("dp-1").
					WithHttpServices("web").
					AddOutbound(builders.Outbound().WithService("backend").WithAddress("240.0.0.1").WithPort(2300).WithTags(map[string]string{"version": "2"})).
					AddOutbound(builders.Outbound().WithService("backend").WithAddress("240.0.0.2").WithPort(2300).WithTags(map[string]string{"version": "1"})).
					AddOutbound(builders.Outbound().WithService("backend").WithAddress("10.3.2.3").WithPort(2300)).
					AddOutbound(builders.Outbound().WithService("backend").WithAddress("240.0.0.0").WithPort(80)).
					Build(),
				builders.MeshTrafficPermission().
					WithName("mtp-1").
					WithTargetRef(builders.TargetRefService("web")).
					AddFrom(builders.TargetRefServiceSubset("client", "kuma.io/zone", "east", "version", "2"), v1alpha1.Allow).
					Build(),
				builders.MeshTrafficPermission().
					WithName("mtp-2").
					WithTargetRef(builders.TargetRefService("web")).
					AddFrom(builders.TargetRefServiceSubset("client", "kuma.io/zone", "east"), v1alpha1.Deny).
					Build(),
				builders.MeshAccessLog().
					WithTargetRef(builders.TargetRefService("web")).
					AddTo(builders.TargetRefMesh(), samples2.MeshAccessLogFileConf()).
					Build(),
				builders.MeshTrace().
					WithTargetRef(builders.TargetRefService("web")).
					WithZipkinBackend(samples2.ZipkinBackend()).
					Build(),
			},
			contentType: restful.MIME_JSON,
		}),
	)

	It("should change response if state changed", func() {
		// setup
		var apiServer *api_server.ApiServer
		var stop func()
		resourceStore := memory.NewStore()
		rm := manager.NewResourceManager(resourceStore)
		apiServer, _, stop = StartApiServer(NewTestApiServerConfigurer().WithStore(resourceStore))
		defer stop()

		// when init the state
		// TrafficPermission that selects 2 DPPs
		initState := []core_model.Resource{
			builders.Mesh().Build(),
			&core_mesh.TrafficPermissionResource{
				Meta: &test_model.ResourceMeta{Name: "tp-1", Mesh: "default"},
				Spec: &mesh_proto.TrafficPermission{
					Sources: anyService(),
					Destinations: selectors{
						serviceSelector("backend", "http"),
						serviceSelector("redis", "http"),
					},
				},
			},
			builders.Dataplane().WithName("backend-1").WithHttpServices("backend").AddOutboundsToServices("redis", "elastic").Build(),
			builders.Dataplane().WithName("redis-1").WithHttpServices("redis").AddOutboundsToServices("redis", "backend", "elastic").Build(),
		}
		for _, resource := range initState {
			err := rm.Create(context.Background(), resource,
				store.CreateBy(core_model.MetaToResourceKey(resource.GetMeta())))
			Expect(err).ToNot(HaveOccurred())
		}

		// then
		var resp *http.Response
		Eventually(func() error {
			r, err := http.Get((&url.URL{
				Scheme: "http",
				Host:   apiServer.Address(),
				Path:   "/meshes/default/traffic-permissions/tp-1/dataplanes",
			}).String())
			resp = r
			return err
		}, "3s").ShouldNot(HaveOccurred())
		bytes, err := io.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())
		Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", "inspect_changed_state_before.json")))

		// when change the state
		err = rm.Delete(context.Background(), core_mesh.NewDataplaneResource(), store.DeleteByKey("backend-1", "default"))
		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func() error {
			r, err := http.Get((&url.URL{
				Scheme: "http",
				Host:   apiServer.Address(),
				Path:   "/meshes/default/traffic-permissions/tp-1/dataplanes",
			}).String())
			resp = r
			return err
		}, "3s").ShouldNot(HaveOccurred())
		bytes, err = io.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())
		Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", "inspect_changed_state_after.json")))
	})
})
