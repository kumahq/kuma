//nolint:staticcheck // SA1019 Test file: tests backward compatibility with deprecated core_rules.Rule
package v1alpha1_test

import (
	"fmt"
	"strings"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/intstr"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	"github.com/kumahq/kuma/v3/pkg/core/naming"
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	motb_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshopentelemetrybackend/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
	core_matchers "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/subsetutils"
	plugins_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtrace/api/v1alpha1"
	plugin "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtrace/plugin/v1alpha1"
	k8s_metadata "github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/v3/pkg/test/matchers"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	"github.com/kumahq/kuma/v3/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/v3/pkg/test/xds/builders"
	xds_samples "github.com/kumahq/kuma/v3/pkg/test/xds/samples"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	util_yaml "github.com/kumahq/kuma/v3/pkg/util/yaml"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
	. "github.com/kumahq/kuma/v3/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/metadata"
)

func otelBackendResource(name, address string) *motb_api.MeshOpenTelemetryBackendResource {
	motb := motb_api.NewMeshOpenTelemetryBackendResource()
	motb.SetMeta(&test_model.ResourceMeta{
		Mesh:   "default",
		Name:   name,
		Labels: map[string]string{mesh_proto.DisplayName: name},
	})
	motb.Spec.Endpoint = &motb_api.Endpoint{
		Address: pointer.To(address),
		Port:    pointer.To(int32(4317)),
	}
	motb.Spec.Protocol = pointer.To(motb_api.ProtocolGRPC)
	return motb
}

func otelBackendRef(name string) *common_api.BackendResourceRef {
	return &common_api.BackendResourceRef{
		Kind:   common_api.BackendResourceMeshOpenTelemetryBackend,
		Labels: map[string]string{mesh_proto.DisplayName: name},
	}
}

