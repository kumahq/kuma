package otelenv

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
)

var _ = Describe("DiscoverWithLookup", func() {
	It("should build inventory", func() {
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

	It("should treat empty env values as unset", func() {
		env := map[string]string{
			"OTEL_EXPORTER_OTLP_ENDPOINT": "   ",
		}

		cfg := discoverWithLookup(false, func(name string) (string, bool) {
			value, ok := env[name]
			return value, ok
		})

		Expect(cfg.Inventory).ToNot(BeNil())
		Expect(cfg.Inventory.Shared).To(BeNil())
		Expect(cfg.Inventory.ValidationErrors).To(BeEmpty())
	})

	It("should report a shared pathful endpoint as invalid when protocol defaults to grpc", func() {
		env := map[string]string{
			"OTEL_EXPORTER_OTLP_ENDPOINT": "https://collector.example:4317/custom",
		}

		cfg := discoverWithLookup(false, func(name string) (string, bool) {
			value, ok := env[name]
			return value, ok
		})

		Expect(cfg.Inventory).ToNot(BeNil())
		Expect(cfg.Inventory.ValidationErrors).To(ContainElement("shared.endpoint"))
	})

	It("should report a signal pathful endpoint as invalid when it inherits grpc from shared config", func() {
		env := map[string]string{
			"OTEL_EXPORTER_OTLP_PROTOCOL":        "grpc",
			"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT": "https://collector.example:4317/traces",
		}

		cfg := discoverWithLookup(false, func(name string) (string, bool) {
			value, ok := env[name]
			return value, ok
		})

		Expect(cfg.Inventory).ToNot(BeNil())
		Expect(cfg.Inventory.ValidationErrors).To(ContainElement("traces.endpoint"))
	})

	It("should trim protocol values before validation", func() {
		env := map[string]string{
			"OTEL_EXPORTER_OTLP_PROTOCOL": " http/protobuf ",
		}

		cfg := discoverWithLookup(false, func(name string) (string, bool) {
			value, ok := env[name]
			return value, ok
		})

		Expect(cfg.Inventory).ToNot(BeNil())
		Expect(cfg.Inventory.ValidationErrors).To(BeEmpty())
		Expect(cfg.Inventory.Shared).ToNot(BeNil())
		Expect(cfg.Inventory.Shared.EffectiveProtocol).To(Equal(core_xds.OtelProtocolHTTPProtobuf))
	})
})
