package generator_test

import (
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	model "github.com/Kong/kuma/pkg/core/xds"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
	xds_context "github.com/Kong/kuma/pkg/xds/context"
	"github.com/Kong/kuma/pkg/xds/generator"

	test_model "github.com/Kong/kuma/pkg/test/resources/model"
)

var _ = Describe("OutboundProxyGenerator", func() {

	meta := &test_model.ResourceMeta{
		Mesh: "mesh1",
		Name: "mesh1",
	}
	plainCtx := xds_context.Context{
		ControlPlane: &xds_context.ControlPlaneContext{},
		Mesh: xds_context.MeshContext{
			Resource: &mesh_core.MeshResource{
				Meta: meta,
			},
		},
	}

	mtlsCtx := xds_context.Context{
		ControlPlane: &xds_context.ControlPlaneContext{
			SdsLocation: "kuma-system:5677",
			SdsTlsCert:  []byte("12345"),
		},
		Mesh: xds_context.MeshContext{
			Resource: &mesh_core.MeshResource{
				Spec: mesh_proto.Mesh{
					Mtls: &mesh_proto.Mesh_Mtls{
						EnabledBackend: "builtin",
						Backends: []*mesh_proto.CertificateAuthorityBackend{
							{
								Name: "builtin",
								Type: "builtin",
							},
						},
					},
				},
				Meta: meta,
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

			dataplane := mesh_proto.Dataplane{}
			Expect(util_proto.FromYAML([]byte(given.dataplane), &dataplane)).To(Succeed())

			proxy := &model.Proxy{
				Id: model.ProxyId{Name: "side-car", Mesh: "default"},
				Dataplane: &mesh_core.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Name:    "dp-1",
						Mesh:    given.ctx.Mesh.Resource.Meta.GetName(),
						Version: "1",
					},
					Spec: dataplane,
				},
				TrafficRoutes: model.RouteMap{
					mesh_proto.OutboundInterface{
						DataplaneIP:   "127.0.0.1",
						DataplanePort: 40001,
					}: &mesh_core.TrafficRouteResource{
						Spec: mesh_proto.TrafficRoute{
							Conf: []*mesh_proto.TrafficRoute_WeightedDestination{{
								Weight:      100,
								Destination: mesh_proto.MatchService("api-http"),
							}},
						},
					},
					mesh_proto.OutboundInterface{
						DataplaneIP:   "127.0.0.1",
						DataplanePort: 40002,
					}: &mesh_core.TrafficRouteResource{
						Spec: mesh_proto.TrafficRoute{
							Conf: []*mesh_proto.TrafficRoute_WeightedDestination{{
								Weight:      100,
								Destination: mesh_proto.MatchService("api-tcp"),
							}},
						},
					},
					mesh_proto.OutboundInterface{
						DataplaneIP:   "127.0.0.1",
						DataplanePort: 18080,
					}: &mesh_core.TrafficRouteResource{
						Spec: mesh_proto.TrafficRoute{
							Conf: []*mesh_proto.TrafficRoute_WeightedDestination{{
								Weight:      100,
								Destination: mesh_proto.MatchService("backend"),
							}},
						},
					},
					mesh_proto.OutboundInterface{
						DataplaneIP:   "127.0.0.1",
						DataplanePort: 54321,
					}: &mesh_core.TrafficRouteResource{
						Spec: mesh_proto.TrafficRoute{
							Conf: []*mesh_proto.TrafficRoute_WeightedDestination{{
								Weight:      10,
								Destination: mesh_proto.TagSelector{"service": "db", "role": "master"},
							}, {
								Weight:      90,
								Destination: mesh_proto.TagSelector{"service": "db", "role": "replica"},
							}, {
								Weight:      0, // should be excluded from Envoy configuration
								Destination: mesh_proto.TagSelector{"service": "db", "role": "canary"},
							}},
						},
					},
				},
				OutboundSelectors: model.DestinationMap{
					"api-http": model.TagSelectorSet{
						{"service": "api-http"},
					},
					"api-tcp": model.TagSelectorSet{
						{"service": "api-tcp"},
					},
					"backend": model.TagSelectorSet{
						{"service": "backend"},
					},
					"db": model.TagSelectorSet{
						{"service": "db", "role": "master"},
						{"service": "db", "role": "replica"},
						{"service": "db", "role": "canary"},
					},
				},
				OutboundTargets: model.EndpointMap{
					"api-http": []model.Endpoint{ // notice that all endpoints have tag `protocol: http`
						{
							Target: "192.168.0.4",
							Port:   8084,
							Tags:   map[string]string{"service": "api-http", "protocol": "http", "region": "us"},
							Weight: 1,
						},
						{
							Target: "192.168.0.5",
							Port:   8085,
							Tags:   map[string]string{"service": "api-http", "protocol": "http", "region": "eu"},
							Weight: 1,
						},
					},
					"api-tcp": []model.Endpoint{ // notice that not every endpoint has a `protocol: http` tag
						{
							Target: "192.168.0.6",
							Port:   8086,
							Tags:   map[string]string{"service": "api-tcp", "protocol": "http", "region": "us"},
							Weight: 1,
						},
						{
							Target: "192.168.0.7",
							Port:   8087,
							Tags:   map[string]string{"service": "api-tcp", "region": "eu"},
							Weight: 1,
						},
					},
					"backend": []model.Endpoint{ // notice that not every endpoint has a tag `protocol: http`
						{
							Target: "192.168.0.1",
							Port:   8081,
							Tags:   map[string]string{"service": "backend", "region": "us"},
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
							Tags:   map[string]string{"service": "db", "role": "master"},
							Weight: 1,
						},
					},
				},
				Logs: model.LogMap{
					"api-http": &mesh_proto.LoggingBackend{
						Name: "file",
						Type: mesh_proto.LoggingFileType,
						Conf: util_proto.MustToStruct(&mesh_proto.FileLoggingBackendConfig{
							Path: "/var/log",
						}),
					},
					"api-tcp": &mesh_proto.LoggingBackend{
						Name: "elk",
						Type: mesh_proto.LoggingTcpType,
						Conf: util_proto.MustToStruct(&mesh_proto.TcpLoggingBackendConfig{
							Address: "logstash:1234",
						}),
					},
				},
				Metadata: &model.DataplaneMetadata{},
				CircuitBreakers: model.CircuitBreakerMap{
					"api-http": &mesh_core.CircuitBreakerResource{
						Spec: mesh_proto.CircuitBreaker{
							Conf: &mesh_proto.CircuitBreaker_Conf{
								Detectors: &mesh_proto.CircuitBreaker_Conf_Detectors{
									TotalErrors: &mesh_proto.CircuitBreaker_Conf_Detectors_Errors{},
								},
							},
						},
					},
				},
			}

			// when
			rs, err := gen.Generate(given.ctx, proxy)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			resp, err := model.ResourceList(rs).ToDeltaDiscoveryResponse()
			// then
			Expect(err).ToNot(HaveOccurred())
			// when
			actual, err := util_proto.ToYAML(resp)
			// then
			Expect(err).ToNot(HaveOccurred())

			expected, err := ioutil.ReadFile(filepath.Join("testdata", "outbound-proxy", given.expected))
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(expected))
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
                  service: gateway
              outbound:
              - port: 18080
                service: backend
              - port: 54321
                service: db
              - port: 40001
                service: api-http
              - port: 40002
                service: api-tcp
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
                  service: web
              outbound:
              - port: 18080
                service: backend
              - port: 54321
                service: db
              - port: 40001
                service: api-http
              - port: 40002
                service: api-tcp
              transparentProxying:
                redirectPort: 15001
`,
			expected: "04.envoy.golden.yaml",
		}),
	)

	It("Add sanitized alternative cluster name for stats", func() {
		// setup
		gen := &generator.OutboundProxyGenerator{}
		dp := `
        networking:
          outbound:
          - port: 18080
            service: backend.kuma-system
          - port: 54321
            service: db.kuma-system`

		dataplane := mesh_proto.Dataplane{}
		Expect(util_proto.FromYAML([]byte(dp), &dataplane)).To(Succeed())

		proxy := &model.Proxy{
			Id: model.ProxyId{Name: "side-car", Mesh: "default"},
			Dataplane: &mesh_core.DataplaneResource{
				Meta: &test_model.ResourceMeta{
					Version: "1",
				},
				Spec: dataplane,
			},
			TrafficRoutes: model.RouteMap{
				mesh_proto.OutboundInterface{
					DataplaneIP:   "127.0.0.1",
					DataplanePort: 18080,
				}: &mesh_core.TrafficRouteResource{
					Spec: mesh_proto.TrafficRoute{
						Conf: []*mesh_proto.TrafficRoute_WeightedDestination{{
							Weight:      100,
							Destination: mesh_proto.MatchService("backend.kuma-system"),
						}},
					},
				},
				mesh_proto.OutboundInterface{
					DataplaneIP:   "127.0.0.1",
					DataplanePort: 54321,
				}: &mesh_core.TrafficRouteResource{
					Spec: mesh_proto.TrafficRoute{
						Conf: []*mesh_proto.TrafficRoute_WeightedDestination{{
							Weight:      100,
							Destination: mesh_proto.TagSelector{"service": "db", "version": "3.2.0"},
						},
						}},
				},
			},
			OutboundSelectors: model.DestinationMap{
				"backend.kuma-system": model.TagSelectorSet{
					{"service": "backend.kuma-system"},
				},
				"db.kuma-system": model.TagSelectorSet{
					{"service": "db", "version": "3.2.0"},
				},
			},
			OutboundTargets: model.EndpointMap{
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
						Tags:   map[string]string{"service": "db", "role": "master"},
						Weight: 1,
					},
				},
			},
			Metadata: &model.DataplaneMetadata{},
		}

		// when
		rs, err := gen.Generate(plainCtx, proxy)

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		resp, err := model.ResourceList(rs).ToDeltaDiscoveryResponse()
		// then
		Expect(err).ToNot(HaveOccurred())
		// when
		actual, err := util_proto.ToYAML(resp)
		// then
		Expect(err).ToNot(HaveOccurred())

		expected, err := ioutil.ReadFile(filepath.Join("testdata", "outbound-proxy", "cluster-dots.envoy.golden.yaml"))
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(MatchYAML(expected))
	})
})