var _ = Describe("MeshTrace", func() {
	type testCase struct {
		resources       []core_xds.Resource
		singleItemRules core_rules.SingleItemRules
		outbounds       xds_types.Outbounds
		goldenFile      string
		features        xds_types.Features
		proxyLabels     map[string]string
		zone            string
		otelBackends    []*motb_api.MeshOpenTelemetryBackendResource
	}
	backendMeshServiceIdentifier := kri.Identifier{
		ResourceType: "MeshService",
		Mesh:         "default",
		Zone:         "zone-1",
		Namespace:    "backend-ns",
		Name:         "backend",
		SectionName:  "",
	}
	inboundAndOutbound := func() []core_xds.Resource {
		return []core_xds.Resource{
			{
				Name:   "inbound",
				Origin: metadata.OriginInbound,
				Resource: NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17777, core_xds.SocketAddressProtocolTCP, true).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(HttpConnectionManager("127.0.0.1:17777", false, nil, true)),
					)).MustBuild(),
			}, {
				Name:   "outbound",
				Origin: metadata.OriginOutbound,
				Resource: NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(HttpConnectionManager("127.0.0.1:27777", false, nil, true)),
					)).MustBuild(),
			},
		}
	}
	inboundAndOutboundRealMeshService := func() []core_xds.Resource {
		return []core_xds.Resource{
			{
				Name:   "inbound",
				Origin: metadata.OriginInbound,
				Resource: NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17777, core_xds.SocketAddressProtocolTCP, true).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(HttpConnectionManager("127.0.0.1:17777", false, nil, true)),
					)).MustBuild(),
			}, {
				Name:   "outbound",
				Origin: metadata.OriginOutbound,
				Resource: NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(HttpConnectionManager("127.0.0.1:27777", false, nil, true)),
					)).MustBuild(),
				ResourceOrigin: backendMeshServiceIdentifier,
			},
		}
	}
	// Unified naming listeners use contextual names matching what real proxy generators produce.
	inboundUnifiedName := naming.MustContextualInboundName(core_mesh.NewDataplaneResource(), uint32(17777))
	outboundUnifiedName := backendMeshServiceIdentifier.String()
	inboundAndOutboundUnifiedNaming := func() []core_xds.Resource {
		return []core_xds.Resource{
			{
				Name:   "inbound",
				Origin: metadata.OriginInbound,
				Resource: NewListenerBuilder(envoy_common.APIV3, inboundUnifiedName).
					Configure(InboundListener("127.0.0.1", 17777, core_xds.SocketAddressProtocolTCP, true)).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(HttpConnectionManager(inboundUnifiedName, false, nil, true)),
					)).MustBuild(),
			}, {
				Name:   "outbound",
				Origin: metadata.OriginOutbound,
				Resource: NewListenerBuilder(envoy_common.APIV3, outboundUnifiedName).
					Configure(OutboundListener("127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP)).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(HttpConnectionManager(outboundUnifiedName, false, nil, true)),
					)).MustBuild(),
				ResourceOrigin: backendMeshServiceIdentifier,
			},
		}
	}
	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			resources := core_xds.NewResourceSet()
			for _, res := range given.resources {
				r := res
				resources.Add(&r)
			}

			meshResources := xds_context.NewResources()
			meshResources.MeshLocalResources[v1alpha1.MeshServiceType] = &v1alpha1.MeshServiceResourceList{
				Items: []*v1alpha1.MeshServiceResource{samples.MeshServiceBackendBuilder().
					WithZone("zone-1").
					WithNamespace("backend-ns").
					Build()},
			}
			if len(given.otelBackends) > 0 {
				meshResources.MeshLocalResources[motb_api.MeshOpenTelemetryBackendType] = &motb_api.MeshOpenTelemetryBackendResourceList{
					Items: given.otelBackends,
				}
			}
			context := *xds_samples.SampleContextWith(meshResources).WithMeshBuilder(samples.MeshDefaultBuilder()).Build()
			dpBuilder := builders.Dataplane().
				WithName("backend").
				AddInbound(builders.Inbound().
					WithService("backend").
					WithAddress("127.0.0.1").
					WithPort(17777))
			if given.proxyLabels != nil {
				dpBuilder = dpBuilder.WithLabels(given.proxyLabels)
			}
			proxyBuilder := xds_builders.Proxy().
				WithDataplane(dpBuilder).
				WithMetadata(&core_xds.DataplaneMetadata{
					Features: given.features,
				}).
				WithOutbounds(given.outbounds).
				WithPolicies(xds_builders.MatchedPolicies().WithSingleItemPolicy(api.MeshTraceType, given.singleItemRules))
			if given.zone != "" {
				proxyBuilder = proxyBuilder.WithZone(given.zone)
			}
			proxy := proxyBuilder.Build()

			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)

			Expect(plugin.Apply(resources, context, proxy)).To(Succeed())

			resource, err := util_yaml.GetResourcesToYaml(resources, envoy_resource.ListenerType)
			Expect(err).ToNot(HaveOccurred())
			Expect(resource).To(matchers.MatchGoldenYAML(fmt.Sprintf("testdata/%s.listener.golden.yaml", given.goldenFile)))
			resource, err = util_yaml.GetResourcesToYaml(resources, envoy_resource.ClusterType)
			Expect(err).ToNot(HaveOccurred())
			Expect(resource).To(matchers.MatchGoldenYAML(fmt.Sprintf("testdata/%s.cluster.golden.yaml", given.goldenFile)))
		},
		Entry("inbound/outbound for zipkin and real MeshService", testCase{
			resources: inboundAndOutboundRealMeshService(),
			outbounds: xds_types.Outbounds{
				{
					Address:  "127.0.0.1",
					Port:     27777,
					Resource: backendMeshServiceIdentifier,
				},
			},
			singleItemRules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{
					{
						Subset: []subsetutils.Tag{},
						Conf: api.Conf{
							Tags: &[]api.Tag{
								{Name: "app", Literal: pointer.To("backend")},
								{Name: "app_code", Header: &api.HeaderTag{Name: "app_code"}},
								{Name: "client_id", Header: &api.HeaderTag{Name: "client_id", Default: pointer.To("none")}},
							},
							Sampling: &api.Sampling{
								Overall: pointer.To(intstr.FromInt(10)),
								Client:  pointer.To(intstr.FromInt(20)),
								Random:  pointer.To(intstr.FromInt(50)),
							},
							Backends: &[]api.Backend{{
								Zipkin: &api.ZipkinBackend{
									Url:               "http://jaeger-collector.mesh-observability:9411/api/v2/spans",
									SharedSpanContext: true,
									ApiVersion:        "httpProto",
									TraceId128Bit:     true,
								},
							}},
						},
					},
				},
			},
			goldenFile: "inbound-outbound-zipkin-real-meshservice",
		}),
		Entry("inbound/outbound for zipkin, real MeshService and unified naming", testCase{
			resources: inboundAndOutboundUnifiedNaming(),
			features: xds_types.Features{
				xds_types.FeatureUnifiedResourceNaming: true,
			},
			outbounds: xds_types.Outbounds{
				{
					Address:  "127.0.0.1",
					Port:     27777,
					Resource: backendMeshServiceIdentifier,
				},
			},
			singleItemRules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{
					{
						Subset: []subsetutils.Tag{},
						Origin: []core_model.ResourceMeta{
							&test_model.ResourceMeta{
								Mesh: "default",
								Name: "mt-1",
							},
						},
						Conf: api.Conf{
							Tags: &[]api.Tag{
								{Name: "app", Literal: pointer.To("backend")},
								{Name: "app_code", Header: &api.HeaderTag{Name: "app_code"}},
								{Name: "client_id", Header: &api.HeaderTag{Name: "client_id", Default: pointer.To("none")}},
							},
							Sampling: &api.Sampling{
								Overall: pointer.To(intstr.FromInt(10)),
								Client:  pointer.To(intstr.FromInt(20)),
								Random:  pointer.To(intstr.FromInt(50)),
							},
							Backends: &[]api.Backend{{
								Zipkin: &api.ZipkinBackend{
									Url:               "http://jaeger-collector.mesh-observability:9411/api/v2/spans",
									SharedSpanContext: true,
									ApiVersion:        "httpProto",
									TraceId128Bit:     true,
								},
							}},
						},
					},
				},
			},
			goldenFile: "inbound-outbound-zipkin-real-meshservice-unified-naming",
		}),
		Entry("inbound/outbound for zipkin", testCase{
			resources: inboundAndOutbound(),
			outbounds: xds_types.Outbounds{
				{
					LegacyOutbound: builders.Outbound().
						WithService("other-service").
						WithAddress("127.0.0.1").
						WithPort(27777).Build(),
				},
			},
			singleItemRules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{
					{
						Subset: []subsetutils.Tag{},
						Conf: api.Conf{
							Tags: &[]api.Tag{
								{Name: "app", Literal: pointer.To("backend")},
								{Name: "app_code", Header: &api.HeaderTag{Name: "app_code"}},
								{Name: "client_id", Header: &api.HeaderTag{Name: "client_id", Default: pointer.To("none")}},
							},
							Sampling: &api.Sampling{
								Overall: pointer.To(intstr.FromInt(10)),
								Client:  pointer.To(intstr.FromInt(20)),
								Random:  pointer.To(intstr.FromInt(50)),
							},
							Backends: &[]api.Backend{{
								Zipkin: &api.ZipkinBackend{
									Url:               "http://jaeger-collector.mesh-observability:9411/api/v2/spans",
									SharedSpanContext: true,
									ApiVersion:        "httpProto",
									TraceId128Bit:     true,
								},
							}},
						},
					},
				},
			},
			goldenFile: "inbound-outbound-zipkin",
		}),
		Entry("inbound/outbound for opentelemetry", testCase{
			resources: inboundAndOutbound(),
			outbounds: xds_types.Outbounds{
				{
					LegacyOutbound: builders.Outbound().
						WithService("other-service").
						WithAddress("127.0.0.1").
						WithPort(27777).Build(),
				},
			},
			singleItemRules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{
					{
						Subset: []subsetutils.Tag{},
						Conf: api.Conf{
							Tags: &[]api.Tag{
								{Name: "app", Literal: pointer.To("backend")},
								{Name: "app_code", Header: &api.HeaderTag{Name: "app_code"}},
								{Name: "client_id", Header: &api.HeaderTag{Name: "client_id", Default: pointer.To("none")}},
							},
							Sampling: &api.Sampling{
								Overall: pointer.To(intstr.FromInt(10)),
								Client:  pointer.To(intstr.FromInt(20)),
								Random:  pointer.To(intstr.FromInt(50)),
							},
							Backends: &[]api.Backend{{
								OpenTelemetry: &api.OpenTelemetryBackend{
									BackendRef: otelBackendRef("otel-collector"),
								},
							}},
						},
					},
				},
			},
			otelBackends: []*motb_api.MeshOpenTelemetryBackendResource{
				otelBackendResource("otel-collector", "jaeger-collector.mesh-observability"),
			},
			goldenFile: "inbound-outbound-otel",
		}),
		Entry("inbound/outbound for opentelemetry with ipv6 endpoint", testCase{
			resources: inboundAndOutbound(),
			outbounds: xds_types.Outbounds{
				{
					LegacyOutbound: builders.Outbound().
						WithService("other-service").
						WithAddress("127.0.0.1").
						WithPort(27777).Build(),
				},
			},
			singleItemRules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{
					{
						Subset: []subsetutils.Tag{},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								OpenTelemetry: &api.OpenTelemetryBackend{
									BackendRef: otelBackendRef("otel-collector-ipv6"),
								},
							}},
						},
					},
				},
			},
			otelBackends: []*motb_api.MeshOpenTelemetryBackendResource{
				otelBackendResource("otel-collector-ipv6", "2001:db8::1"),
			},
			goldenFile: "inbound-outbound-otel-ipv6",
		}),
		Entry("inbound/outbound for datadog", testCase{
			resources: inboundAndOutbound(),
			outbounds: xds_types.Outbounds{
				{
					LegacyOutbound: builders.Outbound().
						WithService("other-service").
						WithAddress("127.0.0.1").
						WithPort(27777).Build(),
				},
			},
			singleItemRules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{
					{
						Subset: []subsetutils.Tag{},
						Conf: api.Conf{
							Sampling: &api.Sampling{
								Random: pointer.To(intstr.FromInt(50)),
							},
							Backends: &[]api.Backend{{
								Datadog: &api.DatadogBackend{
									Url:          "http://ingest.datadog.eu:8126",
									SplitService: true,
								},
							}},
						},
					},
				},
			},
			goldenFile: "inbound-outbound-datadog",
		}),
		Entry("sampling is empty", testCase{
			resources: inboundAndOutbound(),
			outbounds: xds_types.Outbounds{
				{
					LegacyOutbound: builders.Outbound().
						WithService("other-service").
						WithAddress("127.0.0.1").
						WithPort(27777).Build(),
				},
			},
			singleItemRules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{
					{
						Subset: []subsetutils.Tag{},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								Zipkin: &api.ZipkinBackend{
									Url:               "http://jaeger-collector.mesh-observability:9411/api/v2/spans",
									SharedSpanContext: true,
									TraceId128Bit:     true,
								},
							}},
						},
					},
				},
			},
			goldenFile: "empty-sampling",
		}),
		Entry("inbound/outbound for zipkin with workload identity", testCase{
			resources: inboundAndOutbound(),
			outbounds: xds_types.Outbounds{
				{
					LegacyOutbound: builders.Outbound().
						WithService("other-service").
						WithAddress("127.0.0.1").
						WithPort(27777).Build(),
				},
			},
			singleItemRules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{
					{
						Subset: []subsetutils.Tag{},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								Zipkin: &api.ZipkinBackend{
									Url:               "http://jaeger-collector.mesh-observability:9411/api/v2/spans",
									SharedSpanContext: true,
									ApiVersion:        "httpProto",
									TraceId128Bit:     true,
								},
							}},
						},
					},
				},
			},
			goldenFile: "inbound-outbound-zipkin-workload-identity",
			proxyLabels: map[string]string{
				"kuma.io/workload":      "backend",
				mesh_proto.ZoneTag:      "zone-1",
				"k8s.kuma.io/namespace": "kuma-demo",
			},
			zone: "zone-1",
		}),
		Entry("inbound/outbound for zipkin, user-defined kuma.mesh tag not overridden", testCase{
			resources: inboundAndOutbound(),
			outbounds: xds_types.Outbounds{
				{
					LegacyOutbound: builders.Outbound().
						WithService("other-service").
						WithAddress("127.0.0.1").
						WithPort(27777).Build(),
				},
			},
			singleItemRules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{
					{
						Subset: []subsetutils.Tag{},
						Conf: api.Conf{
							Tags: &[]api.Tag{
								{Name: "kuma.mesh", Literal: pointer.To("user-mesh")},
							},
							Backends: &[]api.Backend{{
								Zipkin: &api.ZipkinBackend{
									Url:               "http://jaeger-collector.mesh-observability:9411/api/v2/spans",
									SharedSpanContext: true,
									ApiVersion:        "httpProto",
									TraceId128Bit:     true,
								},
							}},
						},
					},
				},
			},
			goldenFile: "inbound-outbound-zipkin-user-tag-no-override",
			proxyLabels: map[string]string{
				"kuma.io/workload":      "backend",
				mesh_proto.ZoneTag:      "zone-1",
				"k8s.kuma.io/namespace": "kuma-demo",
			},
			zone: "zone-1",
		}),
		Entry("backends list is empty", testCase{
			resources: inboundAndOutbound(),
			outbounds: xds_types.Outbounds{
				{
					LegacyOutbound: builders.Outbound().
						WithService("other-service").
						WithAddress("127.0.0.1").
						WithPort(27777).Build(),
				},
			},
			singleItemRules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{
					{
						Subset: []subsetutils.Tag{},
						Conf: api.Conf{
							Backends: &[]api.Backend{},
						},
					},
				},
			},
			goldenFile: "empty-backend-list",
		}),
	)

	It("should skip opentelemetry provider for legacy inline endpoint config without backendRef", func() {
		resources := core_xds.NewResourceSet()
		for _, resource := range inboundAndOutbound() {
			r := resource
			resources.Add(&r)
		}

		meshResources := xds_context.NewResources()
		meshResources.MeshLocalResources[v1alpha1.MeshServiceType] = &v1alpha1.MeshServiceResourceList{
			Items: []*v1alpha1.MeshServiceResource{samples.MeshServiceBackendBuilder().
				WithZone("zone-1").
				WithNamespace("backend-ns").
				Build()},
		}

		context := *xds_samples.SampleContextWith(meshResources).Build()
		proxy := xds_builders.Proxy().
			WithDataplane(
				builders.Dataplane().
					WithName("backend").
					AddInbound(builders.Inbound().
						WithService("backend").
						WithAddress("127.0.0.1").
						WithPort(17777)),
			).
			WithOutbounds(xds_types.Outbounds{
				{
					LegacyOutbound: builders.Outbound().
						WithService("other-service").
						WithAddress("127.0.0.1").
						WithPort(27777).Build(),
				},
			}).
			WithPolicies(xds_builders.MatchedPolicies().WithSingleItemPolicy(api.MeshTraceType, core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{
					{
						Subset: []subsetutils.Tag{},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								OpenTelemetry: &api.OpenTelemetryBackend{},
							}},
						},
					},
				},
			})).
			Build()

		meshTracePlugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)
		Expect(meshTracePlugin.Apply(resources, context, proxy)).To(Succeed())

		listenerResources, err := util_yaml.GetResourcesToYaml(resources, envoy_resource.ListenerType)
		Expect(err).ToNot(HaveOccurred())
		Expect(listenerResources).ToNot(ContainSubstring("envoy.tracers.opentelemetry"))

		clusterResources, err := util_yaml.GetResourcesToYaml(resources, envoy_resource.ClusterType)
		Expect(err).ToNot(HaveOccurred())
		Expect(strings.TrimSpace(string(clusterResources))).To(Equal("{}"))
	})

	It("should skip opentelemetry provider for dangling backendRef", func() {
		resources := core_xds.NewResourceSet()
		for _, resource := range inboundAndOutbound() {
			r := resource
			resources.Add(&r)
		}

		meshResources := xds_context.NewResources()
		meshResources.MeshLocalResources[v1alpha1.MeshServiceType] = &v1alpha1.MeshServiceResourceList{
			Items: []*v1alpha1.MeshServiceResource{samples.MeshServiceBackendBuilder().
				WithZone("zone-1").
				WithNamespace("backend-ns").
				Build()},
		}

		context := *xds_samples.SampleContextWith(meshResources).Build()
		proxy := xds_builders.Proxy().
			WithDataplane(
				builders.Dataplane().
					WithName("backend").
					AddInbound(builders.Inbound().
						WithService("backend").
						WithAddress("127.0.0.1").
						WithPort(17777)),
			).
			WithOutbounds(xds_types.Outbounds{
				{
					LegacyOutbound: builders.Outbound().
						WithService("other-service").
						WithAddress("127.0.0.1").
						WithPort(27777).Build(),
				},
			}).
			WithPolicies(xds_builders.MatchedPolicies().WithSingleItemPolicy(api.MeshTraceType, core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{
					{
						Subset: []subsetutils.Tag{},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								OpenTelemetry: &api.OpenTelemetryBackend{
									BackendRef: &common_api.BackendResourceRef{
										Kind: common_api.BackendResourceMeshOpenTelemetryBackend,
										Labels: map[string]string{
											"kuma.io/test": "non-existing",
										},
									},
								},
							}},
						},
					},
				},
			})).
			Build()

		meshTracePlugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)
		Expect(meshTracePlugin.Apply(resources, context, proxy)).To(Succeed())

		listenerResources, err := util_yaml.GetResourcesToYaml(resources, envoy_resource.ListenerType)
		Expect(err).ToNot(HaveOccurred())
		Expect(listenerResources).ToNot(ContainSubstring("envoy.tracers.opentelemetry"))

		clusterResources, err := util_yaml.GetResourcesToYaml(resources, envoy_resource.ClusterType)
		Expect(err).ToNot(HaveOccurred())
		Expect(strings.TrimSpace(string(clusterResources))).To(Equal("{}"))
	})

	It("should route opentelemetry via kuma-dp when feature is enabled", func() {
		const (
			workDir     = "/tmp"
			backendName = "otel-backend"
		)

		resources := core_xds.NewResourceSet()
		for _, resource := range inboundAndOutbound() {
			r := resource
			resources.Add(&r)
		}

		motb := motb_api.NewMeshOpenTelemetryBackendResource()
		motb.SetMeta(&test_model.ResourceMeta{
			Mesh:   "default",
			Name:   backendName,
			Labels: map[string]string{mesh_proto.DisplayName: backendName},
		})
		motb.Spec.Endpoint = &motb_api.Endpoint{
			Address: pointer.To("collector.mesh"),
			Port:    pointer.To(int32(4317)),
		}
		motb.Spec.Protocol = pointer.To(motb_api.ProtocolGRPC)

		meshResources := xds_context.NewResources()
		meshResources.MeshLocalResources[motb_api.MeshOpenTelemetryBackendType] = &motb_api.MeshOpenTelemetryBackendResourceList{
			Items: []*motb_api.MeshOpenTelemetryBackendResource{motb},
		}

		context := *xds_samples.SampleContextWith(meshResources).Build()
		proxy := xds_builders.Proxy().
			WithDataplane(
				builders.Dataplane().
					WithName("backend").
					AddInbound(builders.Inbound().
						WithService("backend").
						WithAddress("127.0.0.1").
						WithPort(17777)),
			).
			WithMetadata(&core_xds.DataplaneMetadata{
				WorkDir: workDir,
				Features: xds_types.Features{
					xds_types.FeatureOtelViaKumaDp: true,
				},
			}).
			WithOutbounds(xds_types.Outbounds{
				{
					LegacyOutbound: builders.Outbound().
						WithService("other-service").
						WithAddress("127.0.0.1").
						WithPort(27777).Build(),
				},
			}).
			WithPolicies(xds_builders.MatchedPolicies().WithSingleItemPolicy(api.MeshTraceType, core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{
					{
						Subset: []subsetutils.Tag{},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								OpenTelemetry: &api.OpenTelemetryBackend{
									BackendRef: &common_api.BackendResourceRef{
										Kind: common_api.BackendResourceMeshOpenTelemetryBackend,
										Labels: map[string]string{
											"kuma.io/display-name": backendName,
										},
									},
								},
							}},
						},
					},
				},
			})).
			Build()

		proxy.OtelPipeBackends = &core_xds.OtelPipeBackends{}

		meshTracePlugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)
		Expect(meshTracePlugin.Apply(resources, context, proxy)).To(Succeed())

		expectedSocket := core_xds.OpenTelemetrySocketName(workDir, backendName)

		clusterResources, err := util_yaml.GetResourcesToYaml(resources, envoy_resource.ClusterType)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(clusterResources)).To(ContainSubstring(expectedSocket))
		Expect(string(clusterResources)).ToNot(ContainSubstring("collector.mesh"))

		// Plugin adds to OtelPipeBackends accumulator instead of writing dynconf directly.
		// The generator writes the /otel route after all plugins run.
		Expect(proxy.OtelPipeBackends.Empty()).To(BeFalse())
		backends := proxy.OtelPipeBackends.All()
		Expect(backends).To(HaveLen(1))
		Expect(backends[0].SocketPath).To(Equal(expectedSocket))
		Expect(backends[0].Endpoint).To(Equal("collector.mesh:4317"))
		Expect(backends[0].UseHTTP).To(BeFalse())
		Expect(backends[0].Traces).ToNot(BeNil())
		Expect(backends[0].Traces.Enabled).To(BeTrue())
	})
})

