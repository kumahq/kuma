package xds_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
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
			nil, "", dummyParser, dummyNamer, emptyResources,
		)
		Expect(result).To(BeNil())
	})

	Describe("priority order", func() {
		backendRef := &common_api.BackendResourceRef{
			Kind: common_api.BackendResourceMeshOpenTelemetryBackend,
			Name: "my-backend",
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
				backendRef, "inline-collector:4317", dummyParser, dummyNamer, resources,
			)
			// Dangling backendRef returns nil, does NOT fall through
			Expect(result).To(BeNil())
		})

		It("should resolve inline endpoint when no backendRef", func() {
			result := policies_xds.ResolveOtelBackend(
				nil, "inline-collector:4317", dummyParser, dummyNamer, emptyResources,
			)
			Expect(result).ToNot(BeNil())
			Expect(result.Endpoint.Target).To(Equal("inline-collector:4317"))
			Expect(result.Protocol).To(Equal(motb_api.ProtocolGRPC))
		})
	})

	Describe("inline endpoint uses ParseOtelEndpoint", func() {
		It("should resolve via ParseOtelEndpoint when no backendRef", func() {
			result := policies_xds.ResolveOtelBackend(
				nil, "collector:4317", policies_xds.ParseOtelEndpoint, func(ep string) string { return ep }, emptyResources,
			)
			Expect(result).ToNot(BeNil())
			Expect(result.Endpoint.Target).To(Equal("collector"))
			Expect(result.Endpoint.Port).To(Equal(uint32(4317)))
		})
	})

	Describe("optional endpoint fields", func() {
		makeResources := func(backend *motb_api.MeshOpenTelemetryBackendResource) xds_context.Resources {
			list := &motb_api.MeshOpenTelemetryBackendResourceList{Items: []*motb_api.MeshOpenTelemetryBackendResource{backend}}
			return xds_context.Resources{
				MeshLocalResources: map[core_model.ResourceType]core_model.ResourceList{
					motb_api.MeshOpenTelemetryBackendType: list,
				},
			}
		}
		backendRef := &common_api.BackendResourceRef{
			Kind: common_api.BackendResourceMeshOpenTelemetryBackend,
			Name: "collector",
		}

		It("should default to port 4317 and empty address when endpoint is nil", func() {
			backend := motb_api.NewMeshOpenTelemetryBackendResource()
			backend.SetMeta(&test_model.ResourceMeta{Name: "collector", Mesh: "default"})
			backend.Spec.Protocol = pointer.To(motb_api.ProtocolGRPC)

			result := policies_xds.ResolveOtelBackend(
				backendRef, "", dummyParser, dummyNamer, makeResources(backend),
			)
			Expect(result).ToNot(BeNil())
			Expect(result.Endpoint.Target).To(BeEmpty())
			Expect(result.Endpoint.Port).To(Equal(uint32(4317)))
		})

		It("should default to port 4317 when only address is set", func() {
			backend := motb_api.NewMeshOpenTelemetryBackendResource()
			backend.SetMeta(&test_model.ResourceMeta{Name: "collector", Mesh: "default"})
			backend.Spec.Endpoint = &motb_api.Endpoint{Address: pointer.To("collector.example")}
			backend.Spec.Protocol = pointer.To(motb_api.ProtocolGRPC)

			result := policies_xds.ResolveOtelBackend(
				backendRef, "", dummyParser, dummyNamer, makeResources(backend),
			)
			Expect(result).ToNot(BeNil())
			Expect(result.Endpoint.Target).To(Equal("collector.example"))
			Expect(result.Endpoint.Port).To(Equal(uint32(4317)))
		})

		It("should use empty address when only port is set", func() {
			backend := motb_api.NewMeshOpenTelemetryBackendResource()
			backend.SetMeta(&test_model.ResourceMeta{Name: "collector", Mesh: "default"})
			backend.Spec.Endpoint = &motb_api.Endpoint{Port: pointer.To(int32(4318))}
			backend.Spec.Protocol = pointer.To(motb_api.ProtocolGRPC)

			result := policies_xds.ResolveOtelBackend(
				backendRef, "", dummyParser, dummyNamer, makeResources(backend),
			)
			Expect(result).ToNot(BeNil())
			Expect(result.Endpoint.Target).To(BeEmpty())
			Expect(result.Endpoint.Port).To(Equal(uint32(4318)))
		})
	})

	Describe("HTTP backend transport", func() {
		makeResources := func(backends ...*motb_api.MeshOpenTelemetryBackendResource) xds_context.Resources {
			list := &motb_api.MeshOpenTelemetryBackendResourceList{Items: backends}
			return xds_context.Resources{
				MeshLocalResources: map[core_model.ResourceType]core_model.ResourceList{
					motb_api.MeshOpenTelemetryBackendType: list,
				},
			}
		}

		It("should enable HTTPS for HTTP protocol on port 443", func() {
			backend := motb_api.NewMeshOpenTelemetryBackendResource()
			backend.SetMeta(&test_model.ResourceMeta{Name: "https-collector", Mesh: "default"})
			backend.Spec.Endpoint = &motb_api.Endpoint{Address: pointer.To("collector.example"), Port: pointer.To(int32(443))}
			backend.Spec.Protocol = pointer.To(motb_api.ProtocolHTTP)

			result := policies_xds.ResolveOtelBackend(
				&common_api.BackendResourceRef{
					Kind: common_api.BackendResourceMeshOpenTelemetryBackend,
					Name: "https-collector",
				},
				"",
				dummyParser,
				dummyNamer,
				makeResources(backend),
			)
			Expect(result).ToNot(BeNil())
			Expect(result.UseHTTPS).To(BeTrue())
		})

		It("should not enable HTTPS for HTTP protocol on non-443 port", func() {
			backend := motb_api.NewMeshOpenTelemetryBackendResource()
			backend.SetMeta(&test_model.ResourceMeta{Name: "http-collector", Mesh: "default"})
			backend.Spec.Endpoint = &motb_api.Endpoint{Address: pointer.To("collector.example"), Port: pointer.To(int32(4318))}
			backend.Spec.Protocol = pointer.To(motb_api.ProtocolHTTP)

			result := policies_xds.ResolveOtelBackend(
				&common_api.BackendResourceRef{
					Kind: common_api.BackendResourceMeshOpenTelemetryBackend,
					Name: "http-collector",
				},
				"",
				dummyParser,
				dummyNamer,
				makeResources(backend),
			)
			Expect(result).ToNot(BeNil())
			Expect(result.UseHTTPS).To(BeFalse())
		})

		It("should resolve by labels and pick oldest on collision", func() {
			now := time.Now()

			backendA := motb_api.NewMeshOpenTelemetryBackendResource()
			backendA.SetMeta(&test_model.ResourceMeta{
				Name:         "collector-a.kuma-system",
				Mesh:         "default",
				CreationTime: now.Add(-2 * time.Hour),
				Labels: map[string]string{
					mesh_proto.DisplayName: "collector",
				},
			})
			backendA.Spec.Endpoint = &motb_api.Endpoint{Address: pointer.To("collector-a"), Port: pointer.To(int32(4317))}

			backendB := motb_api.NewMeshOpenTelemetryBackendResource()
			backendB.SetMeta(&test_model.ResourceMeta{
				Name:         "collector-b.kuma-system",
				Mesh:         "default",
				CreationTime: now.Add(-1 * time.Hour),
				Labels: map[string]string{
					mesh_proto.DisplayName: "collector",
				},
			})
			backendB.Spec.Endpoint = &motb_api.Endpoint{Address: pointer.To("collector-b"), Port: pointer.To(int32(4317))}

			result := policies_xds.ResolveOtelBackend(
				&common_api.BackendResourceRef{
					Kind: common_api.BackendResourceMeshOpenTelemetryBackend,
					Labels: map[string]string{
						mesh_proto.DisplayName: "collector",
					},
				},
				"",
				dummyParser,
				dummyNamer,
				makeResources(backendA, backendB),
			)
			Expect(result).ToNot(BeNil())
			Expect(result.Endpoint.Target).To(Equal("collector-a"))
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

var _ = Describe("EndpointForDirectOtelExport", func() {
	It("should enable TLS transport for HTTPS-resolved HTTP backend", func() {
		resolved := &policies_xds.ResolvedOtelBackend{
			Endpoint: &core_xds.Endpoint{
				Target: "collector.example",
				Port:   443,
			},
			UseHTTPS: true,
		}

		ep := policies_xds.EndpointForDirectOtelExport(resolved, "")
		Expect(ep).ToNot(BeNil())
		Expect(ep.ExternalService).ToNot(BeNil())
		Expect(ep.ExternalService.TLSEnabled).To(BeTrue())
		Expect(ep.ExternalService.FallbackToSystemCa).To(BeTrue())
	})

	It("should fill in nodeHostIP when target is empty", func() {
		resolved := &policies_xds.ResolvedOtelBackend{
			Endpoint: &core_xds.Endpoint{
				Target: "",
				Port:   4317,
			},
		}

		ep := policies_xds.EndpointForDirectOtelExport(resolved, "192.168.1.5")
		Expect(ep).ToNot(BeNil())
		Expect(ep.Target).To(Equal("192.168.1.5"))
		Expect(ep.Port).To(Equal(uint32(4317)))
	})

	It("should fall back to 127.0.0.1 when target and nodeHostIP are both empty", func() {
		resolved := &policies_xds.ResolvedOtelBackend{
			Endpoint: &core_xds.Endpoint{
				Target: "",
				Port:   4317,
			},
		}

		ep := policies_xds.EndpointForDirectOtelExport(resolved, "")
		Expect(ep).ToNot(BeNil())
		Expect(ep.Target).To(Equal("127.0.0.1"))
	})

	It("should not override existing target", func() {
		resolved := &policies_xds.ResolvedOtelBackend{
			Endpoint: &core_xds.Endpoint{
				Target: "collector.example",
				Port:   4317,
			},
		}

		ep := policies_xds.EndpointForDirectOtelExport(resolved, "192.168.1.5")
		Expect(ep).ToNot(BeNil())
		Expect(ep.Target).To(Equal("collector.example"))
	})
})

var _ = Describe("BuildSignalRuntimePlan", func() {
	It("should mark signal overrides as blocked when policy disallows them", func() {
		plan := policies_xds.BuildSignalRuntimePlan(
			&core_xds.OtelBootstrapInventory{
				Shared: &core_xds.OtelSignalEnvInventory{
					EndpointPresent: true,
				},
				Traces: &core_xds.OtelSignalEnvInventory{
					OverrideKinds: []string{"endpoint"},
				},
			},
			core_xds.OtelResolvedEnvPolicy{
				Mode:                 motb_api.EnvModeOptional,
				Precedence:           motb_api.EnvPrecedenceEnvFirst,
				AllowSignalOverrides: false,
			},
			core_xds.OtelPipeBackend{Endpoint: "collector:4317", UseHTTP: true},
			core_xds.OtelSignalTraces,
			policies_xds.AddResolvedBackendOptions{},
		)

		Expect(plan.Enabled).To(BeTrue())
		Expect(plan.EnvInputPresent).To(BeTrue())
		Expect(plan.OverrideKinds).To(ConsistOf("endpoint"))
		Expect(plan.BlockedReasons).To(ContainElement(core_xds.OtelBlockedReasonSignalOverridesBlocked))
	})

	It("should mark required env as missing when inventory is absent", func() {
		plan := policies_xds.BuildSignalRuntimePlan(
			nil,
			core_xds.OtelResolvedEnvPolicy{
				Mode:                 motb_api.EnvModeRequired,
				Precedence:           motb_api.EnvPrecedenceEnvFirst,
				AllowSignalOverrides: true,
			},
			core_xds.OtelPipeBackend{Endpoint: "collector:4317"},
			core_xds.OtelSignalLogs,
			policies_xds.AddResolvedBackendOptions{},
		)

		Expect(plan.Enabled).To(BeTrue())
		Expect(plan.EnvInputPresent).To(BeFalse())
		Expect(plan.BlockedReasons).To(ContainElement(core_xds.OtelBlockedReasonRequiredEnvMissing))
	})

	It("should mark invalid required shared env fields as missing", func() {
		plan := policies_xds.BuildSignalRuntimePlan(
			&core_xds.OtelBootstrapInventory{
				Shared: &core_xds.OtelSignalEnvInventory{
					ProtocolPresent:   true,
					EffectiveProtocol: core_xds.OtelProtocolUnknown,
				},
				ValidationErrors: []string{"shared.protocol"},
			},
			core_xds.OtelResolvedEnvPolicy{
				Mode:                 motb_api.EnvModeRequired,
				Precedence:           motb_api.EnvPrecedenceEnvFirst,
				AllowSignalOverrides: true,
			},
			core_xds.OtelPipeBackend{Endpoint: "collector:4317"},
			core_xds.OtelSignalMetrics,
			policies_xds.AddResolvedBackendOptions{},
		)

		Expect(plan.EnvInputPresent).To(BeTrue())
		Expect(plan.MissingFields).To(ConsistOf("protocol"))
		Expect(plan.BlockedReasons).To(ContainElement(core_xds.OtelBlockedReasonRequiredEnvMissing))
	})

	It("should mark invalid required timeout and compression as missing", func() {
		plan := policies_xds.BuildSignalRuntimePlan(
			&core_xds.OtelBootstrapInventory{
				Shared: &core_xds.OtelSignalEnvInventory{
					TimeoutPresent:     true,
					CompressionPresent: true,
				},
				ValidationErrors: []string{"shared.timeout", "shared.compression"},
			},
			core_xds.OtelResolvedEnvPolicy{
				Mode:                 motb_api.EnvModeRequired,
				Precedence:           motb_api.EnvPrecedenceEnvFirst,
				AllowSignalOverrides: true,
			},
			core_xds.OtelPipeBackend{Endpoint: "collector:4317"},
			core_xds.OtelSignalMetrics,
			policies_xds.AddResolvedBackendOptions{},
		)

		Expect(plan.EnvInputPresent).To(BeTrue())
		Expect(plan.MissingFields).To(ConsistOf("timeout", "compression"))
		Expect(plan.BlockedReasons).To(ContainElement(core_xds.OtelBlockedReasonRequiredEnvMissing))
	})

	It("should mark invalid required signal mTLS input as missing", func() {
		plan := policies_xds.BuildSignalRuntimePlan(
			&core_xds.OtelBootstrapInventory{
				Traces: &core_xds.OtelSignalEnvInventory{
					ClientCertificatePresent: true,
				},
				ValidationErrors: []string{"traces.mtls"},
			},
			core_xds.OtelResolvedEnvPolicy{
				Mode:                 motb_api.EnvModeRequired,
				Precedence:           motb_api.EnvPrecedenceEnvFirst,
				AllowSignalOverrides: true,
			},
			core_xds.OtelPipeBackend{Endpoint: "collector:4317"},
			core_xds.OtelSignalTraces,
			policies_xds.AddResolvedBackendOptions{},
		)

		Expect(plan.EnvInputPresent).To(BeTrue())
		Expect(plan.MissingFields).To(ConsistOf("client_certificate", "client_key"))
		Expect(plan.BlockedReasons).To(ContainElement(core_xds.OtelBlockedReasonRequiredEnvMissing))
	})

	It("should ignore invalid optional env fields in the runtime plan", func() {
		plan := policies_xds.BuildSignalRuntimePlan(
			&core_xds.OtelBootstrapInventory{
				Shared: &core_xds.OtelSignalEnvInventory{
					ProtocolPresent:   true,
					EffectiveProtocol: core_xds.OtelProtocolUnknown,
				},
				ValidationErrors: []string{"shared.protocol"},
			},
			core_xds.OtelResolvedEnvPolicy{
				Mode:                 motb_api.EnvModeOptional,
				Precedence:           motb_api.EnvPrecedenceEnvFirst,
				AllowSignalOverrides: true,
			},
			core_xds.OtelPipeBackend{Endpoint: "collector:4317"},
			core_xds.OtelSignalLogs,
			policies_xds.AddResolvedBackendOptions{},
		)

		Expect(plan.EnvInputPresent).To(BeTrue())
		Expect(plan.MissingFields).To(BeEmpty())
		Expect(plan.BlockedReasons).To(BeEmpty())
	})

	It("should mark disabled env mode with env input as blocked by policy", func() {
		plan := policies_xds.BuildSignalRuntimePlan(
			&core_xds.OtelBootstrapInventory{
				Shared: &core_xds.OtelSignalEnvInventory{
					EndpointPresent:   true,
					ProtocolPresent:   true,
					EffectiveProtocol: core_xds.OtelProtocolGRPC,
				},
			},
			core_xds.OtelResolvedEnvPolicy{
				Mode:                 motb_api.EnvModeDisabled,
				Precedence:           motb_api.EnvPrecedenceEnvFirst,
				AllowSignalOverrides: true,
			},
			core_xds.OtelPipeBackend{Endpoint: "collector:4317"},
			core_xds.OtelSignalTraces,
			policies_xds.AddResolvedBackendOptions{},
		)

		Expect(plan.EnvInputPresent).To(BeTrue())
		Expect(plan.BlockedReasons).To(ContainElement(core_xds.OtelBlockedReasonEnvDisabledByPolicy))
	})

	It("should flag ambiguity when multiple backends reuse env input for one signal", func() {
		backends := core_xds.OtelPipeBackends{}
		base := core_xds.OtelPipeBackend{
			Name:       "collector",
			Endpoint:   "collector:4317",
			SocketPath: "/tmp/collector.sock",
			EnvPolicy: core_xds.OtelResolvedEnvPolicy{
				Mode:                 motb_api.EnvModeOptional,
				Precedence:           motb_api.EnvPrecedenceEnvFirst,
				AllowSignalOverrides: true,
			},
		}

		backends.AddSignal("collector", base, core_xds.OtelSignalTraces, core_xds.OtelSignalRuntimePlan{
			Enabled:         true,
			EnvInputPresent: true,
		})
		backends.AddSignal("second", core_xds.OtelPipeBackend{
			Name:       "second",
			Endpoint:   "second:4317",
			SocketPath: "/tmp/second.sock",
			EnvPolicy:  base.EnvPolicy,
		}, core_xds.OtelSignalTraces, core_xds.OtelSignalRuntimePlan{
			Enabled:         true,
			EnvInputPresent: true,
		})

		all := backends.All()
		Expect(all).To(HaveLen(2))
		Expect(all[0].Traces.BlockedReasons).To(ContainElement(core_xds.OtelBlockedReasonMultipleBackends))
		Expect(all[1].Traces.BlockedReasons).To(ContainElement(core_xds.OtelBlockedReasonMultipleBackends))
	})

	It("should merge signal plans when two plugins add the same backend", func() {
		backends := core_xds.OtelPipeBackends{}
		base := core_xds.OtelPipeBackend{
			Name:       "collector",
			Endpoint:   "collector:4317",
			SocketPath: "/tmp/collector.sock",
			EnvPolicy: core_xds.OtelResolvedEnvPolicy{
				Mode:                 motb_api.EnvModeOptional,
				Precedence:           motb_api.EnvPrecedenceEnvFirst,
				AllowSignalOverrides: true,
			},
		}

		backends.AddSignal("collector", base, core_xds.OtelSignalTraces, core_xds.OtelSignalRuntimePlan{
			Enabled:         true,
			EnvInputPresent: true,
		})
		backends.AddSignal("collector", base, core_xds.OtelSignalMetrics, core_xds.OtelSignalRuntimePlan{
			Enabled:         true,
			EnvInputPresent: true,
		})

		all := backends.All()
		Expect(all).To(HaveLen(1))
		Expect(all[0].Name).To(Equal("collector"))
		Expect(all[0].Traces).ToNot(BeNil())
		Expect(all[0].Traces.Enabled).To(BeTrue())
		Expect(all[0].Metrics).ToNot(BeNil())
		Expect(all[0].Metrics.Enabled).To(BeTrue())
	})
})

var _ = Describe("OtelEnvPlanningEnabled", func() {
	It("should return false when proxy is nil", func() {
		ctx := xds_context.Context{}
		Expect(policies_xds.OtelEnvPlanningEnabled(ctx, nil)).To(BeFalse())
	})

	It("should return false when proxy metadata is nil", func() {
		ctx := xds_context.Context{}
		proxy := &core_xds.Proxy{}
		Expect(policies_xds.OtelEnvPlanningEnabled(ctx, proxy)).To(BeFalse())
	})

	It("should return false when proxy has no inventory", func() {
		ctx := xds_context.Context{}
		proxy := &core_xds.Proxy{
			Metadata: &core_xds.DataplaneMetadata{},
		}
		Expect(policies_xds.OtelEnvPlanningEnabled(ctx, proxy)).To(BeFalse())
	})

	It("should return true when proxy reports OTEL inventory", func() {
		ctx := xds_context.Context{}
		proxy := &core_xds.Proxy{
			Metadata: &core_xds.DataplaneMetadata{
				OtelEnvInventory: &core_xds.OtelBootstrapInventory{},
			},
		}
		Expect(policies_xds.OtelEnvPlanningEnabled(ctx, proxy)).To(BeTrue())
	})
})
