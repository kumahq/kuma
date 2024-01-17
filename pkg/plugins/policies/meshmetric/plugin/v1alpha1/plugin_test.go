package v1alpha1

import (
	"path/filepath"
	"strings"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	"github.com/kumahq/kuma/pkg/core/xds"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/pkg/test/xds/builders"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

func getResource(resourceSet *core_xds.ResourceSet, typ envoy_resource.Type) []byte {
	resources, err := resourceSet.ListOf(typ).ToDeltaDiscoveryResponse()
	Expect(err).ToNot(HaveOccurred())
	actual, err := util_proto.ToYAML(resources)
	Expect(err).ToNot(HaveOccurred())

	return actual
}

func testCaseName(ginkgo FullGinkgoTInterface) string {
	nameSplit := strings.Split(ginkgo.Name(), " ")
	return nameSplit[len(nameSplit)-1]
}

var _ = Describe("MeshMetric", func() {
	type testCase struct {
		proxy   *core_xds.Proxy
		context xds_context.Context
	}

	DescribeTable("Apply to sidecar Dataplane", func(given testCase) {
		resources := core_xds.NewResourceSet()
		plugin := NewPlugin().(core_plugins.PolicyPlugin)

		Expect(plugin.Apply(resources, given.context, given.proxy)).To(Succeed())

		name := testCaseName(GinkgoT())

		Expect(getResource(resources, envoy_resource.ListenerType)).To(matchers.MatchGoldenYAML(filepath.Join("testdata", name+".listeners.golden.yaml")))
		Expect(getResource(resources, envoy_resource.ClusterType)).To(matchers.MatchGoldenYAML(filepath.Join("testdata", name+".clusters.golden.yaml")))
	},
		Entry("basic", testCase{
			proxy: xds_builders.Proxy().
				WithDataplane(samples.DataplaneBackendBuilder()).
				WithMetadata(&xds.DataplaneMetadata{MetricsSocketPath: "/tmp/kuma-metrics-backend-default.sock"}).
				WithPolicies(xds_builders.MatchedPolicies().
					WithSingleItemPolicy(api.MeshMetricType, core_rules.SingleItemRules{
						Rules: []*core_rules.Rule{
							{
								Subset: []core_rules.Tag{},
								Conf: api.Conf{
									Sidecar: &api.Sidecar{
										Regex:         pointer.To("http.*"),
										IncludeUnused: pointer.To(false),
									},
									Applications: &[]api.Application{
										{
											Path: pointer.To("/metrics"),
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
					}),
				).
				Build(),
		}),
		Entry("provided_tls", testCase{
			context: *xds_builders.Context().WithMesh(samples.MeshMTLSBuilder()).Build(),
			proxy: xds_builders.Proxy().
				WithDataplane(samples.DataplaneBackendBuilder()).
				WithMetadata(&xds.DataplaneMetadata{
					MetricsSocketPath: "/tmp/kuma-metrics-backend-default.sock",
					MetricsCertPath:   "/path/cert",
					MetricsKeyPath:    "/path/key",
				}).
				WithPolicies(xds_builders.MatchedPolicies().
					WithSingleItemPolicy(api.MeshMetricType, core_rules.SingleItemRules{
						Rules: []*core_rules.Rule{
							{
								Subset: []core_rules.Tag{},
								Conf: api.Conf{
									Backends: &[]api.Backend{
										{
											Type: api.PrometheusBackendType,
											Prometheus: &api.PrometheusBackend{
												Path: "/metrics",
												Port: 5670,
												Tls: &api.PrometheusTls{
													Mode: api.ProvidedTLS,
												},
											},
										},
									},
								},
							},
						},
					}),
				).
				Build(),
		}),
	)
})
