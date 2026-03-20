package otelenv

import (
	"os"
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
			EnvPolicy: &core_xds.OtelResolvedEnvPolicy{
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
		Expect(runtime.Logs.Transport.UseTLS).To(HaveValue(BeTrue()))
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
			EnvPolicy: &core_xds.OtelResolvedEnvPolicy{
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

	It("should ignore incomplete optional mTLS env overrides", func() {
		cfg := Config{
			Shared: Layer{
				Endpoint:          FieldValue{Present: true, Value: "https://env.example:4318"},
				Protocol:          FieldValue{Present: true, Value: "http/protobuf"},
				ClientCertificate: FieldValue{Present: true, Value: "/tmp/client.crt"},
			},
		}
		backend := core_xds.OtelPipeBackend{
			Endpoint: "collector.example:4317",
			EnvPolicy: &core_xds.OtelResolvedEnvPolicy{
				Mode:       motb_api.EnvModeOptional,
				Precedence: motb_api.EnvPrecedenceEnvFirst,
			},
			Traces: &core_xds.OtelSignalRuntimePlan{
				Enabled:         true,
				EnvInputPresent: true,
			},
		}

		runtime := cfg.ResolveBackend(backend)

		Expect(runtime.Traces.Transport.Protocol).To(Equal(core_xds.OtelProtocolHTTPProtobuf))
		Expect(runtime.Traces.Transport.Endpoint).To(Equal("env.example:4318"))
		Expect(runtime.Traces.Transport.UseTLS).To(HaveValue(BeTrue()))
		Expect(runtime.Traces.Transport.ClientCertificate).To(BeEmpty())
		Expect(runtime.Traces.Transport.ClientKey).To(BeEmpty())
	})

	It("should prefer explicit endpoint+protocol over env in ExplicitFirst mode", func() {
		cfg := Config{
			Shared: Layer{
				Endpoint: FieldValue{Present: true, Value: "https://env.example:4318"},
				Protocol: FieldValue{Present: true, Value: "http/protobuf"},
			},
		}
		backend := core_xds.OtelPipeBackend{
			Endpoint: "collector.example:4317",
			EnvPolicy: &core_xds.OtelResolvedEnvPolicy{
				Mode:       motb_api.EnvModeOptional,
				Precedence: motb_api.EnvPrecedenceExplicitFirst,
			},
			Traces: &core_xds.OtelSignalRuntimePlan{
				Enabled:         true,
				EnvInputPresent: true,
			},
		}

		runtime := cfg.ResolveBackend(backend)

		Expect(runtime.Traces.Transport.Protocol).To(Equal(core_xds.OtelProtocolGRPC))
		Expect(runtime.Traces.Transport.Endpoint).To(Equal("collector.example:4317"))
	})

	It("should preserve explicit TLS=false in ExplicitFirst mode when env has https", func() {
		cfg := Config{
			Shared: Layer{
				Endpoint: FieldValue{Present: true, Value: "https://env.example:4318"},
			},
		}
		backend := core_xds.OtelPipeBackend{
			Endpoint: "collector.example:4317",
			UseHTTPS: false,
			UseHTTP:  true,
			EnvPolicy: &core_xds.OtelResolvedEnvPolicy{
				Mode:       motb_api.EnvModeOptional,
				Precedence: motb_api.EnvPrecedenceExplicitFirst,
			},
			Traces: &core_xds.OtelSignalRuntimePlan{
				Enabled:         true,
				EnvInputPresent: true,
			},
		}

		runtime := cfg.ResolveBackend(backend)

		Expect(runtime.Traces.Transport.UseTLS).To(HaveValue(BeFalse()))
	})

	It("should handle bare host:port endpoint without scheme", func() {
		cfg := Config{
			Shared: Layer{
				Endpoint: FieldValue{Present: true, Value: "collector:4317"},
			},
		}
		backend := core_xds.OtelPipeBackend{
			Endpoint: "explicit.example:4317",
			EnvPolicy: &core_xds.OtelResolvedEnvPolicy{
				Mode:       motb_api.EnvModeOptional,
				Precedence: motb_api.EnvPrecedenceEnvFirst,
			},
			Traces: &core_xds.OtelSignalRuntimePlan{
				Enabled:         true,
				EnvInputPresent: true,
			},
		}

		runtime := cfg.ResolveBackend(backend)

		// bare host:port is not a valid URL, so it's used as-is for the endpoint
		Expect(runtime.Traces.Transport.Endpoint).To(Equal("collector:4317"))
	})

	It("should accept compression=none without validation error", func() {
		cfg := Config{
			Shared: Layer{
				Compression: FieldValue{Present: true, Value: "none"},
			},
		}
		backend := core_xds.OtelPipeBackend{
			Endpoint: "collector.example:4317",
			EnvPolicy: &core_xds.OtelResolvedEnvPolicy{
				Mode:       motb_api.EnvModeOptional,
				Precedence: motb_api.EnvPrecedenceEnvFirst,
			},
			Traces: &core_xds.OtelSignalRuntimePlan{
				Enabled:         true,
				EnvInputPresent: true,
			},
		}

		runtime := cfg.ResolveBackend(backend)

		Expect(runtime.Traces.Transport.Compression).To(BeEmpty())
	})
})

var _ = Describe("resolveEndpointAddress", func() {
	It("should resolve empty host using HOST_IP env var", func() {
		os.Setenv("HOST_IP", "10.0.0.5")
		defer os.Unsetenv("HOST_IP")

		Expect(resolveEndpointAddress(":4317")).To(Equal("10.0.0.5:4317"))
	})

	It("should fall back to 127.0.0.1 when HOST_IP is unset", func() {
		os.Unsetenv("HOST_IP")

		Expect(resolveEndpointAddress(":4317")).To(Equal("127.0.0.1:4317"))
	})

	It("should return endpoint unchanged when host is present", func() {
		Expect(resolveEndpointAddress("collector:4317")).To(Equal("collector:4317"))
	})

	It("should return endpoint unchanged when not in host:port format", func() {
		Expect(resolveEndpointAddress("collector")).To(Equal("collector"))
	})
})
