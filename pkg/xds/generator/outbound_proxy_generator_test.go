package generator_test

import (
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	gtypes "github.com/onsi/gomega/types"

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
						Version: "1",
					},
					Spec: dataplane,
				},
				TrafficRoutes: model.RouteMap{
					"api-http": &mesh_core.TrafficRouteResource{
						Spec: mesh_proto.TrafficRoute{
							Conf: []*mesh_proto.TrafficRoute_WeightedDestination{{
								Weight:      100,
								Destination: mesh_proto.MatchService("api-http"),
							}},
						},
					},
					"api-tcp": &mesh_core.TrafficRouteResource{
						Spec: mesh_proto.TrafficRoute{
							Conf: []*mesh_proto.TrafficRoute_WeightedDestination{{
								Weight:      100,
								Destination: mesh_proto.MatchService("api-tcp"),
							}},
						},
					},
					"backend": &mesh_core.TrafficRouteResource{
						Spec: mesh_proto.TrafficRoute{
							Conf: []*mesh_proto.TrafficRoute_WeightedDestination{{
								Weight:      100,
								Destination: mesh_proto.MatchService("backend"),
							}},
						},
					},
					"db": &mesh_core.TrafficRouteResource{
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
						{Target: "192.168.0.4", Port: 8084, Tags: map[string]string{"service": "api-http", "protocol": "http", "region": "us"}},
						{Target: "192.168.0.5", Port: 8085, Tags: map[string]string{"service": "api-http", "protocol": "http", "region": "eu"}},
					},
					"api-tcp": []model.Endpoint{ // notice that not every endpoint has a `protocol: http` tag
						{Target: "192.168.0.6", Port: 8086, Tags: map[string]string{"service": "api-tcp", "protocol": "http", "region": "us"}},
						{Target: "192.168.0.7", Port: 8087, Tags: map[string]string{"service": "api-tcp", "region": "eu"}},
					},
					"backend": []model.Endpoint{ // notice that not every endpoint has a tag `protocol: http`
						{Target: "192.168.0.1", Port: 8081, Tags: map[string]string{"service": "backend", "region": "us"}},
						{Target: "192.168.0.2", Port: 8082},
					},
					"db": []model.Endpoint{
						{Target: "192.168.0.3", Port: 5432, Tags: map[string]string{"service": "db", "role": "master"}},
					},
				},
				Logs: model.LogMap{
					"api-http": &mesh_proto.LoggingBackend{
						Name: "file",
						Type: &mesh_proto.LoggingBackend_File_{
							File: &mesh_proto.LoggingBackend_File{
								Path: "/var/log",
							},
						},
					},
					"api-tcp": &mesh_proto.LoggingBackend{
						Name: "elk",
						Type: &mesh_proto.LoggingBackend_Tcp_{
							Tcp: &mesh_proto.LoggingBackend_Tcp{
								Address: "logstash:1234",
							},
						},
					},
				},
				Metadata: &model.DataplaneMetadata{},
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
		Entry("4. transparent_proxying=true, mtls=true, outbound=4", testCase{
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
				"backend.kuma-system": &mesh_core.TrafficRouteResource{
					Spec: mesh_proto.TrafficRoute{
						Conf: []*mesh_proto.TrafficRoute_WeightedDestination{{
							Weight:      100,
							Destination: mesh_proto.MatchService("backend.kuma-system"),
						}},
					},
				},
				"db.kuma-system": &mesh_core.TrafficRouteResource{
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
					{Target: "192.168.0.1", Port: 8082},
				},
				"db.kuma-system": []model.Endpoint{
					{Target: "192.168.0.2", Port: 5432, Tags: map[string]string{"service": "db", "role": "master"}},
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

	Describe("fail when a user-defined configuration (Dataplane, TrafficRoute, etc) is not valid", func() {

		type testCase struct {
			ctx                xds_context.Context
			dataplane          string
			chaos              func(*model.Proxy)
			expectedErrMatcher gtypes.GomegaMatcher
		}

		noExtraChaos := func(*model.Proxy) {}

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
							Version: "1",
						},
						Spec: dataplane,
					},
					TrafficRoutes: model.RouteMap{
						"backend": &mesh_core.TrafficRouteResource{
							Spec: mesh_proto.TrafficRoute{
								Conf: []*mesh_proto.TrafficRoute_WeightedDestination{{
									Weight:      100,
									Destination: mesh_proto.MatchService("backend"),
								}},
							},
						},
						"db": &mesh_core.TrafficRouteResource{
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
						"backend": []model.Endpoint{
							{Target: "192.168.0.1", Port: 8081, Tags: map[string]string{"service": "backend", "region": "us"}},
							{Target: "192.168.0.2", Port: 8082},
						},
						"db": []model.Endpoint{
							{Target: "192.168.0.3", Port: 5432, Tags: map[string]string{"service": "db", "role": "master"}},
						},
					},
					Metadata: &model.DataplaneMetadata{},
				}

				By("introducing an error into configuration")
				given.chaos(proxy)

				// when
				rs, err := gen.Generate(given.ctx, proxy)

				// then
				Expect(err).To(HaveOccurred())
				// and
				Expect(err.Error()).To(given.expectedErrMatcher)
				// and
				Expect(rs).To(BeNil())
			},
			Entry("dataplane with an invalid outbound interface", testCase{
				ctx: plainCtx,
				dataplane: `
                networking:
                  outbound:
                  - interface: 127:not-a-port
                    service: backend
`,
				chaos:              noExtraChaos,
				expectedErrMatcher: HavePrefix(`invalid DATAPLANE_IP in "127:not-a-port": "127" is not a valid IP address`),
			}),
			Entry("dataplane with an outbound interface that has no route", testCase{
				ctx: plainCtx,
				dataplane: `
                networking:
                  outbound:
                  - port: 18080
                    service: backend
`,
				chaos: func(proxy *model.Proxy) {
					// simulate missing route
					proxy.TrafficRoutes = nil
				},
				expectedErrMatcher: Equal(`dataplane.networking.outbound[0]{service="backend"}: has no TrafficRoute`),
			}),
			Entry("dataplane with an outbound interface that has a route without destination", testCase{
				ctx: plainCtx,
				dataplane: `
                networking:
                  outbound:
                  - port: 18080
                    service: backend
`,
				chaos: func(proxy *model.Proxy) {
					// simulate missing route
					proxy.TrafficRoutes = model.RouteMap{
						"backend": &mesh_core.TrafficRouteResource{
							Meta: &test_model.ResourceMeta{
								Name: "route-without-destination",
							},
							Spec: mesh_proto.TrafficRoute{
								Sources:      []*mesh_proto.Selector{{Match: mesh_proto.MatchAnyService()}},
								Destinations: []*mesh_proto.Selector{{Match: mesh_proto.MatchAnyService()}},
								Conf: []*mesh_proto.TrafficRoute_WeightedDestination{{
									Weight: 100, Destination: map[string]string{"not-a-service": "value"},
								}},
							},
						},
					}
				},
				expectedErrMatcher: Equal(`trafficroute{name="route-without-destination"}.conf[0].destination: mandatory tag "service" is missing: map[not-a-service:value]`),
			}),
		)
	})
})
