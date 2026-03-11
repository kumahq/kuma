package otelenv

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
)

var _ = Describe("DiscoverWithLookup", func() {
	It("should build inventory and summary", func() {
		env := map[string]string{
			"OTEL_EXPORTER_OTLP_ENDPOINT":         "https://otel-gateway.observability:4318",
			"OTEL_EXPORTER_OTLP_PROTOCOL":         "http/protobuf",
			"OTEL_EXPORTER_OTLP_HEADERS":          "authorization=token",
			"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT":  "https://tempo.observability:4318",
			"OTEL_EXPORTER_OTLP_LOGS_ENDPOINT":    "https://otel-gateway.observability:4318",
			"OTEL_EXPORTER_OTLP_METRICS_PROTOCOL": "grpc",
		}

		cfg := discoverWithLookup(true, func(name string) (string, bool) {
			value, ok := env[name]
			return value, ok
		})

		Expect(cfg.Inventory).ToNot(BeNil())
		Expect(cfg.Inventory.PipeEnabled).To(BeTrue())
		Expect(cfg.Inventory.Shared).ToNot(BeNil())
		Expect(cfg.Inventory.Shared.EndpointPresent).To(BeTrue())
		Expect(cfg.Inventory.Shared.EndpointParsedAsURL).To(BeTrue())
		Expect(cfg.Inventory.Shared.EffectiveProtocol).To(Equal(core_xds.OtelProtocolHTTPProtobuf))
		Expect(cfg.Inventory.Shared.EffectiveAuthMode).To(Equal(core_xds.OtelAuthModeHeaders))
		Expect(cfg.Inventory.Traces).ToNot(BeNil())
		Expect(cfg.Inventory.Traces.OverrideKinds).To(ConsistOf("endpoint"))
		Expect(cfg.Inventory.Traces.EndpointParsedAsURL).To(BeTrue())
		Expect(cfg.Inventory.Logs).ToNot(BeNil())
		Expect(cfg.Inventory.Logs.OverrideKinds).To(BeEmpty())
		Expect(cfg.Inventory.Metrics).ToNot(BeNil())
		Expect(cfg.Inventory.Metrics.OverrideKinds).To(ConsistOf("protocol"))

		summary := cfg.DynamicMetadataSummary()
		Expect(summary).To(HaveKeyWithValue("otel.env.shared.present", "true"))
		Expect(summary).To(HaveKeyWithValue("otel.env.shared.protocol", "http/protobuf"))
		Expect(summary).To(HaveKeyWithValue("otel.env.traces.overrideKinds", "endpoint"))
	})

	It("should report validation errors", func() {
		env := map[string]string{
			"OTEL_EXPORTER_OTLP_PROTOCOL":                  "binary",
			"OTEL_EXPORTER_OTLP_TIMEOUT":                   "later",
			"OTEL_EXPORTER_OTLP_COMPRESSION":               "brotli",
			"OTEL_EXPORTER_OTLP_TRACES_CLIENT_CERTIFICATE": "/cert",
		}

		cfg := discoverWithLookup(false, func(name string) (string, bool) {
			value, ok := env[name]
			return value, ok
		})

		Expect(cfg.Inventory).ToNot(BeNil())
		Expect(cfg.Inventory.ValidationErrors).To(ConsistOf("shared.protocol", "shared.timeout", "shared.compression", "traces.mtls"))
		Expect(cfg.Inventory.Shared).ToNot(BeNil())
		Expect(cfg.Inventory.Shared.EffectiveProtocol).To(Equal(core_xds.OtelProtocolUnknown))
		Expect(cfg.Inventory.Traces).ToNot(BeNil())
		Expect(cfg.Inventory.Traces.EffectiveAuthMode).To(Equal(core_xds.OtelAuthModeNone))
	})

	It("should not report validation error for compression=none", func() {
		env := map[string]string{
			"OTEL_EXPORTER_OTLP_COMPRESSION": "none",
		}

		cfg := discoverWithLookup(false, func(name string) (string, bool) {
			value, ok := env[name]
			return value, ok
		})

		Expect(cfg.Inventory).ToNot(BeNil())
		Expect(cfg.Inventory.ValidationErrors).To(BeEmpty())
		Expect(cfg.Inventory.Shared).ToNot(BeNil())
		Expect(cfg.Inventory.Shared.CompressionPresent).To(BeTrue())
	})
})
