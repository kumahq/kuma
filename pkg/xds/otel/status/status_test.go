package status_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	motb_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshopentelemetrybackend/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
	otelstatus "github.com/kumahq/kuma/v2/pkg/xds/otel/status"
)

var _ = Describe("OTel Status Cache", func() {
	It("should store and retrieve status", func() {
		cache := otelstatus.NewCache()
		key := core_model.ResourceKey{Mesh: "default", Name: "dp-1"}

		status := &mesh_proto.DataplaneInsight_OpenTelemetry{
			Backends: []*mesh_proto.DataplaneInsight_OpenTelemetry_Backend{
				{Name: "otel-main"},
			},
		}

		cache.Set(key, status)
		got := cache.Get(key)
		Expect(got).ToNot(BeNil())
		Expect(got.Backends).To(HaveLen(1))
		Expect(got.Backends[0].Name).To(Equal("otel-main"))
	})

	It("should return nil for missing key", func() {
		cache := otelstatus.NewCache()
		got := cache.Get(core_model.ResourceKey{Mesh: "default", Name: "missing"})
		Expect(got).To(BeNil())
	})

	It("should delete on nil status", func() {
		cache := otelstatus.NewCache()
		key := core_model.ResourceKey{Mesh: "default", Name: "dp-1"}

		cache.Set(key, &mesh_proto.DataplaneInsight_OpenTelemetry{})
		cache.Set(key, nil)
		Expect(cache.Get(key)).To(BeNil())
	})

	It("should isolate cached state from caller mutations", func() {
		cache := otelstatus.NewCache()
		key := core_model.ResourceKey{Mesh: "default", Name: "dp-1"}
		status := &mesh_proto.DataplaneInsight_OpenTelemetry{
			Backends: []*mesh_proto.DataplaneInsight_OpenTelemetry_Backend{
				{Name: "otel-main"},
			},
		}

		cache.Set(key, status)
		status.Backends[0].Name = "mutated-input"

		got := cache.Get(key)
		Expect(got.Backends[0].Name).To(Equal("otel-main"))

		got.Backends[0].Name = "mutated-output"
		Expect(cache.Get(key).Backends[0].Name).To(Equal("otel-main"))
	})

	It("should be safe on nil cache", func() {
		var cache *otelstatus.Cache
		cache.Set(core_model.ResourceKey{}, nil)
		Expect(cache.Get(core_model.ResourceKey{})).To(BeNil())
	})
})

var _ = Describe("Build", func() {
	It("should return nil for empty backends", func() {
		Expect(otelstatus.Build(nil)).To(BeNil())
	})

	It("should build status from backends", func() {
		backends := []core_xds.OtelPipeBackend{
			{
				Name: "otel-main",
				EnvPolicy: &core_xds.OtelResolvedEnvPolicy{
					Mode: motb_api.EnvModeOptional,
				},
				Traces: &core_xds.OtelSignalRuntimePlan{
					Enabled:         true,
					EnvInputPresent: true,
				},
				Logs: &core_xds.OtelSignalRuntimePlan{
					Enabled: true,
				},
				Metrics: &core_xds.OtelSignalRuntimePlan{
					Enabled:         true,
					EnvInputPresent: false,
				},
			},
		}

		result := otelstatus.Build(backends)
		Expect(result).ToNot(BeNil())
		Expect(result.Backends).To(HaveLen(1))

		b := result.Backends[0]
		Expect(b.Name).To(Equal("otel-main"))
		Expect(b.Traces.State).To(Equal(otelstatus.SignalStateReady))
		Expect(b.Traces.EnvAllowed).To(BeTrue())
		Expect(b.Traces.EnvInputPresent).To(BeTrue())
		Expect(b.Logs.State).To(Equal(otelstatus.SignalStateReady))
		Expect(b.Metrics.State).To(Equal(otelstatus.SignalStateReady))
		Expect(b.Metrics.EnvInputPresent).To(BeFalse())
	})

	It("should report ambiguous state for multiple backends", func() {
		backends := []core_xds.OtelPipeBackend{
			{
				Name: "otel-1",
				Traces: &core_xds.OtelSignalRuntimePlan{
					Enabled:        true,
					BlockedReasons: []string{core_xds.OtelBlockedReasonMultipleBackends},
				},
			},
		}
		result := otelstatus.Build(backends)
		Expect(result.Backends[0].Traces.State).To(Equal(otelstatus.SignalStateAmbiguous))
	})

	It("should report missing state for required env missing", func() {
		backends := []core_xds.OtelPipeBackend{
			{
				Name: "otel-1",
				Traces: &core_xds.OtelSignalRuntimePlan{
					Enabled:        true,
					BlockedReasons: []string{core_xds.OtelBlockedReasonRequiredEnvMissing},
					MissingFields:  []string{"endpoint"},
				},
			},
		}
		result := otelstatus.Build(backends)
		Expect(result.Backends[0].Traces.State).To(Equal(otelstatus.SignalStateMissing))
		Expect(result.Backends[0].Traces.MissingFields).To(ContainElement("endpoint"))
	})

	It("should pick highest-precedence state when multiple blocked reasons coexist", func() {
		backends := []core_xds.OtelPipeBackend{
			{
				Name: "otel-1",
				Traces: &core_xds.OtelSignalRuntimePlan{
					Enabled:        true,
					BlockedReasons: []string{core_xds.OtelBlockedReasonMultipleBackends, core_xds.OtelBlockedReasonRequiredEnvMissing},
					MissingFields:  []string{"endpoint"},
				},
			},
		}
		result := otelstatus.Build(backends)
		// ambiguous takes precedence over missing
		Expect(result.Backends[0].Traces.State).To(Equal(otelstatus.SignalStateAmbiguous))
	})

	It("should report missing when RequiredEnvMissing without MultipleBackends", func() {
		backends := []core_xds.OtelPipeBackend{
			{
				Name: "otel-1",
				Traces: &core_xds.OtelSignalRuntimePlan{
					Enabled:        true,
					BlockedReasons: []string{core_xds.OtelBlockedReasonRequiredEnvMissing, core_xds.OtelBlockedReasonEnvDisabledByPolicy},
					MissingFields:  []string{"endpoint"},
				},
			},
		}
		result := otelstatus.Build(backends)
		// missing takes precedence over soft blocks
		Expect(result.Backends[0].Traces.State).To(Equal(otelstatus.SignalStateMissing))
	})

	It("should report env not allowed when policy disables it", func() {
		backends := []core_xds.OtelPipeBackend{
			{
				Name: "otel-1",
				EnvPolicy: &core_xds.OtelResolvedEnvPolicy{
					Mode: motb_api.EnvModeDisabled,
				},
				Traces: &core_xds.OtelSignalRuntimePlan{
					Enabled:        true,
					BlockedReasons: []string{core_xds.OtelBlockedReasonEnvDisabledByPolicy},
				},
			},
		}
		result := otelstatus.Build(backends)
		Expect(result.Backends[0].Traces.EnvAllowed).To(BeFalse())
		Expect(result.Backends[0].Traces.State).To(Equal(otelstatus.SignalStateReady))
	})

	It("should default env allowed when env policy is nil", func() {
		backends := []core_xds.OtelPipeBackend{
			{
				Name: "otel-1",
				Traces: &core_xds.OtelSignalRuntimePlan{
					Enabled: true,
				},
			},
		}
		result := otelstatus.Build(backends)
		Expect(result.Backends[0].Traces.EnvAllowed).To(BeTrue())
	})
})
