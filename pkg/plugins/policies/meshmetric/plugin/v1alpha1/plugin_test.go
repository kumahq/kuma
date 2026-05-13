//nolint:staticcheck // SA1019 Test file: tests backward compatibility with deprecated core_rules.Rule
package v1alpha1_test

import (
	"path/filepath"
	"time"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/v2/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	motb_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshopentelemetrybackend/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/v2/pkg/core/xds/types"
	core_rules "github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules/subsetutils"
	api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshmetric/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/meshmetric/plugin/v1alpha1"
	k8s_metadata "github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/v2/pkg/test/matchers"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v2/pkg/test/resources/model"
	"github.com/kumahq/kuma/v2/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/v2/pkg/test/xds/builders"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
	util_yaml "github.com/kumahq/kuma/v2/pkg/util/yaml"
	xds_context "github.com/kumahq/kuma/v2/pkg/xds/context"
)

func workloadLabels() map[string]string {
	return map[string]string{
		k8s_metadata.KumaWorkload:   "backend",
		mesh_proto.ZoneTag:          "zone-1",
		mesh_proto.KubeNamespaceTag: "kuma-demo",
	}
}

func zoneEgressOnlyDataplane(name string) *builders.DataplaneBuilder {
	return builders.Dataplane().
		WithName(name).
		WithAddress("192.168.0.10").
		WithoutInbounds().
		With(func(d *core_mesh.DataplaneResource) {
			d.Spec.Networking.Listeners = []*mesh_proto.Dataplane_Networking_Listener{
				{
					Type:    mesh_proto.Dataplane_Networking_Listener_ZoneEgress,
					Address: "192.168.0.10",
					Port:    10002,
					Name:    "ze-port",
				},
			}
		})
}

func zoneIngressOnlyDataplane(name string) *builders.DataplaneBuilder {
	return builders.Dataplane().
		WithName(name).
		WithAddress("192.168.0.11").
		WithoutInbounds().
		With(func(d *core_mesh.DataplaneResource) {
			d.Spec.Networking.Listeners = []*mesh_proto.Dataplane_Networking_Listener{
				{
					Type:    mesh_proto.Dataplane_Networking_Listener_ZoneIngress,
					Address: "192.168.0.11",
					Port:    10001,
					Name:    "zi-port",
				},
			}
		})
}

