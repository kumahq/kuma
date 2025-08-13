package v1alpha1_test

import (
	"path/filepath"
	"time"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"

	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/subsetutils"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/plugin/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/pkg/test/xds/builders"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_yaml "github.com/kumahq/kuma/pkg/util/yaml"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

var _ = Describe("MeshMetric", func() {
	type testCase struct {
		proxy   *core_xds.Proxy
		context xds_context.Context
	}

	DescribeTable("Apply to sidecar Dataplane", func(given testCase) {
		resources := core_xds.NewResourceSet()
		plugin := v1alpha1.NewPlugin().(core_plugins.PolicyPlugin)
		name := CurrentSpecReport().LeafNodeText

		Expect(plugin.Apply(resources, given.context, given.proxy)).To(Succeed())

		resource, err := util_yaml.GetResourcesToYaml(resources, envoy_resource.ListenerType)
		Expect(err).ToNot(HaveOccurred())
		Expect(resource).To(matchers.MatchGoldenYAML(filepath.Join("testdata", name+".listeners.golden.yaml")))
		resource, err = util_yaml.GetResourcesToYaml(resources, envoy_resource.ClusterType)
		Expect(err).ToNot(HaveOccurred())
		Expect(resource).To(matchers.MatchGoldenYAML(filepath.Join("testdata", name+".clusters.golden.yaml")))
	},
		Entry("default", testCase{
			proxy: xds_builders.Proxy().
				WithID(*core_xds.BuildProxyId("default", "backend")).
				WithMetadata(&core_xds.DataplaneMetadata{WorkDir: "/tmp"}).
				WithDataplane(samples.DataplaneBackendBuilder()).
				WithPolicies(xds_builders.MatchedPolicies().
					WithSingleItemPolicy(api.MeshMetricType, core_rules.SingleItemRules{
						Rules: []*core_rules.Rule{
							{
								Subset: []subsetutils.Tag{},
								Conf: api.Conf{
									Applications: &[]api.Application{
										{
											Name: pointer.To("test-app"),
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
					}),
				).
				Build(),
		}),
		Entry("basic", testCase{
			proxy: xds_builders.Proxy().
				WithID(*core_xds.BuildProxyId("default", "backend")).
				WithDataplane(samples.DataplaneBackendBuilder()).
				WithMetadata(&core_xds.DataplaneMetadata{WorkDir: "/tmp"}).
				WithPolicies(xds_builders.MatchedPolicies().
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
					}),
				).
				Build(),
		}),
		Entry("multiple_prometheus", testCase{
			proxy: xds_builders.Proxy().
				WithID(*core_xds.BuildProxyId("default", "backend")).
				WithDataplane(samples.DataplaneBackendBuilder()).
				WithMetadata(&core_xds.DataplaneMetadata{WorkDir: "/tmp"}).
				WithPolicies(xds_builders.MatchedPolicies().
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
												ClientId: pointer.To("first-backend"),
												Path:     "/metrics",
												Port:     5670,
											},
										},
										{
											Type: api.PrometheusBackendType,
											Prometheus: &api.PrometheusBackend{
												ClientId: pointer.To("second-backend"),
												Path:     "/metrics",
												Port:     5671,
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
		Entry("openTelemetry", testCase{
			proxy: xds_builders.Proxy().
				WithID(*core_xds.BuildProxyId("default", "backend")).
				WithDataplane(samples.DataplaneBackendBuilder()).
				WithMetadata(&core_xds.DataplaneMetadata{WorkDir: "/tmp"}).
				WithPolicies(xds_builders.MatchedPolicies().
					WithSingleItemPolicy(api.MeshMetricType, core_rules.SingleItemRules{
						Rules: []*core_rules.Rule{
							{
								Subset: []subsetutils.Tag{},
								Conf: api.Conf{
									Applications: &[]api.Application{
										{
											Path: "/metrics",
											Port: 8080,
										},
									},
									Backends: &[]api.Backend{
										{
											Type: api.OpenTelemetryBackendType,
											OpenTelemetry: &api.OpenTelemetryBackend{
												Endpoint:        "otel-collector.observability.svc:4317",
												RefreshInterval: &k8s.Duration{Duration: 10 * time.Second},
											},
										},
									},
								},
							},
						},
					})).
				Build(),
		}),
		Entry("provided_tls", testCase{
			context: *xds_builders.Context().WithMeshBuilder(samples.MeshMTLSBuilder()).Build(),
			proxy: xds_builders.Proxy().
				WithID(*core_xds.BuildProxyId("default", "backend")).
				WithDataplane(samples.DataplaneBackendBuilder()).
				WithMetadata(&core_xds.DataplaneMetadata{
					WorkDir:         "/tmp",
					MetricsCertPath: "/path/cert",
					MetricsKeyPath:  "/path/key",
				}).
				WithPolicies(xds_builders.MatchedPolicies().
					WithSingleItemPolicy(api.MeshMetricType, core_rules.SingleItemRules{
						Rules: []*core_rules.Rule{
							{
								Subset: []subsetutils.Tag{},
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
		Entry("otel_and_prometheus", testCase{
			proxy: xds_builders.Proxy().
				WithID(*core_xds.BuildProxyId("default", "backend")).
				WithDataplane(samples.DataplaneBackendBuilder()).
				WithMetadata(&core_xds.DataplaneMetadata{
					WorkDir: "/tmp",
					Features: map[string]bool{
						xds_types.FeatureUnifiedResourceNaming: true,
					},
				}).
				WithPolicies(xds_builders.MatchedPolicies().
					WithSingleItemPolicy(api.MeshMetricType, core_rules.SingleItemRules{
						Rules: []*core_rules.Rule{
							{
								Subset: []subsetutils.Tag{},
								Origin: []core_model.ResourceMeta{
									&test_model.ResourceMeta{
										Mesh: "default",
										Name: "meshmetric1",
									},
								},
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
										{
											Type: api.OpenTelemetryBackendType,
											OpenTelemetry: &api.OpenTelemetryBackend{
												Endpoint: "otel-collector.observability.svc:4317",
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
		Entry("multiple_otel", testCase{
			proxy: xds_builders.Proxy().
				WithDataplane(samples.DataplaneBackendBuilder()).
				WithMetadata(&core_xds.DataplaneMetadata{WorkDir: "/tmp"}).
				WithPolicies(xds_builders.MatchedPolicies().
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
											Type: api.OpenTelemetryBackendType,
											OpenTelemetry: &api.OpenTelemetryBackend{
												Endpoint: "otel-collector.observability.svc:4317",
											},
										},
										{
											Type: api.OpenTelemetryBackendType,
											OpenTelemetry: &api.OpenTelemetryBackend{
												Endpoint: "second-collector.observability.svc:4317",
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
