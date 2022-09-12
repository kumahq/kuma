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
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/kds/samples"
	"github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

type dataplaneBuilder core_mesh.DataplaneResource
type zoneIngressBuilder core_mesh.ZoneIngressResource
type zoneEgressBuilder core_mesh.ZoneEgressResource

func newMesh(name string) *core_mesh.MeshResource {
	return &core_mesh.MeshResource{
		Meta: &test_model.ResourceMeta{Name: name},
		Spec: &mesh_proto.Mesh{},
	}
}

func newZoneEgress() *zoneEgressBuilder {
	return &zoneEgressBuilder{
		Spec: &mesh_proto.ZoneEgress{
			Networking: &mesh_proto.ZoneEgress_Networking{},
		},
	}
}

func (b *zoneEgressBuilder) meta(name string) *zoneEgressBuilder {
	b.Meta = &test_model.ResourceMeta{Name: name, Mesh: core_model.NoMesh}
	return b
}

func (b *zoneEgressBuilder) address(address string) *zoneEgressBuilder {
	b.Spec.Networking.Address = address
	return b
}

func (b *zoneEgressBuilder) port(port uint32) *zoneEgressBuilder {
	b.Spec.Networking.Port = port
	return b
}

func (b *zoneEgressBuilder) admin(port uint32) *zoneEgressBuilder {
	b.Spec.Networking.Admin = &mesh_proto.EnvoyAdmin{Port: port}
	return b
}

func (b *zoneEgressBuilder) build() *core_mesh.ZoneEgressResource {
	return (*core_mesh.ZoneEgressResource)(b)
}

func newZoneIngress() *zoneIngressBuilder {
	return &zoneIngressBuilder{
		Spec: &mesh_proto.ZoneIngress{
			Networking: &mesh_proto.ZoneIngress_Networking{},
		},
	}
}

func (b *zoneIngressBuilder) meta(name string) *zoneIngressBuilder {
	b.Meta = &test_model.ResourceMeta{Name: name, Mesh: core_model.NoMesh}
	return b
}

func (b *zoneIngressBuilder) zone(name string) *zoneIngressBuilder {
	b.Spec.Zone = name
	return b
}

func (b *zoneIngressBuilder) address(address string) *zoneIngressBuilder {
	b.Spec.Networking.Address = address
	return b
}

func (b *zoneIngressBuilder) port(port uint32) *zoneIngressBuilder {
	b.Spec.Networking.Port = port
	return b
}

func (b *zoneIngressBuilder) advertisedAddress(address string) *zoneIngressBuilder {
	b.Spec.Networking.AdvertisedAddress = address
	return b
}

func (b *zoneIngressBuilder) advertisedPort(port uint32) *zoneIngressBuilder {
	b.Spec.Networking.AdvertisedPort = port
	return b
}

func (b *zoneIngressBuilder) admin(port uint32) *zoneIngressBuilder {
	b.Spec.Networking.Admin = &mesh_proto.EnvoyAdmin{Port: port}
	return b
}

func (b *zoneIngressBuilder) build() *core_mesh.ZoneIngressResource {
	return (*core_mesh.ZoneIngressResource)(b)
}

func newDataplane() *dataplaneBuilder {
	return &dataplaneBuilder{
		Spec: &mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Address: "1.1.1.1",
			},
		},
	}
}

func (b *dataplaneBuilder) build() *core_mesh.DataplaneResource {
	return (*core_mesh.DataplaneResource)(b)
}

func (b *dataplaneBuilder) meta(name, mesh string) *dataplaneBuilder {
	b.Meta = &test_model.ResourceMeta{Name: name, Mesh: mesh}
	return b
}

func (b *dataplaneBuilder) builtin(service string) *dataplaneBuilder {
	b.Spec.Networking.Gateway = &mesh_proto.Dataplane_Networking_Gateway{
		Tags: map[string]string{
			mesh_proto.ServiceTag: service,
		},
		Type: mesh_proto.Dataplane_Networking_Gateway_BUILTIN,
	}
	return b
}