func zoneEgressOnlyDataplane() *builders.DataplaneBuilder {
	return builders.Dataplane().
		WithName("zone-egress-1").
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
	return builders.Dataplane().
		WithName("backend").
		WithAddress("192.168.0.1").
		AddInbound(builders.Inbound().
			WithService("backend").
			WithAddress("192.168.0.1").
			WithPort(17777)).
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

func zoneEgressListenerResource() core_xds.Resource {
	name := naming.ContextualZoneEgressListenerName("ze-port")
	return core_xds.Resource{
		Name:   name,
		Origin: metadata.OriginEgress,
		Resource: NewListenerBuilder(envoy_common.APIV3, name).
			Configure(InboundListener("192.168.0.10", 10002, core_xds.SocketAddressProtocolTCP, true)).
			Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, "mes-http").
				Configure(MatchTransportProtocol("tls")).
				Configure(MatchServerNames("sni.extsvc.default.zone-1.aws-aurora.8443")).
				Configure(HttpConnectionManager("mes-http", false, nil, true)),
			)).
			Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, "mes-tcp").
				Configure(MatchTransportProtocol("tls")).
				Configure(MatchServerNames("sni.extsvc.default.zone-1.redis.6379")).
				Configure(TCPProxy("mes-tcp", plugins_xds.NewSplitBuilder().WithClusterName("mes-tcp").WithWeight(1).Build())),
			)).MustBuild(),
	}
}

