package generator_test

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	model "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/subsetutils"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/plugin/v1alpha1"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	xds_builders "github.com/kumahq/kuma/pkg/test/xds/builders"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ = Describe("DNSGenerator", func() {
	var ctx xds_context.Context
	var proxy *model.Proxy
	var gen *generator.DNSGenerator
	var dataplane mesh_proto.Dataplane

	BeforeEach(func() {
		ctx = xds_context.Context{
			ControlPlane: &xds_context.ControlPlaneContext{},
			Mesh: xds_context.MeshContext{
				Resource: &core_mesh.MeshResource{
					Meta: &test_model.ResourceMeta{
						Name: "default",
					},
					Spec: &mesh_proto.Mesh{},
				},
				VIPDomains: []xds_types.VIPDomains{
					{Address: "240.0.0.1", Domains: []string{"httpbin.mesh"}},
					{Address: "240.0.0.0", Domains: []string{"backend.test-ns.svc.8080.mesh", "backend_test-ns_svc_8080.mesh"}},
					{Address: "2001:db8::ff00:42:8329", Domains: []string{"frontend.test-ns.svc.8080.mesh", "frontend_test-ns_svc_8080.mesh"}}, // this is ignored because there is no outbounds for it
				},
			},
		}
		proxy = &model.Proxy{
			Id: *model.BuildProxyId("", "side-car"),
			Dataplane: &core_mesh.DataplaneResource{
				Meta: &test_model.ResourceMeta{
					Version: "1",
				},
				Spec: nil,
			},
			APIVersion: envoy_common.APIV3,
			Routing:    model.Routing{},
			Metadata: &model.DataplaneMetadata{
				DNSPort:  53001,
				Version:  &mesh_proto.Version{Envoy: &mesh_proto.EnvoyVersion{Version: "1.20.0"}},
				Features: nil,
			},
			InternalAddresses: DummyInternalAddresses,
		}
		gen = &generator.DNSGenerator{}
		dataplane = mesh_proto.Dataplane{}
	})

	type testCase struct {
		dataplaneFile string
		expected      string
		features      map[string]bool
		hasMeshMetric bool
	}

	DescribeTable("Generate Envoy xDS resources",
		func(given testCase) {
			// setup
			dpBytes, err := os.ReadFile(filepath.Join("testdata", "dns", given.dataplaneFile))
			Expect(err).ToNot(HaveOccurred())
			Expect(util_proto.FromYAML(dpBytes, &dataplane)).To(Succeed())
			proxy.Dataplane.Spec = &dataplane
			proxy.Metadata.Features = given.features

			for _, dppOutbound := range dataplane.GetNetworking().GetOutbound() {
				proxy.Outbounds = append(proxy.Outbounds, &xds_types.Outbound{LegacyOutbound: dppOutbound})
			}

			// when
			rs := model.NewResourceSet()
			resources, err := gen.Generate(context.Background(), rs, ctx, proxy)

			// then
			Expect(err).ToNot(HaveOccurred())
			rs.AddSet(resources)

			// and output matches golden files
			resp, err := rs.List().ToDeltaDiscoveryResponse()
			Expect(err).ToNot(HaveOccurred())
			actual, err := util_proto.ToYAML(resp)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchGoldenYAML(filepath.Join("testdata", "dns", given.expected)))
		},
		Entry("01. DNS enabled", testCase{
			dataplaneFile: "1-dataplane.input.yaml",
			expected:      "1-envoy-config.golden.yaml",
		}),
		Entry("02. DNS disabled", testCase{
			dataplaneFile: "2-dataplane.input.yaml",
			expected:      "2-envoy-config.golden.yaml",
		}),
		Entry("03. DNS enabled no ipv6", testCase{
			dataplaneFile: "3-dataplane.input.yaml",
			expected:      "3-envoy-config.golden.yaml",
		}),
		Entry("04. DNS using proxy map", testCase{
			features: map[string]bool{
				"feature-embedded-dns": true,
			},
			dataplaneFile: "4-dataplane.input.yaml",
			expected:      "4-envoy-config.golden.yaml",
		}),
	)

	DescribeTable("Generate Envoy xDS resources based on embedded DNS and mesh metric",
		func(given testCase) {
			// setup
			dpBytes, err := os.ReadFile(filepath.Join("testdata", "dns", given.dataplaneFile))
			Expect(err).ToNot(HaveOccurred())
			Expect(util_proto.FromYAML(dpBytes, &dataplane)).To(Succeed())
			proxy.Dataplane.Spec = &dataplane
			proxy.Metadata.Features = given.features

			for _, dppOutbound := range dataplane.GetNetworking().GetOutbound() {
				proxy.Outbounds = append(proxy.Outbounds, &xds_types.Outbound{LegacyOutbound: dppOutbound})
			}

			// when
			rs := model.NewResourceSet()
			resources, err := gen.Generate(context.Background(), rs, ctx, proxy)

			// then
			Expect(err).ToNot(HaveOccurred())
			rs.AddSet(resources)

			if given.hasMeshMetric {
				// when
				proxy.Policies = *xds_builders.MatchedPolicies().
					WithSingleItemPolicy(api.MeshMetricType, core_rules.SingleItemRules{
						Rules: []*core_rules.Rule{
							{
								Subset: []subsetutils.Tag{},
								Conf: api.Conf{
									Sidecar: &api.Sidecar{
										IncludeUnused: pointer.To(false),
									},
									Applications: &[]api.Application{
										{
											Path: "/metrics",
											Port: 8080,
										},
									},
									Backends: &[]api.Backend{
										{
											Type: api.PrometheusBackendType,
											Prometheus: &api.PrometheusBackend{
												Path: "/metrics",
												Port: 5670,
											},
										},
									},
								},
							},
						},
					}).Build()
				plugin := v1alpha1.NewPlugin().(core_plugins.PolicyPlugin)

				// then
				Expect(plugin.Apply(rs, ctx, proxy)).To(Succeed())
			}

			// and output matches golden files
			resp, err := rs.List().ToDeltaDiscoveryResponse()
			Expect(err).ToNot(HaveOccurred())
			actual, err := util_proto.ToYAML(resp)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchGoldenYAML(filepath.Join("testdata", "dns", given.expected)))
		},
		Entry("05. Embedded DNS enabled without mesh metric", testCase{ // maxDirectResponseBodySizeBytes: 492
			features: map[string]bool{
				"feature-embedded-dns": true,
			},
			dataplaneFile: "5-dataplane.input.yaml",
			expected:      "5-envoy-config.golden.yaml",
			hasMeshMetric: false,
		}),
		Entry("06. Embedded DNS disabled with mesh metric", testCase{ // maxDirectResponseBodySizeBytes: 505
			features: map[string]bool{
				"feature-embedded-dns": false,
			},
			dataplaneFile: "6-dataplane.input.yaml",
			expected:      "6-envoy-config.golden.yaml",
			hasMeshMetric: true,
		}),
		Entry("07. Embedded DNS enabled with mesh metric", testCase{ // maxDirectResponseBodySizeBytes: 505
			features: map[string]bool{
				"feature-embedded-dns": true,
			},
			dataplaneFile: "7-dataplane.input.yaml",
			expected:      "7-envoy-config.golden.yaml",
			hasMeshMetric: true,
		}),
	)
})
