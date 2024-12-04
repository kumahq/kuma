package v1alpha1_test

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/intstr"

	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrace/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshtrace/plugin/v1alpha1"
	gateway_plugin "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/pkg/test/xds/builders"
	xds_samples "github.com/kumahq/kuma/pkg/test/xds/samples"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

func getResource(
	resourceSet *core_xds.ResourceSet,
	typ envoy_resource.Type,
) []byte {
	resources, err := resourceSet.ListOf(typ).ToDeltaDiscoveryResponse()
	Expect(err).ToNot(HaveOccurred())
	actual, err := util_proto.ToYAML(resources)
	Expect(err).ToNot(HaveOccurred())

	return actual
}

var _ = Describe("MeshTrace", func() {
	type testCase struct {
		resources       []core_xds.Resource
		singleItemRules core_rules.SingleItemRules
		outbounds       xds_types.Outbounds
		goldenFile      string
	}
	backendMeshServiceIdentifier := core_model.TypedResourceIdentifier{
		ResourceIdentifier: core_model.ResourceIdentifier{
			Name:      "backend",
			Mesh:      "default",
			Namespace: "backend-ns",
			Zone:      "zone-1",
		},
		ResourceType: "MeshService",
		SectionName:  "",
	}
	inboundAndOutbound := func() []core_xds.Resource {
		return []core_xds.Resource{
			{
				Name:   "inbound",
				Origin: generator.OriginInbound,
				Resource: NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17777, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(HttpConnectionManager("127.0.0.1:17777", false)),
					)).MustBuild(),
			}, {
				Name:   "outbound",
				Origin: generator.OriginOutbound,
				Resource: NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(HttpConnectionManager("127.0.0.1:27777", false)),
					)).MustBuild(),
			},
		}
	}
	inboundAndOutboundRealMeshService := func() []core_xds.Resource {
		return []core_xds.Resource{
			{
				Name:   "inbound",
				Origin: generator.OriginInbound,
				Resource: NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17777, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(HttpConnectionManager("127.0.0.1:17777", false)),
					)).MustBuild(),
			}, {
				Name:   "outbound",
				Origin: generator.OriginOutbound,
				Resource: NewOutboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 27777, core_xds.SocketAddressProtocolTCP).
					Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
						Configure(HttpConnectionManager("127.0.0.1:27777", false)),
					)).MustBuild(),
				ResourceOrigin: &backendMeshServiceIdentifier,
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

			context := xds_samples.SampleContext()
			proxy := xds_builders.Proxy().
				WithDataplane(
					builders.Dataplane().
						WithName("backend").
						AddInbound(builders.Inbound().
							WithService("backend").
							WithAddress("127.0.0.1").
							WithPort(17777)),
				).
				WithOutbounds(given.outbounds).
				WithPolicies(xds_builders.MatchedPolicies().WithSingleItemPolicy(api.MeshTraceType, given.singleItemRules)).
				Build()
			context.Mesh.MeshServiceByIdentifier = map[core_model.ResourceIdentifier]*meshservice_api.MeshServiceResource{
				backendMeshServiceIdentifier.ResourceIdentifier : samples.MeshServiceBackend(),
			}
			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)

			Expect(plugin.Apply(resources, context, proxy)).To(Succeed())

			Expect(getResource(resources, envoy_resource.ListenerType)).
				To(matchers.MatchGoldenYAML(fmt.Sprintf("testdata/%s.listener.golden.yaml", given.goldenFile)))
			Expect(getResource(resources, envoy_resource.ClusterType)).
				To(matchers.MatchGoldenYAML(fmt.Sprintf("testdata/%s.cluster.golden.yaml", given.goldenFile)))
		},
		Entry("inbound/outbound for zipkin and real MeshService", testCase{
			resources: inboundAndOutboundRealMeshService(),
			outbounds: xds_types.Outbounds{
				{
					Address: "127.0.0.1",
					Port: 27777,
					Resource: &backendMeshServiceIdentifier,
				},
			},
			singleItemRules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{
					{
						Subset: []core_rules.Tag{},
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
									SharedSpanContext: pointer.To(true),
									ApiVersion:        pointer.To("httpProto"),
									TraceId128Bit:     pointer.To(true),
								},
							}},
						},
					},
				},
			},
			goldenFile: "inbound-outbound-zipkin-real-meshservice",
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
						Subset: []core_rules.Tag{},
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
									SharedSpanContext: pointer.To(true),
									ApiVersion:        pointer.To("httpProto"),
									TraceId128Bit:     pointer.To(true),
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
						Subset: []core_rules.Tag{},
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
									Endpoint: "jaeger-collector.mesh-observability:4317",
								},
							}},
						},
					},
				},
			},
			goldenFile: "inbound-outbound-otel",
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
						Subset: []core_rules.Tag{},
						Conf: api.Conf{
							Sampling: &api.Sampling{
								Random: pointer.To(intstr.FromInt(50)),
							},
							Backends: &[]api.Backend{{
								Datadog: &api.DatadogBackend{
									Url:          "http://ingest.datadog.eu:8126",
									SplitService: pointer.To(true),
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
						Subset: []core_rules.Tag{},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								Zipkin: &api.ZipkinBackend{
									Url:               "http://jaeger-collector.mesh-observability:9411/api/v2/spans",
									SharedSpanContext: pointer.To(true),
									TraceId128Bit:     pointer.To(true),
								},
							}},
						},
					},
				},
			},
			goldenFile: "empty-sampling",
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
						Subset: []core_rules.Tag{},
						Conf: api.Conf{
							Backends: &[]api.Backend{},
						},
					},
				},
			},
			goldenFile: "empty-backend-list",
		}),
	)
	type gatewayTestCase struct {
		rules core_rules.SingleItemRules
	}
	DescribeTable("should generate proper Envoy config for gateways",
		func(given gatewayTestCase) {
			resources := xds_context.NewResources()
			resources.MeshLocalResources[core_mesh.MeshGatewayType] = &core_mesh.MeshGatewayResourceList{
				Items: []*core_mesh.MeshGatewayResource{samples.GatewayResource()},
			}
			resources.MeshLocalResources[core_mesh.MeshGatewayRouteType] = &core_mesh.MeshGatewayRouteResourceList{
				Items: []*core_mesh.MeshGatewayRouteResource{samples.BackendGatewayRoute()},
			}

			xdsCtx := xds_samples.SampleContextWith(resources)

			proxy := xds_builders.Proxy().
				WithDataplane(samples.GatewayDataplaneBuilder()).
				WithPolicies(xds_builders.MatchedPolicies().WithSingleItemPolicy(api.MeshTraceType, given.rules)).
				Build()
			for n, p := range core_plugins.Plugins().ProxyPlugins() {
				Expect(p.Apply(context.Background(), xdsCtx.Mesh, proxy)).To(Succeed(), n)
			}
			gatewayGenerator := gateway_plugin.NewGenerator("test-zone")
			generatedResources, err := gatewayGenerator.Generate(context.Background(), nil, xdsCtx, proxy)
			Expect(err).NotTo(HaveOccurred())

			// when
			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)
			Expect(plugin.Apply(generatedResources, xdsCtx, proxy)).To(Succeed())

			nameSplit := strings.Split(GinkgoT().Name(), " ")
			name := nameSplit[len(nameSplit)-1]
			// then
			Expect(getResourceYaml(generatedResources.ListOf(envoy_resource.ListenerType))).
				To(matchers.MatchGoldenYAML(filepath.Join("testdata", fmt.Sprintf("%s.listeners.golden.yaml", name))))
		},
		Entry("simple-gateway", gatewayTestCase{
			rules: core_rules.SingleItemRules{
				Rules: []*core_rules.Rule{
					{
						Subset: []core_rules.Tag{},
						Conf: api.Conf{
							Backends: &[]api.Backend{{
								Zipkin: &api.ZipkinBackend{
									Url:               "http://jaeger-collector.mesh-observability:9411/api/v2/spans",
									SharedSpanContext: pointer.To(true),
									TraceId128Bit:     pointer.To(true),
								},
							}},
						},
					},
				},
			},
		}),
	)
})

func getResourceYaml(list core_xds.ResourceList) []byte {
	actualResource, err := util_proto.ToYAML(list[0].Resource)
	Expect(err).ToNot(HaveOccurred())
	return actualResource
}