func zoneIngressListenerResource() core_xds.Resource {
	name := naming.ContextualZoneIngressListenerName("zi-port")
	return core_xds.Resource{
		Name:   name,
		Origin: metadata.OriginIngress,
		Resource: NewListenerBuilder(envoy_common.APIV3, name).
			Configure(InboundListener("192.168.0.11", 10001, core_xds.SocketAddressProtocolTCP, true)).
			Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
				Configure(MatchTransportProtocol("tls")).
				Configure(MatchServerNames("backend{mesh=default}")).
				Configure(TCPProxy("backend", plugins_xds.NewSplitBuilder().WithClusterName("backend").WithWeight(1).Build())),
			)).MustBuild(),
	}
}

func mixedInboundAndZoneEgressResources() []core_xds.Resource {
	inbound := core_xds.Resource{
		Name:   "inbound:192.168.0.1:17777",
		Origin: metadata.OriginInbound,
		Resource: NewInboundListenerBuilder(envoy_common.APIV3, "192.168.0.1", 17777, core_xds.SocketAddressProtocolTCP, true).
			Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
				Configure(HttpConnectionManager("192.168.0.1:17777", false, nil, true)),
			)).MustBuild(),
	}
	egressName := naming.ContextualZoneEgressListenerName("ze-port")
	egress := core_xds.Resource{
		Name:   egressName,
		Origin: metadata.OriginEgress,
		Resource: NewListenerBuilder(envoy_common.APIV3, egressName).
			Configure(InboundListener("192.168.0.1", 10002, core_xds.SocketAddressProtocolTCP, true)).
			Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, "mes-http").
				Configure(MatchTransportProtocol("tls")).
				Configure(MatchServerNames("sni.extsvc.default.zone-1.aws-aurora.8443")).
				Configure(HttpConnectionManager("mes-http", false, nil, true)),
			)).MustBuild(),
	}
	return []core_xds.Resource{inbound, egress}
}