func mixedInboundAndZoneEgressDataplane() *builders.DataplaneBuilder {
	return samples.DataplaneBackendBuilder().
		With(func(d *core_mesh.DataplaneResource) {
			d.Spec.Networking.Listeners = []*mesh_proto.Dataplane_Networking_Listener{
				{
					Type:    mesh_proto.Dataplane_Networking_Listener_ZoneEgress,
					Address: "192.168.0.1",
					Port:    10002,
					Name:    "ze-port",
				},
			}
		})
}

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
			context: *xds_builders.Context().WithMeshBuilder(samples.MeshDefaultBuilder()).Build(),
			proxy: xds_builders.Proxy().
				WithID(*core_xds.BuildProxyId("default", "backend")).
				WithMetadata(&core_xds.DataplaneMetadata{WorkDir: "/tmp"}).
				WithDataplane(
					samples.DataplaneBackendBuilder().
						WithLabels(workloadLabels()),
				).
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
			context: *xds_builders.Context().WithMeshBuilder(samples.MeshDefaultBuilder()).Build(),
			proxy: xds_builders.Proxy().
				WithID(*core_xds.BuildProxyId("default", "backend")).
				WithDataplane(
					samples.DataplaneBackendBuilder().
						WithLabels(workloadLabels()),
				).
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
			context: *xds_builders.Context().WithMeshBuilder(samples.MeshDefaultBuilder()).Build(),
			proxy: xds_builders.Proxy().
				WithID(*core_xds.BuildProxyId("default", "backend")).
				WithDataplane(
					samples.DataplaneBackendBuilder().
						WithLabels(workloadLabels()),
				).
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
			context: *xds_builders.Context().WithMeshBuilder(samples.MeshDefaultBuilder()).Build(),
			proxy: xds_builders.Proxy().
				WithID(*core_xds.BuildProxyId("default", "backend")).
				WithDataplane(
					samples.DataplaneBackendBuilder().
						WithLabels(workloadLabels()),
				).
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
				WithDataplane(
					samples.DataplaneBackendBuilder().
						WithLabels(workloadLabels()),
				).
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
			context: *xds_builders.Context().WithMeshBuilder(
				samples.MeshDefaultBuilder().
					WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Exclusive),
			).Build(),
			proxy: xds_builders.Proxy().
				WithID(*core_xds.BuildProxyId("default", "backend")).
				WithDataplane(
					samples.DataplaneBackendBuilder().
						WithLabels(workloadLabels()),
				).
				WithMetadata(&core_xds.DataplaneMetadata{WorkDir: "/tmp"}).
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
		Entry("otel_and_prometheus_unified_naming", testCase{
			context: *xds_builders.Context().WithMeshBuilder(
				samples.MeshDefaultBuilder().
					WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Exclusive),
			).Build(),
			proxy: xds_builders.Proxy().
				WithID(*core_xds.BuildProxyId("default", "backend")).
				WithDataplane(
					samples.DataplaneBackendBuilder().
						WithLabels(workloadLabels()),
				).
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
			context: *xds_builders.Context().WithMeshBuilder(samples.MeshDefaultBuilder()).Build(),
			proxy: xds_builders.Proxy().
				WithDataplane(
					samples.DataplaneBackendBuilder().
						WithLabels(workloadLabels()),
				).
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
		Entry("zone_egress_only", testCase{
			context: *xds_builders.Context().WithMeshBuilder(samples.MeshDefaultBuilder()).Build(),
			proxy: xds_builders.Proxy().
				WithID(*core_xds.BuildProxyId("default", "zone-egress-1")).
				WithMetadata(&core_xds.DataplaneMetadata{WorkDir: "/tmp"}).
				WithDataplane(zoneEgressOnlyDataplane("zone-egress-1")).
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
		Entry("zone_ingress_only", testCase{
			context: *xds_builders.Context().WithMeshBuilder(samples.MeshDefaultBuilder()).Build(),
			proxy: xds_builders.Proxy().
				WithID(*core_xds.BuildProxyId("default", "zone-ingress-1")).
				WithMetadata(&core_xds.DataplaneMetadata{WorkDir: "/tmp"}).
				WithDataplane(zoneIngressOnlyDataplane("zone-ingress-1")).
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
		Entry("mixed_inbound_and_zone_egress", testCase{
			context: *xds_builders.Context().WithMeshBuilder(samples.MeshDefaultBuilder()).Build(),
			proxy: xds_builders.Proxy().
				WithID(*core_xds.BuildProxyId("default", "backend")).
				WithMetadata(&core_xds.DataplaneMetadata{WorkDir: "/tmp"}).
				WithDataplane(mixedInboundAndZoneEgressDataplane().WithLabels(workloadLabels())).
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
	)

	Describe("dynconf payload on zone-proxy DPPs", func() {
		buildProxy := func(dp *builders.DataplaneBuilder, name string) *core_xds.Proxy {
			return xds_builders.Proxy().
				WithID(*core_xds.BuildProxyId("default", name)).
				WithMetadata(&core_xds.DataplaneMetadata{WorkDir: "/tmp"}).
				WithDataplane(dp).
				WithPolicies(xds_builders.MatchedPolicies().
					WithSingleItemPolicy(api.MeshMetricType, core_rules.SingleItemRules{
						Rules: []*core_rules.Rule{
							{
								Subset: []subsetutils.Tag{},
								Conf: api.Conf{
									Applications: &[]api.Application{
										{Name: pointer.To("my-app"), Path: "/metrics", Port: 8080},
									},
									Backends: &[]api.Backend{
										{
											Type:       api.PrometheusBackendType,
											Prometheus: &api.PrometheusBackend{Path: "/metrics", Port: 5670},
										},
									},
								},
							},
						},
					}),
				).
				Build()
		}

		applyAndExtractDynconfBody := func(proxy *core_xds.Proxy) string {
			resources := core_xds.NewResourceSet()
			plug := v1alpha1.NewPlugin().(core_plugins.PolicyPlugin)
			ctx := *xds_builders.Context().WithMeshBuilder(samples.MeshDefaultBuilder()).Build()
			Expect(plug.Apply(resources, ctx, proxy)).To(Succeed())
			listeners, err := util_yaml.GetResourcesToYaml(resources, envoy_resource.ListenerType)
			Expect(err).ToNot(HaveOccurred())
			return string(listeners)
		}

		It("zone-egress-only: drops applications and uses dataplane name as service label", func() {
			body := applyAndExtractDynconfBody(buildProxy(zoneEgressOnlyDataplane("zone-egress-1"), "zone-egress-1"))
			// Applications cleared by sanitizeConfForProxy.
			Expect(body).ToNot(ContainSubstring("my-app"))
			Expect(body).ToNot(ContainSubstring(`"applications":[{`))
			// service label falls back to dataplane name (not "unknown").
			Expect(body).To(ContainSubstring(`"service":"zone-egress-1"`))
			Expect(body).ToNot(ContainSubstring(`"service":"unknown"`))
		})

		It("zone-ingress-only: uses dataplane name as service label", func() {
			body := applyAndExtractDynconfBody(buildProxy(zoneIngressOnlyDataplane("zone-ingress-1"), "zone-ingress-1"))
			Expect(body).To(ContainSubstring(`"service":"zone-ingress-1"`))
			Expect(body).ToNot(ContainSubstring(`"service":"unknown"`))
		})

		It("mixed inbound + zone-egress: preserves applications and uses inbound service tag", func() {
			dp := mixedInboundAndZoneEgressDataplane().WithLabels(map[string]string{
				mesh_proto.ZoneTag: "zone-1",
			})
			body := applyAndExtractDynconfBody(buildProxy(dp, "backend"))
			// Applications preserved because the DPP is not zone-proxy-only.
			Expect(body).To(ContainSubstring("my-app"))
			// service label resolves to the inbound's service tag, not "unknown" or the DP name.
			Expect(body).To(ContainSubstring(`"service":"backend"`))
		})
	})

	Describe("pipe mode (FeatureOtelViaKumaDp)", func() {
		const (
			workDir     = "/tmp"
			backendName = "otel-backend"
		)

		newMotb := func() *motb_api.MeshOpenTelemetryBackendResource {
			motb := motb_api.NewMeshOpenTelemetryBackendResource()
			motb.SetMeta(&test_model.ResourceMeta{
				Mesh: "default",
				Name: backendName,
			})
			motb.Spec.Endpoint = &motb_api.Endpoint{
				Address: new("collector.mesh"),
				Port:    new(int32(4317)),
			}
			motb.Spec.Protocol = new(motb_api.ProtocolGRPC)
			return motb
		}

		pipeProxy := func(backends *[]api.Backend) *core_xds.Proxy {
			proxy := xds_builders.Proxy().
				WithID(*core_xds.BuildProxyId("default", "backend")).
				WithDataplane(
					samples.DataplaneBackendBuilder().
						WithLabels(workloadLabels()),
				).
				WithMetadata(&core_xds.DataplaneMetadata{
					WorkDir: workDir,
					Features: xds_types.Features{
						xds_types.FeatureOtelViaKumaDp: true,
					},
				}).
				WithPolicies(xds_builders.MatchedPolicies().
					WithSingleItemPolicy(api.MeshMetricType, core_rules.SingleItemRules{
						Rules: []*core_rules.Rule{
							{
								Subset: []subsetutils.Tag{},
								Conf: api.Conf{
									Backends: backends,
								},
							},
						},
					}),
				).
				Build()
			proxy.OtelPipeBackends = &core_xds.OtelPipeBackends{}
			return proxy
		}

		It("inline endpoint should not add to pipe accumulator", func() {
			backends := &[]api.Backend{{
				Type: api.OpenTelemetryBackendType,
				OpenTelemetry: &api.OpenTelemetryBackend{
					Endpoint:        "otel-collector.observability.svc:4317",
					RefreshInterval: &k8s.Duration{Duration: 10 * time.Second},
				},
			}}

			proxy := pipeProxy(backends)
			resources := core_xds.NewResourceSet()
			plugin := v1alpha1.NewPlugin().(core_plugins.PolicyPlugin)

			Expect(plugin.Apply(resources, *xds_builders.Context().
				WithMeshBuilder(samples.MeshDefaultBuilder()).Build(), proxy)).To(Succeed())

			// Accumulator stays empty - inline endpoints don't use pipe
			Expect(proxy.OtelPipeBackends.Empty()).To(BeTrue())

			// Envoy-side OTel resources still created
			listeners, err := util_yaml.GetResourcesToYaml(resources, envoy_resource.ListenerType)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(listeners)).To(ContainSubstring("_kuma:metrics:opentelemetry:"))

			clusters, err := util_yaml.GetResourcesToYaml(resources, envoy_resource.ClusterType)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(clusters)).To(ContainSubstring("_kuma:metrics:opentelemetry:"))
		})

		It("backendRef should add to pipe accumulator and skip Envoy OTel resources", func() {
			motb := newMotb()
			backends := &[]api.Backend{{
				Type: api.OpenTelemetryBackendType,
				OpenTelemetry: &api.OpenTelemetryBackend{
					BackendRef: &common_api.BackendResourceRef{
						Kind: common_api.BackendResourceMeshOpenTelemetryBackend,
						Name: backendName,
					},
					RefreshInterval: &k8s.Duration{Duration: 10 * time.Second},
				},
			}}

			proxy := pipeProxy(backends)
			resources := core_xds.NewResourceSet()
			plugin := v1alpha1.NewPlugin().(core_plugins.PolicyPlugin)

			ctx := xds_builders.Context().
				WithMeshBuilder(samples.MeshDefaultBuilder()).
				WithMeshLocalResources([]core_model.Resource{motb}).
				Build()

			Expect(plugin.Apply(resources, *ctx, proxy)).To(Succeed())

			// Accumulator has the backend
			Expect(proxy.OtelPipeBackends.Empty()).To(BeFalse())
			pipeBackends := proxy.OtelPipeBackends.All()
			Expect(pipeBackends).To(HaveLen(1))
			Expect(pipeBackends[0].SocketPath).To(Equal(core_xds.OpenTelemetrySocketName(workDir, backendName)))
			Expect(pipeBackends[0].Endpoint).To(Equal("collector.mesh:4317"))
			Expect(pipeBackends[0].Metrics).ToNot(BeNil())
			Expect(pipeBackends[0].Metrics.RefreshInterval).To(Equal("10s"))

			// No Envoy-side OTel listener/cluster for this backend
			listeners, err := util_yaml.GetResourcesToYaml(resources, envoy_resource.ListenerType)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(listeners)).ToNot(ContainSubstring("_kuma:metrics:opentelemetry:"))

			clusters, err := util_yaml.GetResourcesToYaml(resources, envoy_resource.ClusterType)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(clusters)).ToNot(ContainSubstring("_kuma:metrics:opentelemetry:"))
		})

		It("mixed inline + backendRef should handle each correctly", func() {
			motb := newMotb()
			backends := &[]api.Backend{
				{
					Type: api.OpenTelemetryBackendType,
					OpenTelemetry: &api.OpenTelemetryBackend{
						Endpoint:        "inline-collector.svc:4317",
						RefreshInterval: &k8s.Duration{Duration: 10 * time.Second},
					},
				},
				{
					Type: api.OpenTelemetryBackendType,
					OpenTelemetry: &api.OpenTelemetryBackend{
						BackendRef: &common_api.BackendResourceRef{
							Kind: common_api.BackendResourceMeshOpenTelemetryBackend,
							Name: backendName,
						},
						RefreshInterval: &k8s.Duration{Duration: 10 * time.Second},
					},
				},
			}

			proxy := pipeProxy(backends)
			resources := core_xds.NewResourceSet()
			plugin := v1alpha1.NewPlugin().(core_plugins.PolicyPlugin)

			ctx := xds_builders.Context().
				WithMeshBuilder(samples.MeshDefaultBuilder()).
				WithMeshLocalResources([]core_model.Resource{motb}).
				Build()

			Expect(plugin.Apply(resources, *ctx, proxy)).To(Succeed())

			// Only backendRef in accumulator
			Expect(proxy.OtelPipeBackends.Empty()).To(BeFalse())
			pipeBackends := proxy.OtelPipeBackends.All()
			Expect(pipeBackends).To(HaveLen(1))
			Expect(pipeBackends[0].Endpoint).To(Equal("collector.mesh:4317"))
			Expect(pipeBackends[0].Metrics).ToNot(BeNil())

			// Envoy-side resources created for inline only
			listeners, err := util_yaml.GetResourcesToYaml(resources, envoy_resource.ListenerType)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(listeners)).To(ContainSubstring("inline-collector"))
			Expect(string(listeners)).ToNot(ContainSubstring("_kuma:metrics:opentelemetry:" + backendName))

			clusters, err := util_yaml.GetResourcesToYaml(resources, envoy_resource.ClusterType)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(clusters)).To(ContainSubstring("inline-collector"))
		})
	})
})
