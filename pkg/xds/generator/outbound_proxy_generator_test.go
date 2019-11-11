package generator_test

import (
	"io/ioutil"
	"path/filepath"

	"github.com/Kong/kuma/pkg/core/logs"

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

	plainCtx := xds_context.Context{
		ControlPlane: &xds_context.ControlPlaneContext{},
		Mesh: xds_context.MeshContext{
			TlsEnabled: false,
		},
	}

	mtlsCtx := xds_context.Context{
		ControlPlane: &xds_context.ControlPlaneContext{
			SdsLocation: "kuma-system:5677",
			SdsTlsCert:  []byte("12345"),
		},
		Mesh: xds_context.MeshContext{
			TlsEnabled: true,
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
				Id: model.ProxyId{Name: "side-car", Namespace: "default", Mesh: "default"},
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
				Logs:     logs.NewMatchedLogs(),
				Metadata: &model.DataplaneMetadata{},
			}

			// when
			rs, err := gen.Generate(given.ctx, proxy)

			// then
			Expect(err).ToNot(HaveOccurred())

			// then
			resp := generator.ResourceList(rs).ToDeltaDiscoveryResponse()
			actual, err := util_proto.ToYAML(resp)
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
		Entry("03. transparent_proxying=false, mtls=false, outbound=1", testCase{
			ctx: plainCtx,
			dataplane: `
            networking:
              outbound:
              - interface: :18080
                service: backend
`,
			expected: "03.envoy.golden.yaml",
		}),
		Entry("04. transparent_proxying=true, mtls=false, outbound=1", testCase{
			ctx: mtlsCtx,
			dataplane: `
            networking:
              outbound:
              - interface: :18080
                service: backend
              transparentProxying:
                redirectPort: 15001
`,
			expected: "04.envoy.golden.yaml",
		}),
		Entry("05. transparent_proxying=false, mtls=true, outbound=1", testCase{
			ctx: plainCtx,
			dataplane: `
            networking:
              outbound:
              - interface: :18080
                service: backend
`,
			expected: "05.envoy.golden.yaml",
		}),
		Entry("06. transparent_proxying=true, mtls=true, outbound=1", testCase{
			ctx: mtlsCtx,
			dataplane: `
            networking:
              outbound:
              - interface: :18080
                service: backend
              transparentProxying:
                redirectPort: 15001
`,
			expected: "06.envoy.golden.yaml",
		}),
		Entry("07. transparent_proxying=false, mtls=false, outbound=2", testCase{
			ctx: plainCtx,
			dataplane: `
            networking:
              outbound:
              - interface: :18080
                service: backend
              - interface: :54321
                service: db
`,
			expected: "07.envoy.golden.yaml",
		}),
		Entry("08. transparent_proxying=true, mtls=true, outbound=2", testCase{
			ctx: mtlsCtx,
			dataplane: `
            networking:
              outbound:
              - interface: :18080
                service: backend
              - interface: :54321
                service: db
              transparentProxying:
                redirectPort: 15001
`,
			expected: "08.envoy.golden.yaml",
		}),
	)
})