var _ = Describe("MeshTrace on zone proxy Dataplane", func() {
	type testCase struct {
		dp                  *builders.DataplaneBuilder
		resources           []core_xds.Resource
		singleItemRules     core_rules.SingleItemRules
		inboundTagsDisabled bool
		goldenFile          string
		otelBackends        []*motb_api.MeshOpenTelemetryBackendResource
	}
	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			rs := core_xds.NewResourceSet()
			for _, res := range given.resources {
				r := res
				rs.Add(&r)
			}

			meshResources := xds_context.NewResources()
			if len(given.otelBackends) > 0 {
				meshResources.MeshLocalResources[motb_api.MeshOpenTelemetryBackendType] = &motb_api.MeshOpenTelemetryBackendResourceList{
					Items: given.otelBackends,
				}
			}
			ctxBuilder := xds_samples.SampleContextWith(meshResources).
				WithMeshBuilder(samples.MeshDefaultBuilder())
			if given.inboundTagsDisabled {
				ctxBuilder = ctxBuilder.With(func(c *xds_context.Context) {
					c.ControlPlane.InboundTagsDisabled = true
				})
			}
			xdsCtx := *ctxBuilder.Build()

			proxy := xds_builders.Proxy().
				WithDataplane(given.dp).
				WithMetadata(&core_xds.DataplaneMetadata{}).
				WithPolicies(xds_builders.MatchedPolicies().
					WithSingleItemPolicy(api.MeshTraceType, given.singleItemRules)).
				WithZone("zone-1").
				Build()

			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)
			Expect(plugin.Apply(rs, xdsCtx, proxy)).To(Succeed())

			listenerYaml, err := util_yaml.GetResourcesToYaml(rs, envoy_resource.ListenerType)
			Expect(err).ToNot(HaveOccurred())
			Expect(listenerYaml).To(matchers.MatchGoldenYAML(fmt.Sprintf("testdata/%s.listener.golden.yaml", given.goldenFile)))

			clusterYaml, err := util_yaml.GetResourcesToYaml(rs, envoy_resource.ClusterType)
			Expect(err).ToNot(HaveOccurred())
			Expect(clusterYaml).To(matchers.MatchGoldenYAML(fmt.Sprintf("testdata/%s.cluster.golden.yaml", given.goldenFile)))
		},
		Entry("zone-egress-only zipkin", testCase{
			dp:        zoneEgressOnlyDataplane(),
			resources: []core_xds.Resource{zoneEgressListenerResource()},
			singleItemRules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{{
					Subset: []subsetutils.Tag{},
					Conf: api.Conf{
						Backends: &[]api.Backend{{
							Zipkin: &api.ZipkinBackend{
								Url:               "http://jaeger-collector.mesh-observability:9411/api/v2/spans",
								SharedSpanContext: true,
								TraceId128Bit:     true,
							},
						}},
					},
				}},
			},
			goldenFile: "zone-egress-only-zipkin",
		}),
		Entry("zone-egress-only datadog", testCase{
			dp:        zoneEgressOnlyDataplane(),
			resources: []core_xds.Resource{zoneEgressListenerResource()},
			singleItemRules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{{
					Subset: []subsetutils.Tag{},
					Conf: api.Conf{
						Backends: &[]api.Backend{{
							Datadog: &api.DatadogBackend{
								Url:          "http://ingest.datadog.eu:8126",
								SplitService: true,
							},
						}},
					},
				}},
			},
			goldenFile: "zone-egress-only-datadog",
		}),
		Entry("zone-egress-only opentelemetry", testCase{
			dp:        zoneEgressOnlyDataplane(),
			resources: []core_xds.Resource{zoneEgressListenerResource()},
			singleItemRules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{{
					Subset: []subsetutils.Tag{},
					Conf: api.Conf{
						Backends: &[]api.Backend{{
							OpenTelemetry: &api.OpenTelemetryBackend{
								BackendRef: otelBackendRef("otel-collector"),
							},
						}},
					},
				}},
			},
			otelBackends: []*motb_api.MeshOpenTelemetryBackendResource{
				otelBackendResource("otel-collector", "jaeger-collector.mesh-observability"),
			},
			goldenFile: "zone-egress-only-otel",
		}),
		Entry("zone-ingress-only zipkin", testCase{
			dp:        zoneIngressOnlyDataplane("zone-ingress-1"),
			resources: []core_xds.Resource{zoneIngressListenerResource()},
			singleItemRules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{{
					Subset: []subsetutils.Tag{},
					Conf: api.Conf{
						Backends: &[]api.Backend{{
							Zipkin: &api.ZipkinBackend{
								Url:               "http://jaeger-collector.mesh-observability:9411/api/v2/spans",
								SharedSpanContext: true,
								TraceId128Bit:     true,
							},
						}},
					},
				}},
			},
			goldenFile: "zone-ingress-only-zipkin",
		}),
		Entry("zone-ingress-only datadog", testCase{
			dp:        zoneIngressOnlyDataplane("zone-ingress-1"),
			resources: []core_xds.Resource{zoneIngressListenerResource()},
			singleItemRules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{{
					Subset: []subsetutils.Tag{},
					Conf: api.Conf{
						Backends: &[]api.Backend{{
							Datadog: &api.DatadogBackend{
								Url:          "http://ingest.datadog.eu:8126",
								SplitService: true,
							},
						}},
					},
				}},
			},
			goldenFile: "zone-ingress-only-datadog",
		}),
		Entry("zone-ingress-only opentelemetry", testCase{
			dp:        zoneIngressOnlyDataplane("zone-ingress-1"),
			resources: []core_xds.Resource{zoneIngressListenerResource()},
			singleItemRules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{{
					Subset: []subsetutils.Tag{},
					Conf: api.Conf{
						Backends: &[]api.Backend{{
							OpenTelemetry: &api.OpenTelemetryBackend{
								BackendRef: otelBackendRef("otel-collector"),
							},
						}},
					},
				}},
			},
			otelBackends: []*motb_api.MeshOpenTelemetryBackendResource{
				otelBackendResource("otel-collector", "jaeger-collector.mesh-observability"),
			},
			goldenFile: "zone-ingress-only-otel",
		}),
		Entry("mixed inbound and zone egress zipkin", testCase{
			dp:        mixedInboundAndZoneEgressDataplane(),
			resources: mixedInboundAndZoneEgressResources(),
			singleItemRules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{{
					Subset: []subsetutils.Tag{},
					Conf: api.Conf{
						Backends: &[]api.Backend{{
							Zipkin: &api.ZipkinBackend{
								Url:               "http://jaeger-collector.mesh-observability:9411/api/v2/spans",
								SharedSpanContext: true,
								TraceId128Bit:     true,
							},
						}},
					},
				}},
			},
			goldenFile: "mixed-inbound-and-zone-egress-zipkin",
		}),
		Entry("mixed inbound and zone egress datadog", testCase{
			dp:        mixedInboundAndZoneEgressDataplane(),
			resources: mixedInboundAndZoneEgressResources(),
			singleItemRules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{{
					Subset: []subsetutils.Tag{},
					Conf: api.Conf{
						Backends: &[]api.Backend{{
							Datadog: &api.DatadogBackend{
								Url:          "http://ingest.datadog.eu:8126",
								SplitService: true,
							},
						}},
					},
				}},
			},
			goldenFile: "mixed-inbound-and-zone-egress-datadog",
		}),
		Entry("mixed inbound and zone egress opentelemetry", testCase{
			dp:        mixedInboundAndZoneEgressDataplane(),
			resources: mixedInboundAndZoneEgressResources(),
			singleItemRules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{{
					Subset: []subsetutils.Tag{},
					Conf: api.Conf{
						Backends: &[]api.Backend{{
							OpenTelemetry: &api.OpenTelemetryBackend{
								BackendRef: otelBackendRef("otel-collector"),
							},
						}},
					},
				}},
			},
			otelBackends: []*motb_api.MeshOpenTelemetryBackendResource{
				otelBackendResource("otel-collector", "jaeger-collector.mesh-observability"),
			},
			goldenFile: "mixed-inbound-and-zone-egress-otel",
		}),
		// Spec declares a zone-egress listener but the xDS ResourceSet has no
		// matching resource. The plugin must skip the lookup miss without
		// producing listener output; the meshtrace cluster is still added.
		Entry("zone-egress-only with missing xDS listener", testCase{
			dp:        zoneEgressOnlyDataplane(),
			resources: nil,
			singleItemRules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{{
					Subset: []subsetutils.Tag{},
					Conf: api.Conf{
						Backends: &[]api.Backend{{
							Zipkin: &api.ZipkinBackend{
								Url:               "http://jaeger-collector.mesh-observability:9411/api/v2/spans",
								SharedSpanContext: true,
								TraceId128Bit:     true,
							},
						}},
					},
				}},
			},
			goldenFile: "zone-egress-only-missing-listener",
		}),
		// IdentifyingName(true) walks the workload label first. For a
		// zone-proxy-only DPP without a workload label, it returns "unknown"
		// and the fallback to Dataplane.Name fires the same way as the default
		// path. Reuses the zone-egress-only-zipkin golden to pin that equivalence.
		Entry("zone-egress-only zipkin with inboundTagsDisabled", testCase{
			dp:                  zoneEgressOnlyDataplane(),
			resources:           []core_xds.Resource{zoneEgressListenerResource()},
			inboundTagsDisabled: true,
			singleItemRules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{{
					Subset: []subsetutils.Tag{},
					Conf: api.Conf{
						Backends: &[]api.Backend{{
							Zipkin: &api.ZipkinBackend{
								Url:               "http://jaeger-collector.mesh-observability:9411/api/v2/spans",
								SharedSpanContext: true,
								TraceId128Bit:     true,
							},
						}},
					},
				}},
			},
			goldenFile: "zone-egress-only-zipkin",
		}),
		// When the kuma.io/workload label is set (K8s pod-to-Dataplane mapping
		// always populates it for zone-proxy DPPs), the unknown-fallback should
		// prefer the workload name over Dataplane.Name. The Dataplane.Name on K8s
		// is the pod name (includes pod-hash + random suffix) and churns on every
		// rollout; the workload label is stable across restarts.
		Entry("zone-egress-only datadog with workload label", testCase{
			dp: zoneEgressOnlyDataplane().WithLabels(map[string]string{
				k8s_metadata.KumaWorkload: "kuma-default-egress",
			}),
			resources: []core_xds.Resource{zoneEgressListenerResource()},
			singleItemRules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{{
					Subset: []subsetutils.Tag{},
					Conf: api.Conf{
						Backends: &[]api.Backend{{
							Datadog: &api.DatadogBackend{
								Url:          "http://datadog-collector.mesh-observability:8126",
								SplitService: true,
							},
						}},
					},
				}},
			},
			goldenFile: "zone-egress-only-datadog-workload-label",
		}),
	)
})