func (b *dataplaneBuilder) inbound80to81(service, ip string) *dataplaneBuilder {
	b.Spec.Networking.Inbound = append(b.Spec.Networking.Inbound, &mesh_proto.Dataplane_Networking_Inbound{
		Address:     ip,
		Port:        80,
		ServicePort: 81,
		Tags: map[string]string{
			mesh_proto.ServiceTag:  service,
			mesh_proto.ProtocolTag: "http",
		},
	})
	return b
}

func (b *dataplaneBuilder) outbound8080(service, ip string) *dataplaneBuilder {
	b.Spec.Networking.Outbound = append(b.Spec.Networking.Outbound, &mesh_proto.Dataplane_Networking_Outbound{
		Address: ip,
		Port:    8080,
		Tags: map[string]string{
			mesh_proto.ServiceTag: service,
		},
	})
	return b
}

func (b *dataplaneBuilder) admin() *dataplaneBuilder {
	var port uint32 = 3301
	b.Spec.Networking.Admin = &mesh_proto.EnvoyAdmin{Port: port}
	return b
}

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
				newMesh("default"),
				newDataplane().
					meta("backend-1", "default").
					inbound80to81("backend", "192.168.0.1").
					outbound8080("redis", "192.168.0.2").
					outbound8080("gateway", "192.168.0.3").
					outbound8080("postgres", "192.168.0.4").
					outbound8080("web", "192.168.0.2").
					build(),
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
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect dataplane, empty response", testCase{
			path:    "/meshes/default/dataplanes/backend-1/policies",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_dataplane_empty-response.json")),
			resources: []core_model.Resource{
				newMesh("default"),
				newDataplane().
					meta("backend-1", "default").
					inbound80to81("backend", "192.168.0.1").
					outbound8080("redis", "192.168.0.2").
					outbound8080("gateway", "192.168.0.3").
					outbound8080("postgres", "192.168.0.4").
					outbound8080("web", "192.168.0.2").
					build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect gateway dataplane", testCase{
			path:    "/meshes/default/dataplanes/gateway-1/policies",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_gateway_dataplane.json")),
			resources: []core_model.Resource{
				newMesh("default"),
				newDataplane().
					meta("gateway-1", "default").
					builtin("gateway").
					build(),
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
				&core_mesh.MeshGatewayRouteResource{
					Meta: &test_model.ResourceMeta{Name: "route-1", Mesh: "default"},
					Spec: &mesh_proto.MeshGatewayRoute{
						Selectors: selectors{
							serviceSelector("gateway", ""),
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
							serviceSelector("gateway", ""),
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
				newMesh("default"),
				newDataplane().
					meta("gateway-1", "default").
					builtin("gateway").
					build(),
				newDataplane().
					meta("othergateway-1", "default").
					builtin("othergateway").
					build(),
				newDataplane().
					meta("redis-1", "default").
					inbound80to81("redis", "192.168.0.1").
					outbound8080("backend", "192.168.0.2").
					outbound8080("gateway", "192.168.0.3").
					build(),
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
				newMesh("default"),
				newDataplane().
					meta("gateway-1", "default").
					builtin("gateway").
					build(),
				newDataplane().
					meta("othergateway-1", "default").
					builtin("othergateway").
					build(),
				newDataplane().
					meta("redis-1", "default").
					inbound80to81("redis", "192.168.0.1").
					outbound8080("backend", "192.168.0.2").
					outbound8080("gateway", "192.168.0.3").
					build(),
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
				&core_mesh.MeshGatewayRouteResource{
					Meta: &test_model.ResourceMeta{Name: "gatewayroute", Mesh: "default"},
					Spec: &mesh_proto.MeshGatewayRoute{
						Selectors: selectors{
							serviceSelector("gateway", ""),
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
				newMesh("default"),
				&core_mesh.TrafficPermissionResource{
					Meta: &test_model.ResourceMeta{Name: "tp-1", Mesh: "default"},
					Spec: &mesh_proto.TrafficPermission{
						Sources: anyService(),
						Destinations: selectors{
							serviceSelector("backend", "http"),
							serviceSelector("redis", "http"),
							serviceSelector("gateway", "http"),
						},
					},
				},
				newDataplane().
					meta("backend-1", "default").
					inbound80to81("backend", "192.168.0.1").
					outbound8080("redis", "192.168.0.2").
					outbound8080("gateway", "192.168.0.3").
					build(),
				newDataplane().
					meta("redis-1", "default").
					inbound80to81("redis", "192.168.0.1").
					outbound8080("backend", "192.168.0.2").
					outbound8080("gateway", "192.168.0.3").
					build(),
				newDataplane().
					meta("gateway-1", "default").
					inbound80to81("gateway", "192.168.0.1").
					outbound8080("backend", "192.168.0.2").
					outbound8080("redis", "192.168.0.3").
					build(),
				newDataplane(). // not matched by TrafficPermission
						meta("web-1", "default").
						inbound80to81("web", "192.168.0.1").
						build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect fault injection", testCase{
			path:    "/meshes/mesh-1/fault-injections/fi-1/dataplanes",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_fault-injection.json")),
			resources: []core_model.Resource{
				newMesh("mesh-1"),
				&core_mesh.FaultInjectionResource{
					Meta: &test_model.ResourceMeta{Name: "fi-1", Mesh: "mesh-1"},
					Spec: &mesh_proto.FaultInjection{
						Sources: anyService(),
						Destinations: selectors{
							serviceSelector("backend", "http"),
							serviceSelector("redis", "http"),
							serviceSelector("gateway", "http"),
						},
						Conf: samples.FaultInjection.Conf,
					},
				},
				newDataplane().
					meta("backend-redis-1", "mesh-1").
					inbound80to81("backend", "192.168.0.1").
					inbound80to81("redis", "192.168.0.2").
					build(),
				newDataplane().
					meta("gateway-1", "mesh-1").
					inbound80to81("gateway", "192.168.0.1").
					outbound8080("backend", "192.168.0.2").
					outbound8080("redis", "192.168.0.3").
					build(),
				newDataplane(). // not matched by FaultInjection
						meta("web-1", "mesh-1").
						inbound80to81("web", "192.168.0.1").
						build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect rate limit", testCase{
			path:    "/meshes/mesh-1/rate-limits/rl-1/dataplanes",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_rate-limit.json")),
			resources: []core_model.Resource{
				newMesh("mesh-1"),
				&core_mesh.RateLimitResource{
					Meta: &test_model.ResourceMeta{Name: "rl-1", Mesh: "mesh-1"},
					Spec: &mesh_proto.RateLimit{
						Sources: anyService(),
						Destinations: selectors{
							serviceSelector("backend", "http"),
							serviceSelector("redis", "http"),
							serviceSelector("gateway", "http"),
							serviceSelector("es", ""),
						},
						Conf: samples.RateLimit.Conf,
					},
				},
				newDataplane().
					meta("gateway-1", "mesh-1").
					inbound80to81("gateway", "192.168.0.1").
					outbound8080("backend", "192.168.0.2").
					outbound8080("redis", "192.168.0.3").
					outbound8080("es", "192.168.0.4").
					build(),
				newDataplane(). // not matched by RateLimit
						meta("web-1", "mesh-1").
						inbound80to81("web", "192.168.0.1").
						build(),
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
				newMesh("mesh-1"),
				&core_mesh.TrafficLogResource{
					Meta: &test_model.ResourceMeta{Name: "tl-1", Mesh: "mesh-1"},
					Spec: &mesh_proto.TrafficLog{
						Sources: anyService(),
						Destinations: selectors{
							serviceSelector("redis", ""),
							serviceSelector("gateway", ""),
						},
						Conf: samples.TrafficLog.Conf,
					},
				},
				newDataplane().
					meta("backend-1", "mesh-1").
					inbound80to81("backend", "192.168.0.1").
					outbound8080("redis", "192.168.0.2").
					outbound8080("gateway", "192.168.0.3").
					outbound8080("web", "192.168.0.4").
					build(),
				newDataplane().
					meta("redis-1", "mesh-1").
					inbound80to81("redis", "192.168.0.1").
					outbound8080("gateway", "192.168.0.2").
					outbound8080("web", "192.168.0.4").
					build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect health check", testCase{
			path:    "/meshes/mesh-1/health-checks/hc-1/dataplanes",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_health-check.json")),
			resources: []core_model.Resource{
				newMesh("mesh-1"),
				&core_mesh.HealthCheckResource{
					Meta: &test_model.ResourceMeta{Name: "hc-1", Mesh: "mesh-1"},
					Spec: &mesh_proto.HealthCheck{
						Sources: anyService(),
						Destinations: selectors{
							serviceSelector("redis", ""),
							serviceSelector("gateway", ""),
						},
						Conf: samples.HealthCheck.Conf,
					},
				},
				newDataplane().
					meta("backend-1", "mesh-1").
					inbound80to81("backend", "192.168.0.1").
					outbound8080("redis", "192.168.0.2").
					outbound8080("gateway", "192.168.0.3").
					outbound8080("web", "192.168.0.4").
					build(),
				newDataplane().
					meta("redis-1", "mesh-1").
					inbound80to81("redis", "192.168.0.1").
					outbound8080("gateway", "192.168.0.2").
					outbound8080("web", "192.168.0.4").
					build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect circuit breaker", testCase{
			path:    "/meshes/mesh-1/circuit-breakers/cb-1/dataplanes",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_circuit-breaker.json")),
			resources: []core_model.Resource{
				newMesh("mesh-1"),
				&core_mesh.CircuitBreakerResource{
					Meta: &test_model.ResourceMeta{Name: "cb-1", Mesh: "mesh-1"},
					Spec: &mesh_proto.CircuitBreaker{
						Sources: anyService(),
						Destinations: selectors{
							serviceSelector("redis", ""),
							serviceSelector("gateway", ""),
						},
						Conf: samples.CircuitBreaker.Conf,
					},
				},
				newDataplane().
					meta("backend-1", "mesh-1").
					inbound80to81("backend", "192.168.0.1").
					outbound8080("redis", "192.168.0.2").
					outbound8080("gateway", "192.168.0.3").
					outbound8080("web", "192.168.0.4").
					build(),
				newDataplane().
					meta("redis-1", "mesh-1").
					inbound80to81("redis", "192.168.0.1").
					outbound8080("gateway", "192.168.0.2").
					outbound8080("web", "192.168.0.4").
					build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect retry", testCase{
			path:    "/meshes/mesh-1/retries/r-1/dataplanes",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_retry.json")),
			resources: []core_model.Resource{
				newMesh("mesh-1"),
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
							serviceSelector("gateway", ""),
						},
						Conf: samples.Retry.Conf,
					},
				},
				newDataplane().
					meta("meshgateway-1", "mesh-1").
					builtin("meshgateway").
					build(),
				newDataplane().
					meta("backend-1", "mesh-1").
					inbound80to81("backend", "192.168.0.1").
					outbound8080("redis", "192.168.0.2").
					outbound8080("gateway", "192.168.0.3").
					outbound8080("web", "192.168.0.4").
					build(),
				newDataplane().
					meta("redis-1", "mesh-1").
					inbound80to81("redis", "192.168.0.1").
					outbound8080("gateway", "192.168.0.2").
					outbound8080("web", "192.168.0.4").
					build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect timeout", testCase{
			path:    "/meshes/mesh-1/timeouts/t-1/dataplanes",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_timeout.json")),
			resources: []core_model.Resource{
				newMesh("mesh-1"),
				&core_mesh.TimeoutResource{
					Meta: &test_model.ResourceMeta{Name: "t-1", Mesh: "mesh-1"},
					Spec: &mesh_proto.Timeout{
						Sources: anyService(),
						Destinations: selectors{
							serviceSelector("redis", ""),
							serviceSelector("gateway", ""),
						},
						Conf: samples.Timeout.Conf,
					},
				},
				newDataplane().
					meta("backend-1", "mesh-1").
					inbound80to81("backend", "192.168.0.1").
					outbound8080("redis", "192.168.0.2").
					outbound8080("gateway", "192.168.0.3").
					outbound8080("web", "192.168.0.4").
					build(),
				newDataplane().
					meta("redis-1", "mesh-1").
					inbound80to81("redis", "192.168.0.1").
					outbound8080("gateway", "192.168.0.2").
					outbound8080("web", "192.168.0.4").
					build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect traffic route", testCase{
			path:    "/meshes/mesh-1/traffic-routes/t-1/dataplanes",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_traffic-route.json")),
			resources: []core_model.Resource{
				newMesh("mesh-1"),
				&core_mesh.TrafficRouteResource{
					Meta: &test_model.ResourceMeta{Name: "t-1", Mesh: "mesh-1"},
					Spec: &mesh_proto.TrafficRoute{
						Sources:      anyService(),
						Destinations: anyService(),
						Conf:         samples.TrafficRoute.Conf,
					},
				},
				newDataplane().
					meta("backend-1", "mesh-1").
					inbound80to81("backend", "192.168.0.1").
					outbound8080("redis", "192.168.0.2").
					outbound8080("web", "192.168.0.4").
					build(),
				newDataplane().
					meta("redis-1", "mesh-1").
					inbound80to81("redis", "192.168.0.1").
					outbound8080("gateway", "192.168.0.2").
					outbound8080("web", "192.168.0.4").
					build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect traffic trace", testCase{
			path:    "/meshes/mesh-1/traffic-traces/tt-1/dataplanes",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_traffic-trace.json")),
			resources: []core_model.Resource{
				newMesh("mesh-1"),
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
				newDataplane().
					meta("meshgateway-1", "mesh-1").
					builtin("meshgateway").
					build(),
				newDataplane().
					meta("backend-1", "mesh-1").
					inbound80to81("backend", "192.168.0.1").
					outbound8080("redis", "192.168.0.2").
					outbound8080("gateway", "192.168.0.3").
					outbound8080("web", "192.168.0.4").
					build(),
				newDataplane().
					meta("redis-1", "mesh-1").
					inbound80to81("redis", "192.168.0.1").
					outbound8080("gateway", "192.168.0.2").
					outbound8080("web", "192.168.0.4").
					build(),
				newDataplane().
					meta("web-1", "mesh-1").
					inbound80to81("web", "192.168.0.1").
					outbound8080("gateway", "192.168.0.2").
					outbound8080("backend", "192.168.0.4").
					build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect proxytemplate", testCase{
			path:    "/meshes/mesh-1/proxytemplates/tt-1/dataplanes",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_proxytemplate.json")),
			resources: []core_model.Resource{
				newMesh("mesh-1"),
				&core_mesh.ProxyTemplateResource{
					Meta: &test_model.ResourceMeta{Name: "tt-1", Mesh: "mesh-1"},
					Spec: &mesh_proto.ProxyTemplate{
						Selectors: anyService(),
						Conf:      samples.ProxyTemplate.Conf,
					},
				},
				newDataplane().
					meta("backend-1", "mesh-1").
					inbound80to81("backend", "192.168.0.1").
					outbound8080("redis", "192.168.0.2").
					outbound8080("gateway", "192.168.0.3").
					outbound8080("web", "192.168.0.4").
					build(),
				newDataplane().
					meta("redis-1", "mesh-1").
					inbound80to81("redis", "192.168.0.1").
					outbound8080("gateway", "192.168.0.2").
					outbound8080("web", "192.168.0.4").
					build(),
				newDataplane().
					meta("web-1", "mesh-1").
					inbound80to81("web", "192.168.0.1").
					outbound8080("gateway", "192.168.0.2").
					outbound8080("backend", "192.168.0.4").
					build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect traffic trace, empty response", testCase{
			path:    "/meshes/mesh-1/traffic-traces/tt-1/dataplanes",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_traffic-trace_empty-response.json")),
			resources: []core_model.Resource{
				newMesh("mesh-1"),
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
		Entry("inspect xds for dataplane", testCase{
			path:    "/meshes/mesh-1/dataplanes/backend-1/xds",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_xds_dataplane.json")),
			resources: []core_model.Resource{
				newMesh("mesh-1"),
				newDataplane().
					meta("backend-1", "mesh-1").
					admin().
					inbound80to81("backend", "192.168.0.1").
					outbound8080("redis", "192.168.0.2").
					outbound8080("gateway", "192.168.0.3").
					outbound8080("web", "192.168.0.4").
					build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect xds for local zone ingress", testCase{
			path:    "/zoneingresses/zi-1/xds",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_xds_local_zoneingress.json")),
			resources: []core_model.Resource{
				newZoneIngress().
					meta("zi-1").
					zone(""). // local zone ingress has empty "zone" field
					admin(2201).
					address("2.2.2.2").port(8080).
					advertisedAddress("3.3.3.3").advertisedPort(80).
					build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect xds for zone ingress from another zone", testCase{
			path:    "/zoneingresses/zi-1/xds",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_xds_remote_zoneingress.json")),
			resources: []core_model.Resource{
				newZoneIngress().
					meta("zi-1").
					zone("not-local-zone").
					admin(2201).
					address("2.2.2.2").port(8080).
					advertisedAddress("3.3.3.3").advertisedPort(80).
					build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect xds for zone ingress on global", testCase{
			global:  true,
			path:    "/zoneingresses/zi-1/xds",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_xds_local_zoneingress.json")),
			resources: []core_model.Resource{
				newZoneIngress().
					meta("zi-1").
					zone(""). // local zone ingress has empty "zone" field
					admin(2201).
					address("2.2.2.2").port(8080).
					advertisedAddress("3.3.3.3").advertisedPort(80).
					build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect xds for dataplane on global", testCase{
			global:  true,
			path:    "/meshes/mesh-1/dataplanes/backend-1/xds",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_xds_dataplane.json")),
			resources: []core_model.Resource{
				newMesh("mesh-1"),
				newDataplane().
					meta("backend-1", "mesh-1").
					admin().
					inbound80to81("backend", "192.168.0.1").
					outbound8080("redis", "192.168.0.2").
					outbound8080("gateway", "192.168.0.3").
					outbound8080("web", "192.168.0.4").
					build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect xds for zone egress", testCase{
			path:    "/zoneegresses/ze-1/xds",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_xds_zoneegress.json")),
			resources: []core_model.Resource{
				newZoneEgress().
					meta("ze-1").
					address("4.4.4.4").
					port(8080).
					admin(4321).
					build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect stats for dataplane", testCase{
			path:    "/meshes/mesh-1/dataplanes/backend-1/stats",
			matcher: matchers.MatchGoldenEqual(path.Join("testdata", "inspect_stats_dataplane.out")),
			resources: []core_model.Resource{
				newMesh("mesh-1"),
				newDataplane().
					meta("backend-1", "mesh-1").
					admin().
					inbound80to81("backend", "192.168.0.1").
					outbound8080("redis", "192.168.0.2").
					outbound8080("gateway", "192.168.0.3").
					outbound8080("web", "192.168.0.4").
					build(),
			},
			contentType: "text/plain",
		}),
		Entry("inspect clusters for dataplane", testCase{
			path:    "/meshes/mesh-1/dataplanes/backend-1/clusters",
			matcher: matchers.MatchGoldenEqual(path.Join("testdata", "inspect_clusters_dataplane.out")),
			resources: []core_model.Resource{
				newMesh("mesh-1"),
				newDataplane().
					meta("backend-1", "mesh-1").
					admin().
					inbound80to81("backend", "192.168.0.1").
					outbound8080("redis", "192.168.0.2").
					outbound8080("gateway", "192.168.0.3").
					outbound8080("web", "192.168.0.4").
					build(),
			},
			contentType: "text/plain",
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
			newMesh("default"),
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
			newDataplane().
				meta("backend-1", "default").
				inbound80to81("backend", "192.168.0.1").
				outbound8080("redis", "192.168.0.2").
				outbound8080("gateway", "192.168.0.3").
				build(),
			newDataplane().
				meta("redis-1", "default").
				inbound80to81("redis", "192.168.0.1").
				outbound8080("backend", "192.168.0.2").
				outbound8080("gateway", "192.168.0.3").
				build(),
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
