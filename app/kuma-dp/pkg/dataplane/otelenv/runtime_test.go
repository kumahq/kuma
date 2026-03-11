package otelenv

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	motb_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshopentelemetrybackend/api/v1alpha1"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
)

var _ = Describe("ResolveBackend", func() {
	It("should merge explicit config with shared and signal env overrides", func() {
		cfg := Config{
			Shared: Layer{
				Protocol:    FieldValue{Present: true, Value: "http/protobuf"},
				Headers:     FieldValue{Present: true, Value: "authorization=token"},
				Compression: FieldValue{Present: true, Value: "gzip"},
				Timeout:     FieldValue{Present: true, Value: "5000"},
			},
			Logs: Layer{
				Endpoint: FieldValue{Present: true, Value: "https://logs.example:4318/v1/logs"},
			},
		}
		backend := core_xds.OtelPipeBackend{
			Endpoint: "collector.example:4317",
			EnvPolicy: core_xds.OtelResolvedEnvPolicy{
				Mode:                 motb_api.EnvModeOptional,
				Precedence:           motb_api.EnvPrecedenceEnvFirst,
				AllowSignalOverrides: true,
			},
			Traces: &core_xds.OtelSignalRuntimePlan{
				Enabled: true,
			},
			Logs: &core_xds.OtelSignalRuntimePlan{
				Enabled:         true,
				EnvInputPresent: true,
				OverrideKinds:   []string{"endpoint"},
			},
		}

		runtime := cfg.ResolveBackend(backend)

		Expect(runtime.Traces.Enabled).To(BeTrue())
		Expect(runtime.Traces.Transport.Protocol).To(Equal(core_xds.OtelProtocolHTTPProtobuf))
		Expect(runtime.Traces.Transport.Endpoint).To(Equal("collector.example:4317"))
		Expect(runtime.Traces.Transport.Headers).To(HaveKeyWithValue("authorization", "token"))
		Expect(runtime.Traces.Transport.Compression).To(Equal("gzip"))
		Expect(runtime.Traces.Transport.Timeout).To(Equal(5 * time.Second))
		Expect(runtime.Traces.HTTPPath).To(Equal("/v1/traces"))

		Expect(runtime.Logs.Enabled).To(BeTrue())
		Expect(runtime.Logs.Transport.Protocol).To(Equal(core_xds.OtelProtocolHTTPProtobuf))
		Expect(runtime.Logs.Transport.Endpoint).To(Equal("logs.example:4318"))
		Expect(runtime.Logs.Transport.UseTLS).To(BeTrue())
		Expect(runtime.Logs.HTTPPath).To(Equal("/v1/logs"))
	})

	It("should keep explicit config when env is blocked by policy", func() {
		cfg := Config{
			Shared: Layer{
				Endpoint: FieldValue{Present: true, Value: "https://env.example:4318"},
				Protocol: FieldValue{Present: true, Value: "http/protobuf"},
			},
		}
		backend := core_xds.OtelPipeBackend{
			Endpoint: "collector.example:4317",
			EnvPolicy: core_xds.OtelResolvedEnvPolicy{
				Mode:       motb_api.EnvModeDisabled,
				Precedence: motb_api.EnvPrecedenceEnvFirst,
			},
			Traces: &core_xds.OtelSignalRuntimePlan{
				Enabled:        true,
				BlockedReasons: []string{core_xds.OtelBlockedReasonEnvDisabledByPolicy},
			},
		}

		runtime := cfg.ResolveBackend(backend)

		Expect(runtime.Traces.Transport.Protocol).To(Equal(core_xds.OtelProtocolGRPC))
		Expect(runtime.Traces.Transport.Endpoint).To(Equal("collector.example:4317"))
		Expect(runtime.Traces.HTTPPath).To(BeEmpty())
	})

	It("should leave blocked signals without a transport", func() {
		cfg := Config{
			Shared: Layer{
				Endpoint: FieldValue{Present: true, Value: "https://env.example:4318"},
			},
		}
		backend := core_xds.OtelPipeBackend{
			Endpoint: "collector.example:4317",
			Traces: &core_xds.OtelSignalRuntimePlan{
				Enabled:        true,
				BlockedReasons: []string{core_xds.OtelBlockedReasonRequiredEnvMissing},
			},
		}

		runtime := cfg.ResolveBackend(backend)

		Expect(runtime.Traces.Enabled).To(BeTrue())
		Expect(runtime.Traces.BlockedReasons).To(ContainElement(core_xds.OtelBlockedReasonRequiredEnvMissing))
		Expect(runtime.Traces.Transport.Endpoint).To(BeEmpty())
		Expect(runtime.Traces.Transport.Protocol).To(BeEmpty())
	})
})