var _ = Describe("MeshTrace on mixed Dataplane with sectionName targetRef", func() {
	// A MeshTrace policy that targets only the zone egress section name must apply tracing
	// to that listener and leave the regular inbound untouched. Exercises the matcher path
	// (not a hand-built SingleItemRules), so it catches the per-section filtering in Apply.
	It("only configures the targeted zone listener", func() {
		dpp := mixedInboundAndZoneEgressDataplane().Build()

		mt := &api.MeshTraceResource{
			Meta: &test_model.ResourceMeta{Mesh: "default", Name: "mt-section"},
			Spec: &api.MeshTrace{
				TargetRef: &common_api.TopLevelTargetRef{
					Kind:        common_api.Dataplane,
					SectionName: pointer.To("ze-port"),
				},
				Default: api.Conf{
					Backends: &[]api.Backend{{
						Zipkin: &api.ZipkinBackend{
							Url:               "http://jaeger-collector.mesh-observability:9411/api/v2/spans",
							SharedSpanContext: true,
							TraceId128Bit:     true,
						},
					}},
				},
			},
		}

		meshResources := xds_context.NewResources()
		meshResources.MeshLocalResources[api.MeshTraceType] = &api.MeshTraceResourceList{
			Items: []*api.MeshTraceResource{mt},
		}

		matched, err := core_matchers.MatchedPolicies(api.MeshTraceType, dpp, meshResources)
		Expect(err).ToNot(HaveOccurred())
		Expect(matched.DataplanePolicies).To(HaveLen(1), "matcher must include the MeshTrace policy")
		Expect(matched.SingleItemRules.Rules).To(HaveLen(1), "matcher must produce a single-item rule")

		rs := core_xds.NewResourceSet()
		for _, res := range mixedInboundAndZoneEgressResources() {
			r := res
			rs.Add(&r)
		}

		xdsCtx := *xds_samples.SampleContextWith(meshResources).WithMeshBuilder(samples.MeshDefaultBuilder()).Build()
		proxy := xds_builders.Proxy().
			WithDataplane(mixedInboundAndZoneEgressDataplane()).
			WithMetadata(&core_xds.DataplaneMetadata{}).
			WithZone("zone-1").
			Build()
		proxy.Policies.Dynamic = map[core_model.ResourceType]core_xds.TypedMatchingPolicies{
			api.MeshTraceType: matched,
		}

		Expect(plugin.NewPlugin().(core_plugins.PolicyPlugin).Apply(rs, xdsCtx, proxy)).To(Succeed())

		listenerYaml, err := util_yaml.GetResourcesToYaml(rs, envoy_resource.ListenerType)
		Expect(err).ToNot(HaveOccurred())
		out := string(listenerYaml)

		// The targeted zone listener must contain the tracing provider.
		Expect(out).To(ContainSubstring("self_zoneegress_dp_ze-port"))
		Expect(out).To(ContainSubstring("envoy.tracers.zipkin"))

		// The regular inbound must NOT be configured with tracing — exactly one
		// `tracing:` block (on the zone listener) is expected.
		Expect(strings.Count(out, "tracing:")).To(Equal(1),
			"regular inbound must stay untouched when targetRef.sectionName scopes the policy to a zone listener")
	})
})
