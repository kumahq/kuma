package generator_test

import (
	"context"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	model "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/xds"
	test_xds "github.com/kumahq/kuma/pkg/test/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/cache/cla"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ = Describe("OutboundProxyGenerator", func() {
	meta := &test_model.ResourceMeta{
		Name: "mesh1",
	}
	mesh2Meta := &test_model.ResourceMeta{
		Name: "mesh2",
	}
	logging := &mesh_proto.Logging{
		Backends: []*mesh_proto.LoggingBackend{
			{
				Name: "file",
				Type: mesh_proto.LoggingFileType,
				Conf: util_proto.MustToStruct(&mesh_proto.FileLoggingBackendConfig{
					Path: "/var/log",
				}),
			},
			{
				Name: "elk",
				Type: mesh_proto.LoggingTcpType,
				Conf: util_proto.MustToStruct(&mesh_proto.TcpLoggingBackendConfig{
					Address: "logstash:1234",
				}),
			},
		},
	}

	defaultTrafficRoute := &core_mesh.TrafficRouteResourceList{
		Items: []*core_mesh.TrafficRouteResource{{
			Meta: &test_model.ResourceMeta{Name: "default-allow-all"},
			Spec: &mesh_proto.TrafficRoute{
				Sources: []*mesh_proto.Selector{{
					Match: mesh_proto.MatchAnyService(),
				}},
				Destinations: []*mesh_proto.Selector{{
					Match: mesh_proto.MatchAnyService(),
				}},
				Conf: &mesh_proto.TrafficRoute_Conf{
					Destination: mesh_proto.MatchAnyService(),
					LoadBalancer: &mesh_proto.TrafficRoute_LoadBalancer{
						LbType: &mesh_proto.TrafficRoute_LoadBalancer_RoundRobin_{},
					},
				},
			},
		}},
	}

	timeout := &mesh_proto.Timeout{
		Conf: &mesh_proto.Timeout_Conf{
			ConnectTimeout: util_proto.Duration(100 * time.Second),
			Tcp: &mesh_proto.Timeout_Conf_Tcp{
				IdleTimeout: util_proto.Duration(101 * time.Second),
			},
			Http: &mesh_proto.Timeout_Conf_Http{
				RequestTimeout: util_proto.Duration(102 * time.Second),
				IdleTimeout:    util_proto.Duration(103 * time.Second),
			},
			Grpc: &mesh_proto.Timeout_Conf_Grpc{
				StreamIdleTimeout: util_proto.Duration(104 * time.Second),
				MaxStreamDuration: util_proto.Duration(105 * time.Second),
			},
		},
	}

	plainCtx := xds_context.Context{
		ControlPlane: &xds_context.ControlPlaneContext{},
		Mesh: xds_context.MeshContext{
			Resource: &core_mesh.MeshResource{
				Meta: meta,
				Spec: &mesh_proto.Mesh{
					Logging: logging,
				},
			},
			Resources: xds_context.Resources{
				MeshLocalResources: xds_context.ResourceMap{
					core_mesh.TrafficRouteType: defaultTrafficRoute,
				},
			},
		},
	}

	mtlsCtx := xds_context.Context{
		ControlPlane: &xds_context.ControlPlaneContext{
			Secrets: &xds.TestSecrets{},
		},
		Mesh: xds_context.MeshContext{
			Resources: xds_context.Resources{
				MeshLocalResources: xds_context.ResourceMap{
					core_mesh.TrafficRouteType: defaultTrafficRoute,
				},
			},
			Resource: &core_mesh.MeshResource{
				Spec: &mesh_proto.Mesh{
					Mtls: &mesh_proto.Mesh_Mtls{
						EnabledBackend: "builtin",
						Backends: []*mesh_proto.CertificateAuthorityBackend{
							{
								Name: "builtin",
								Type: "builtin",
							},
						},
					},
					Logging: logging,
				},
				Meta: meta,
			},
		},
	}

	serviceVipCtx := xds_context.Context{
		ControlPlane: &xds_context.ControlPlaneContext{
			Secrets: &xds.TestSecrets{},
		},
		Mesh: xds_context.MeshContext{
			Resources: xds_context.Resources{
				MeshLocalResources: xds_context.ResourceMap{
					core_mesh.TrafficRouteType: defaultTrafficRoute,
				},
			},
			Resource: &core_mesh.MeshResource{
				Spec: &mesh_proto.Mesh{
					Mtls: &mesh_proto.Mesh_Mtls{
						EnabledBackend: "builtin",
						Backends: []*mesh_proto.CertificateAuthorityBackend{
							{
								Name: "builtin",
								Type: "builtin",
							},
						},
					},
					Logging: logging,
				},
				Meta: meta,
			},
			VIPDomains: []xds_types.VIPDomains{
				{
					Address: "240.0.0.3",
					Domains: []string{"backend"},
				},
				{
					Address: "240.0.0.4",
					Domains: []string{"backend"},
				},
			},
		},
	}

	crossMeshCtx := xds_context.Context{
		ControlPlane: &xds_context.ControlPlaneContext{
			Secrets: &xds.TestSecrets{},
		},
		Mesh: xds_context.MeshContext{
			Resource: &core_mesh.MeshResource{
				Spec: &mesh_proto.Mesh{
					Mtls: &mesh_proto.Mesh_Mtls{
						EnabledBackend: "builtin",
						Backends: []*mesh_proto.CertificateAuthorityBackend{
							{
								Name: "builtin",
								Type: "builtin",
							},
						},
					},
					Logging: logging,
				},
				Meta: meta,
			},
			Resources: xds_context.Resources{
				MeshLocalResources: map[core_model.ResourceType]core_model.ResourceList{
					core_mesh.TrafficRouteType: defaultTrafficRoute,
					core_mesh.MeshType: &core_mesh.MeshResourceList{
						Items: []*core_mesh.MeshResource{{
							Spec: &mesh_proto.Mesh{
								Mtls: &mesh_proto.Mesh_Mtls{
									EnabledBackend: "builtin",
									Backends: []*mesh_proto.CertificateAuthorityBackend{
										{
											Name: "builtin",
											Type: "builtin",
										},
									},
								},
								Logging: logging,
							},
							Meta: mesh2Meta,
						}},
					},
				},
				CrossMeshResources: map[string]xds_context.ResourceMap{
					"mesh-2": {
						core_mesh.MeshGatewayType: &core_mesh.MeshGatewayResourceList{
							Items: []*core_mesh.MeshGatewayResource{{
								Meta: &test_model.ResourceMeta{
									Name: "mesh2",
								},
								Spec: &mesh_proto.MeshGateway{
									Conf: &mesh_proto.MeshGateway_Conf{
										Listeners: []*mesh_proto.MeshGateway_Listener{{
											Hostname: "gateway1.mesh",
											Port:     80,
											Protocol: mesh_proto.MeshGateway_Listener_HTTP,
											Tags: map[string]string{
												"listener": "internal",
											},
										}, {
											Hostname: "*",
											Port:     80,
											Protocol: mesh_proto.MeshGateway_Listener_HTTP,
											Tags: map[string]string{
												"listener": "wildcard",
											},
										}},
									},
									Selectors: []*mesh_proto.Selector{{
										Match: map[string]string{
											mesh_proto.ServiceTag: "gateway",
										},
									}},
									Tags: map[string]string{
										"gateway": "prod",
									},
								},
							}},
						},
					},
				},
			},
			CrossMeshEndpoints: map[model.MeshName]model.EndpointMap{
				"mesh2": {
					"api-http": []model.Endpoint{ // notice that all endpoints have tag `kuma.io/protocol: http`
						{
							Target: "192.168.0.6",
							Port:   8086,
							Tags:   map[string]string{"kuma.io/service": "api-http", "kuma.io/protocol": "http", "region": "eu"},
							Weight: 1,
						},
					},
				},
			},
			ServicesInformation: map[string]*xds_context.ServiceInformation{
				"api-http": {
					Protocol: core_mesh.ProtocolHTTP,
				},
			},
		},
	}

	type testCase struct {
		ctx       xds_context.Context
		dataplane string
		expected  string
	}

	DescribeTable("Generate Envoy xDS resources",
		func(given testCase) {
			// setup
			gen := &generator.OutboundProxyGenerator{}

			dataplane := &mesh_proto.Dataplane{}
			Expect(util_proto.FromYAML([]byte(given.dataplane), dataplane)).To(Succeed())

			outboundTargets := model.EndpointMap{
				"api-http": []model.Endpoint{ // notice that all endpoints have tag `kuma.io/protocol: http`
					{
						Target: "192.168.0.4",
						Port:   8084,
						Tags:   map[string]string{"kuma.io/service": "api-http", "kuma.io/protocol": "http", "region": "us"},
						Weight: 1,
					},
					{
						Target: "192.168.0.5",
						Port:   8085,
						Tags:   map[string]string{"kuma.io/service": "api-http", "kuma.io/protocol": "http", "region": "eu"},
						Weight: 1,
					},
				},
				"api-tcp": []model.Endpoint{ // notice that not every endpoint has a `kuma.io/protocol: http` tag
					{
						Target: "192.168.0.6",
						Port:   8086,
						Tags:   map[string]string{"kuma.io/service": "api-tcp", "kuma.io/protocol": "http", "region": "us"},
						Weight: 1,
					},
					{
						Target: "192.168.0.7",
						Port:   8087,
						Tags:   map[string]string{"kuma.io/service": "api-tcp", "region": "eu"},
						Weight: 1,
					},
				},
				"api-http2": []model.Endpoint{ // notice that all endpoints have tag `kuma.io/protocol: http2`
					{
						Target: "192.168.0.4",
						Port:   8088,
						Tags:   map[string]string{"kuma.io/service": "api-http2", "kuma.io/protocol": "http2"},
						Weight: 1,
					},
				},
				"api-grpc": []model.Endpoint{ // notice that all endpoints have tag `kuma.io/protocol: grpc`
					{
						Target: "192.168.0.4",
						Port:   8089,
						Tags:   map[string]string{"kuma.io/service": "api-grpc", "kuma.io/protocol": "grpc"},
						Weight: 1,
					},
				},
				"backend": []model.Endpoint{ // notice that not every endpoint has a tag `kuma.io/protocol: http`
					{
						Target: "192.168.0.1",
						Port:   8081,
						Tags:   map[string]string{"kuma.io/service": "backend", "region": "us"},
						Weight: 1,
					},
					{
						Target: "192.168.0.2",
						Port:   8082,
						Weight: 1,
					},
				},
				"db": []model.Endpoint{
					{
						Target: "192.168.0.3",
						Port:   5432,
						Tags:   map[string]string{"kuma.io/service": "db", "role": "master"},
						Weight: 1,
					},
					{
						Target: "192.168.0.3",
						Port:   5433,
						Tags:   map[string]string{"kuma.io/service": "db", "role": "replica"},
						Weight: 1,
					},
				},
			}

			esOutboundTargets := model.EndpointMap{
				"es": []model.Endpoint{
					{
						Target:          "10.0.0.1",
						Port:            10001,
						Tags:            map[string]string{"kuma.io/service": "es", "kuma.io/protocol": "http"},
						Weight:          1,
						ExternalService: &model.ExternalService{TLSEnabled: false},
					},
				},
				"es2": []model.Endpoint{
					{
						Target:          "10.0.0.2",
						Port:            10002,
						Tags:            map[string]string{"kuma.io/service": "es2", "kuma.io/protocol": "http2"},
						Weight:          1,
						ExternalService: &model.ExternalService{TLSEnabled: false},
					},
				},
			}

			meshes := []string{given.ctx.Mesh.Resource.Meta.GetName()}
			if given.ctx.Mesh.Resources.MeshLocalResources != nil {
				if meshResources, ok := given.ctx.Mesh.Resources.MeshLocalResources[core_mesh.MeshType]; ok {
					for _, mesh := range meshResources.GetItems() {
						meshes = append(meshes, mesh.GetMeta().GetName())
					}
				}
			}

			proxy := &model.Proxy{
				Id: *model.BuildProxyId("default", "side-car"),
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Name:    "dp-1",
						Mesh:    given.ctx.Mesh.Resource.Meta.GetName(),
						Version: "1",
					},
					Spec: dataplane,
				},
				SecretsTracker: envoy_common.NewSecretsTracker(given.ctx.Mesh.Resource.Meta.GetName(), meshes),
				APIVersion:     envoy_common.APIV3,
				Routing: model.Routing{
					TrafficRoutes: model.RouteMap{
						mesh_proto.OutboundInterface{
							DataplaneIP:   "127.0.0.1",
							DataplanePort: 40001,
						}: &core_mesh.TrafficRouteResource{
							Spec: &mesh_proto.TrafficRoute{
								Conf: &mesh_proto.TrafficRoute_Conf{
									Destination: mesh_proto.MatchService("api-http"),
									LoadBalancer: &mesh_proto.TrafficRoute_LoadBalancer{
										LbType: &mesh_proto.TrafficRoute_LoadBalancer_RoundRobin_{},
									},
								},
							},
						},
						mesh_proto.OutboundInterface{
							DataplaneIP:   "127.0.0.1",
							DataplanePort: 40002,
						}: &core_mesh.TrafficRouteResource{
							Spec: &mesh_proto.TrafficRoute{
								Conf: &mesh_proto.TrafficRoute_Conf{
									Destination: mesh_proto.MatchService("api-tcp"),
									LoadBalancer: &mesh_proto.TrafficRoute_LoadBalancer{
										LbType: &mesh_proto.TrafficRoute_LoadBalancer_LeastRequest_{
											LeastRequest: &mesh_proto.TrafficRoute_LoadBalancer_LeastRequest{
												ChoiceCount: 4,
											},
										},
									},
								},
							},
						},
						mesh_proto.OutboundInterface{
							DataplaneIP:   "127.0.0.1",
							DataplanePort: 40003,
						}: &core_mesh.TrafficRouteResource{
							Spec: &mesh_proto.TrafficRoute{
								Conf: &mesh_proto.TrafficRoute_Conf{
									Destination: mesh_proto.MatchService("api-http2"),
									LoadBalancer: &mesh_proto.TrafficRoute_LoadBalancer{
										LbType: &mesh_proto.TrafficRoute_LoadBalancer_RingHash_{
											RingHash: &mesh_proto.TrafficRoute_LoadBalancer_RingHash{
												HashFunction: "MURMUR_HASH_2",
												MinRingSize:  64,
												MaxRingSize:  1024,
											},
										},
									},
								},
							},
						},
						mesh_proto.OutboundInterface{
							DataplaneIP:   "127.0.0.1",
							DataplanePort: 40004,
						}: &core_mesh.TrafficRouteResource{
							Spec: &mesh_proto.TrafficRoute{
								Conf: &mesh_proto.TrafficRoute_Conf{
									Destination: mesh_proto.MatchService("api-grpc"),
									LoadBalancer: &mesh_proto.TrafficRoute_LoadBalancer{
										LbType: &mesh_proto.TrafficRoute_LoadBalancer_Random_{},
									},
								},
							},
						},
						mesh_proto.OutboundInterface{
							DataplaneIP:   "127.0.0.1",
							DataplanePort: 18080,
						}: &core_mesh.TrafficRouteResource{
							Spec: &mesh_proto.TrafficRoute{
								Conf: &mesh_proto.TrafficRoute_Conf{
									Destination: mesh_proto.MatchService("backend"),
									LoadBalancer: &mesh_proto.TrafficRoute_LoadBalancer{
										LbType: &mesh_proto.TrafficRoute_LoadBalancer_Maglev_{},
									},
								},
							},
						},
						mesh_proto.OutboundInterface{
							DataplaneIP:   "127.0.0.1",
							DataplanePort: 54321,
						}: &core_mesh.TrafficRouteResource{
							Spec: &mesh_proto.TrafficRoute{
								Conf: &mesh_proto.TrafficRoute_Conf{
									Split: []*mesh_proto.TrafficRoute_Split{{
										Weight:      util_proto.UInt32(10),
										Destination: mesh_proto.TagSelector{"kuma.io/service": "db", "role": "master"},
									}, {
										Weight:      util_proto.UInt32(90),
										Destination: mesh_proto.TagSelector{"kuma.io/service": "db", "role": "replica"},
									}, {
										Weight: util_proto.UInt32(0),
										// should be excluded from Envoy configuration
										Destination: mesh_proto.TagSelector{"kuma.io/service": "db", "role": "canary"},
									}},
								},
							},
						},
						mesh_proto.OutboundInterface{
							DataplaneIP:   "127.0.0.1",
							DataplanePort: 18081,
						}: &core_mesh.TrafficRouteResource{
							Spec: &mesh_proto.TrafficRoute{
								Conf: &mesh_proto.TrafficRoute_Conf{
									Destination: mesh_proto.TagSelector{"kuma.io/service": "es", "kuma.io/protocol": "http"},
								},
							},
						},
						mesh_proto.OutboundInterface{
							DataplaneIP:   "127.0.0.1",
							DataplanePort: 18082,
						}: &core_mesh.TrafficRouteResource{
							Spec: &mesh_proto.TrafficRoute{
								Conf: &mesh_proto.TrafficRoute_Conf{
									Destination: mesh_proto.TagSelector{"kuma.io/service": "es2", "kuma.io/protocol": "http2"},
								},
							},
						},
						mesh_proto.OutboundInterface{
							DataplaneIP:   "127.0.0.1",
							DataplanePort: 4040,
						}: nil,
						mesh_proto.OutboundInterface{
							DataplaneIP:   "240.0.0.0",
							DataplanePort: 80,
						}: &core_mesh.TrafficRouteResource{
							Spec: &mesh_proto.TrafficRoute{
								Conf: &mesh_proto.TrafficRoute_Conf{
									Split: []*mesh_proto.TrafficRoute_Split{{
										Weight:      util_proto.UInt32(10),
										Destination: mesh_proto.TagSelector{"kuma.io/service": "es2", "role": "master"},
									}, {
										Weight:      util_proto.UInt32(90),
										Destination: mesh_proto.TagSelector{"kuma.io/service": "es2", "role": "replica"},
									}},
								},
							},
						},
						mesh_proto.OutboundInterface{
							DataplaneIP:   "240.0.0.1",
							DataplanePort: 80,
						}: &core_mesh.TrafficRouteResource{
							Spec: &mesh_proto.TrafficRoute{
								Conf: &mesh_proto.TrafficRoute_Conf{
									Split: []*mesh_proto.TrafficRoute_Split{{
										Weight:      util_proto.UInt32(10),
										Destination: mesh_proto.TagSelector{"kuma.io/service": "es2", "role": "master"},
									}, {
										Weight:      util_proto.UInt32(90),
										Destination: mesh_proto.TagSelector{"kuma.io/service": "es2", "role": "replica"},
									}},
								},
							},
						},
						mesh_proto.OutboundInterface{
							DataplaneIP:   "240.0.0.2",
							DataplanePort: 80,
						}: &core_mesh.TrafficRouteResource{
							Spec: &mesh_proto.TrafficRoute{
								Conf: &mesh_proto.TrafficRoute_Conf{
									Split: []*mesh_proto.TrafficRoute_Split{{
										Weight:      util_proto.UInt32(10),
										Destination: mesh_proto.TagSelector{"kuma.io/service": "es2", "role": "master"},
									}, {
										Weight:      util_proto.UInt32(90),
										Destination: mesh_proto.TagSelector{"kuma.io/service": "es2", "role": "replica"},
									}},
								},
							},
						},
						mesh_proto.OutboundInterface{
							DataplaneIP:   "127.0.0.1",
							DataplanePort: 30001,
						}: &core_mesh.TrafficRouteResource{
							Spec: &mesh_proto.TrafficRoute{
								Conf: &mesh_proto.TrafficRoute_Conf{
									Destination: mesh_proto.MatchTags(map[string]string{
										mesh_proto.ServiceTag: "api-http",
										"kuma.io/mesh":        "mesh2",
									}),
									LoadBalancer: &mesh_proto.TrafficRoute_LoadBalancer{
										LbType: &mesh_proto.TrafficRoute_LoadBalancer_RoundRobin_{},
									},
								},
							},
						},
					},
					OutboundTargets:                outboundTargets,
					ExternalServiceOutboundTargets: esOutboundTargets,
				},
				Policies: model.MatchedPolicies{
					TrafficLogs: model.TrafficLogMap{
						"api-http": &core_mesh.TrafficLogResource{
							Spec: &mesh_proto.TrafficLog{
								Conf: &mesh_proto.TrafficLog_Conf{
									Backend: "file",
								},
							},
						},
						"api-tcp": &core_mesh.TrafficLogResource{
							Spec: &mesh_proto.TrafficLog{
								Conf: &mesh_proto.TrafficLog_Conf{
									Backend: "elk",
								},
							},
						},
					},
					CircuitBreakers: model.CircuitBreakerMap{
						"api-http": &core_mesh.CircuitBreakerResource{
							Spec: &mesh_proto.CircuitBreaker{
								Conf: &mesh_proto.CircuitBreaker_Conf{
									Detectors: &mesh_proto.CircuitBreaker_Conf_Detectors{
										TotalErrors: &mesh_proto.CircuitBreaker_Conf_Detectors_Errors{},
									},
								},
							},
						},
					},
					Timeouts: map[mesh_proto.OutboundInterface]*core_mesh.TimeoutResource{
						{DataplaneIP: "127.0.0.1", DataplanePort: 40002}: {Spec: timeout},
						{DataplaneIP: "127.0.0.1", DataplanePort: 40003}: {Spec: timeout},
						{DataplaneIP: "127.0.0.1", DataplanePort: 40004}: {Spec: timeout},
						{DataplaneIP: "127.0.0.1", DataplanePort: 18082}: {Spec: timeout},
					},
				},
				Metadata: &model.DataplaneMetadata{
					Features: model.Features{"feature-tcp-accesslog-via-named-pipe": true},
				},
			}

			// when
			metrics, err := core_metrics.NewMetrics("Zone")
			Expect(err).ToNot(HaveOccurred())
			given.ctx.Mesh.EndpointMap = outboundTargets
			given.ctx.Mesh.ServicesInformation = map[string]*xds_context.ServiceInformation{
				"api-http": {
					TLSReadiness: true,
					Protocol:     core_mesh.ProtocolHTTP,
				},
				"api-tcp": {
					TLSReadiness: true,
					Protocol:     core_mesh.ProtocolTCP,
				},
				"api-http2": {
					TLSReadiness: true,
					Protocol:     core_mesh.ProtocolHTTP2,
				},
				"api-grpc": {
					TLSReadiness: true,
					Protocol:     core_mesh.ProtocolGRPC,
				},
				"backend": {
					TLSReadiness: true,
					Protocol:     core_mesh.ProtocolUnknown,
				},
				"db": {
					TLSReadiness: true,
					Protocol:     core_mesh.ProtocolUnknown,
				},
				"es": {
					TLSReadiness: true,
					Protocol:     core_mesh.ProtocolHTTP,
				},
				"es2": {
					TLSReadiness: true,
					Protocol:     core_mesh.ProtocolHTTP2,
				},
			}
			given.ctx.ControlPlane.CLACache, err = cla.NewCache(0*time.Second, metrics)
			Expect(err).ToNot(HaveOccurred())
			rs, err := gen.Generate(context.Background(), nil, given.ctx, proxy)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			resp, err := rs.List().ToDeltaDiscoveryResponse()
			// then
			Expect(err).ToNot(HaveOccurred())
			// when
			actual, err := util_proto.ToYAML(resp)
			// then
			Expect(err).ToNot(HaveOccurred())

			// and output matches golden files
			Expect(actual).To(MatchGoldenYAML(filepath.Join("testdata", "outbound-proxy", given.expected)))
		},
		Entry("01. transparent_proxying=false, mtls=false, outbound=0", testCase{
			ctx:       plainCtx,
			dataplane: ``,
			expected:  "01.envoy.golden.yaml",
		}),
		Entry("02. transparent_proxying=true, mtls=false, outbound=0", testCase{
			ctx: mtlsCtx,
			dataplane: `
            networking:
              transparentProxying:
                redirectPort: 15001
`,
			expected: "02.envoy.golden.yaml",
		}),
		Entry("03. transparent_proxying=false, mtls=false, outbound=4", testCase{
			ctx: plainCtx,
			dataplane: `
            networking:
              address: 10.0.0.1
              gateway:
                tags:
                  kuma.io/service: gateway
              outbound:
              - port: 18080
                tags:
                  kuma.io/service: backend
              - port: 54321
                tags:
                  kuma.io/service: db
              - port: 40001
                tags:
                  kuma.io/service: api-http
              - port: 40002
                tags:
                  kuma.io/service: api-tcp
              - port: 40003
                tags:
                  kuma.io/service: api-http2
              - port: 40004
                tags:
                  kuma.io/service: api-grpc
`,
			expected: "03.envoy.golden.yaml",
		}),
		Entry("04. transparent_proxying=true, mtls=true, outbound=4", testCase{
			ctx: mtlsCtx,
			dataplane: `
            networking:
              address: 10.0.0.1
              inbound:
              - port: 8080
                tags:
                  kuma.io/service: web
              outbound:
              - port: 18080
                tags:
                  kuma.io/service: backend
              - port: 54321
                tags:
                  kuma.io/service: db
              - port: 40001
                tags:
                  kuma.io/service: api-http
              - port: 40002
                tags:
                  kuma.io/service: api-tcp
              transparentProxying:
                redirectPortOutbound: 15001
                redirectPortInbound: 15006
`,
			expected: "04.envoy.golden.yaml",
		}),
		Entry("05. transparent_proxying=true, mtls=true, outbound=1 with ExternalService", testCase{
			ctx: mtlsCtx,
			dataplane: `
            networking:
              address: 10.0.0.1
              inbound:
              - port: 8080
                tags:
                  kuma.io/service: web
              outbound:
              - port: 18081
                tags:
                  kuma.io/service: es
              transparentProxying:
                redirectPortOutbound: 15001
                redirectPortInbound: 15006
`,
			expected: "05.envoy.golden.yaml",
		}),
		Entry("06. transparent_proxying=true, mtls=true, outbound=1 with ExternalService http2", testCase{
			ctx: mtlsCtx,
			dataplane: `
            networking:
              address: 10.0.0.1
              inbound:
              - port: 8080
                tags:
                  kuma.io/service: web
              outbound:
              - port: 18082
                tags:
                  kuma.io/service: es2
              transparentProxying:
                redirectPortOutbound: 15001
                redirectPortInbound: 15006
`,
			expected: "06.envoy.golden.yaml",
		}),
		Entry("07. no TrafficRoute", testCase{
			ctx: mtlsCtx,
			dataplane: `
            networking:
              address: 10.0.0.1
              inbound:
              - port: 8080
                tags:
                  kuma.io/service: web
              outbound:
              - port: 4040
                tags:
                  kuma.io/service: service-without-traffic-route
`,
			expected: "07.envoy.golden.yaml",
		}),
		Entry("08. several outbounds for the same external service with TrafficRoute", testCase{
			ctx: mtlsCtx,
			dataplane: `
            networking:
              address: 10.0.0.1
              inbound:
              - port: 8080
                tags:
                  kuma.io/service: web
              outbound:
              - port: 80
                address: 240.0.0.0
                tags:
                  kuma.io/service: es2
              - port: 80
                address: 240.0.0.1
                tags:
                  kuma.io/service: es2
              - port: 80
                address: 240.0.0.2
                tags:
                  kuma.io/service: es2
              transparentProxying:
                redirectPortOutbound: 15001
                redirectPortInbound: 15006
`,
			expected: "08.envoy.golden.yaml",
		}),
		Entry("09. cross-mesh", testCase{
			ctx: crossMeshCtx,
			dataplane: `
            networking:
              address: 10.0.0.1
              inbound:
              - port: 8080
                tags:
                  kuma.io/service: web
              outbound:
              - port: 30001
                tags:
                  kuma.io/mesh: mesh2
                  kuma.io/service: api-http
`,
			expected: "09.envoy.golden.yaml",
		}),
		Entry("10. service vips", testCase{
			ctx: serviceVipCtx,
			dataplane: `
            networking:
              address: 10.0.0.1
              inbound:
              - port: 8080
                tags:
                  kuma.io/service: web
              outbound:
              - port: 18080
                tags:
                  kuma.io/service: backend
              - port: 80
                address: 240.0.0.3
                tags:
                  kuma.io/service: backend
              - port: 80
                address: 240.0.0.4
                tags:
                  kuma.io/service: backend
              - port: 8080
                address: 240.0.0.4
                tags:
                  kuma.io/service: backend
              transparentProxying:
                redirectPortOutbound: 15001
                redirectPortInbound: 15006
`,
			expected: "10.envoy.golden.yaml",
		}),
		Entry("11. service vips with outbound of multiple tags (headless service)", testCase{
			ctx: serviceVipCtx,
			dataplane: `
            networking:
              address: 10.0.0.1
              inbound:
              - port: 8080
                tags:
                  kuma.io/service: web
              outbound:
              - port: 80
                address: 240.0.0.3
                tags:
                  kuma.io/service: backend
              - port: 80
                address: 10.0.0.1
                tags:
                  kuma.io/service: backend
                  kuma.io/instance: instance-1
              - port: 80
                address: 10.0.0.2
                tags:
                  kuma.io/service: backend
                  kuma.io/instance: instance-2
              transparentProxying:
                redirectPortOutbound: 15001
                redirectPortInbound: 15006
`,
			expected: "11.envoy.golden.yaml",
		}),
	)

	It("Add sanitized alternative cluster name for stats", func() {
		// setup
		gen := &generator.OutboundProxyGenerator{}
		dp := `
        networking:
          outbound:
          - port: 18080
            tags:
              kuma.io/service: backend.kuma-system
          - port: 54321
            tags:
              kuma.io/service: db.kuma-system`

		dataplane := &mesh_proto.Dataplane{}
		Expect(util_proto.FromYAML([]byte(dp), dataplane)).To(Succeed())

		outboundTargets := model.EndpointMap{
			"backend.kuma-system": []model.Endpoint{
				{
					Target: "192.168.0.1",
					Port:   8082,
					Weight: 1,
				},
			},
			"db.kuma-system": []model.Endpoint{
				{
					Target: "192.168.0.2",
					Port:   5432,
					Tags:   map[string]string{"kuma.io/service": "db", "role": "master"},
					Weight: 1,
				},
			},
		}
		proxy := &model.Proxy{
			Id: *model.BuildProxyId("default", "side-car"),
			Dataplane: &core_mesh.DataplaneResource{
				Meta: &test_model.ResourceMeta{
					Version: "1",
				},
				Spec: dataplane,
			},
			APIVersion: envoy_common.APIV3,
			Routing: model.Routing{
				TrafficRoutes: model.RouteMap{
					mesh_proto.OutboundInterface{
						DataplaneIP:   "127.0.0.1",
						DataplanePort: 18080,
					}: &core_mesh.TrafficRouteResource{
						Spec: &mesh_proto.TrafficRoute{
							Conf: &mesh_proto.TrafficRoute_Conf{
								Destination: mesh_proto.MatchService("backend.kuma-system"),
							},
						},
					},
					mesh_proto.OutboundInterface{
						DataplaneIP:   "127.0.0.1",
						DataplanePort: 54321,
					}: &core_mesh.TrafficRouteResource{
						Spec: &mesh_proto.TrafficRoute{
							Conf: &mesh_proto.TrafficRoute_Conf{
								Destination: mesh_proto.TagSelector{"kuma.io/service": "db", "version": "3.2.0"},
							},
						},
					},
				},
				OutboundTargets: outboundTargets,
			},
			Metadata: &model.DataplaneMetadata{},
		}

		// when
		plainCtx.ControlPlane.CLACache = &test_xds.DummyCLACache{OutboundTargets: outboundTargets}
		plainCtx.Mesh.ServicesInformation = map[string]*xds_context.ServiceInformation{
			"backend.kuma-system": {
				Protocol: core_mesh.ProtocolUnknown,
			},
			"db.kuma-system": {
				Protocol: core_mesh.ProtocolUnknown,
			},
		}
		rs, err := gen.Generate(context.Background(), nil, plainCtx, proxy)

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		resp, err := rs.List().ToDeltaDiscoveryResponse()
		// then
		Expect(err).ToNot(HaveOccurred())
		// when
		actual, err := util_proto.ToYAML(resp)
		// then
		Expect(err).ToNot(HaveOccurred())

		// and output matches golden files
		Expect(actual).To(MatchGoldenYAML(filepath.Join("testdata", "outbound-proxy", "cluster-dots.envoy.golden.yaml")))
	})
})
