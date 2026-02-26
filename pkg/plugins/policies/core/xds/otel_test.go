package xds_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	motb_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshopentelemetrybackend/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
	policies_xds "github.com/kumahq/kuma/v2/pkg/plugins/policies/core/xds"
	test_model "github.com/kumahq/kuma/v2/pkg/test/resources/model"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/v2/pkg/xds/context"
)

var _ = Describe("FullPath", func() {
	DescribeTable("should build correct path",
		func(basePath *string, suffix string, expected string) {
			r := &policies_xds.ResolvedOtelBackend{Path: basePath}
			Expect(r.FullPath(suffix)).To(Equal(expected))
		},
		Entry("nil base path", nil, policies_xds.OtelTracesPathSuffix, "/v1/traces"),
		Entry("root base path", pointer.To("/"), policies_xds.OtelTracesPathSuffix, "/v1/traces"),
		Entry("custom base path", pointer.To("/custom"), policies_xds.OtelTracesPathSuffix, "/custom/v1/traces"),
		Entry("trailing slash base path", pointer.To("/custom/"), policies_xds.OtelMetricsPathSuffix, "/custom/v1/metrics"),
		Entry("logs suffix", pointer.To("/otel"), policies_xds.OtelLogsPathSuffix, "/otel/v1/logs"),
	)
})

var _ = Describe("ResolveOtelBackend", func() {
	dummyParser := func(ep string) *core_xds.Endpoint {
		return &core_xds.Endpoint{Target: ep, Port: 4317}
	}
	dummyNamer := func(ep string) string { return ep }
	emptyResources := xds_context.Resources{}

	It("should return nil when no config sources exist", func() {
		result := policies_xds.ResolveOtelBackend(
			nil, "", dummyParser, dummyNamer, emptyResources, "",
		)
		Expect(result).To(BeNil())
	})

	Describe("priority order", func() {
		backendRef := &common_api.TargetRef{
			Kind: "MeshOpenTelemetryBackend",
			Name: pointer.To("my-backend"),
		}
		motbList := &motb_api.MeshOpenTelemetryBackendResourceList{}

		It("should prefer backendRef over inline endpoint", func() {
			// backendRef will be dangling (no MOTBs in resources), but it should
			// NOT fall through to inline endpoint
			resources := xds_context.Resources{
				MeshLocalResources: map[core_model.ResourceType]core_model.ResourceList{
					motb_api.MeshOpenTelemetryBackendType: motbList,
				},
			}
			result := policies_xds.ResolveOtelBackend(
				backendRef, "inline-collector:4317", dummyParser, dummyNamer, resources, "",
			)
			// Dangling backendRef returns nil, does NOT fall through
			Expect(result).To(BeNil())
		})

		It("should resolve inline endpoint when no backendRef", func() {
			result := policies_xds.ResolveOtelBackend(
				nil, "inline-collector:4317", dummyParser, dummyNamer, emptyResources, "",
			)
			Expect(result).ToNot(BeNil())
			Expect(result.Endpoint.Target).To(Equal("inline-collector:4317"))
			Expect(result.Protocol).To(Equal(motb_api.ProtocolGRPC))
		})
	})

	Describe("inline endpoint uses ParseOtelEndpoint", func() {
		It("should resolve via ParseOtelEndpoint when no backendRef", func() {
			result := policies_xds.ResolveOtelBackend(
				nil, "collector:4317", policies_xds.ParseOtelEndpoint, func(ep string) string { return ep }, emptyResources, "",
			)
			Expect(result).ToNot(BeNil())
			Expect(result.Endpoint.Target).To(Equal("collector"))
			Expect(result.Endpoint.Port).To(Equal(uint32(4317)))
		})
	})

	Describe("nodeEndpoint resolution", func() {
		makeResources := func(backend *motb_api.MeshOpenTelemetryBackendResource) xds_context.Resources {
			list := &motb_api.MeshOpenTelemetryBackendResourceList{Items: []*motb_api.MeshOpenTelemetryBackendResource{backend}}
			return xds_context.Resources{
				MeshLocalResources: map[core_model.ResourceType]core_model.ResourceList{
					motb_api.MeshOpenTelemetryBackendType: list,
				},
			}
		}
		backendRef := &common_api.TargetRef{
			Kind: "MeshOpenTelemetryBackend",
			Name: pointer.To("daemonset-collector"),
		}

		It("should use nodeHostIP when nodeEndpoint is set", func() {
			backend := motb_api.NewMeshOpenTelemetryBackendResource()
			backend.SetMeta(&test_model.ResourceMeta{Name: "daemonset-collector", Mesh: "default"})
			backend.Spec.NodeEndpoint = &motb_api.NodeEndpoint{Port: 4317}
			backend.Spec.Protocol = motb_api.ProtocolGRPC

			result := policies_xds.ResolveOtelBackend(
				backendRef, "", dummyParser, dummyNamer, makeResources(backend), "192.168.1.5",
			)
			Expect(result).ToNot(BeNil())
			Expect(result.Endpoint.Target).To(Equal("192.168.1.5"))
			Expect(result.Endpoint.Port).To(Equal(uint32(4317)))
		})

		It("should fall back to 127.0.0.1 when nodeHostIP is empty", func() {
			backend := motb_api.NewMeshOpenTelemetryBackendResource()
			backend.SetMeta(&test_model.ResourceMeta{Name: "daemonset-collector", Mesh: "default"})
			backend.Spec.NodeEndpoint = &motb_api.NodeEndpoint{Port: 4317}
			backend.Spec.Protocol = motb_api.ProtocolGRPC

			result := policies_xds.ResolveOtelBackend(
				backendRef, "", dummyParser, dummyNamer, makeResources(backend), "",
			)
			Expect(result).ToNot(BeNil())
			Expect(result.Endpoint.Target).To(Equal("127.0.0.1"))
			Expect(result.Endpoint.Port).To(Equal(uint32(4317)))
		})
	})
})

var _ = Describe("CollectorEndpointString", func() {
	DescribeTable("should format endpoint correctly",
		func(endpoint *core_xds.Endpoint, expected string) {
			Expect(policies_xds.CollectorEndpointString(endpoint)).To(Equal(expected))
		},
		Entry("ipv4 host and port", &core_xds.Endpoint{Target: "10.0.0.1", Port: 4317}, "10.0.0.1:4317"),
		Entry("ipv6 host and port", &core_xds.Endpoint{Target: "2001:db8::1", Port: 4318}, "[2001:db8::1]:4318"),
		Entry("host without port", &core_xds.Endpoint{Target: "collector.mesh"}, "collector.mesh"),
	)
})

var _ = Describe("ParseOtelEndpoint", func() {
	DescribeTable("should parse endpoint correctly",
		func(input string, expectedTarget string, expectedPort uint32) {
			ep := policies_xds.ParseOtelEndpoint(input)
			Expect(ep.Target).To(Equal(expectedTarget))
			Expect(ep.Port).To(Equal(expectedPort))
		},
		Entry("host:port", "collector:4317", "collector", uint32(4317)),
		Entry("host only", "collector", "collector", uint32(4317)),
		Entry("ipv4:port", "10.0.0.1:4318", "10.0.0.1", uint32(4318)),
		Entry("bracketed ipv6:port", "[2001:db8::1]:4317", "2001:db8::1", uint32(4317)),
		Entry("bare ipv6", "[2001:db8::1]", "2001:db8::1", uint32(4317)),
		Entry("bare ipv6 no brackets", "2001:db8::1", "2001:db8::1", uint32(4317)),
	)
})
