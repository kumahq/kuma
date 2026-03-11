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

var _ = Describe("OTEL status", func() {
	Describe("Build", func() {
		It("should build a sanitized backend summary", func() {
			status := otelstatus.Build([]core_xds.OtelPipeBackend{
				{
					Name:         "main-collector",
					ClientLayout: core_xds.OtelClientLayoutPerSignal,
					EnvPolicy: core_xds.OtelResolvedEnvPolicy{
						Mode:                 motb_api.EnvModeOptional,
						AllowSignalOverrides: true,
					},
					Traces: &core_xds.OtelSignalRuntimePlan{
						Enabled:         true,
						EnvInputPresent: true,
						OverrideKinds:   []string{"endpoint"},
					},
					Logs: &core_xds.OtelSignalRuntimePlan{
						Enabled:         true,
						EnvInputPresent: true,
					},
					Metrics: &core_xds.OtelSignalRuntimePlan{
						Enabled:        true,
						BlockedReasons: []string{core_xds.OtelBlockedReasonMultipleBackends},
					},
				},
			})

			Expect(status).ToNot(BeNil())
			Expect(status.Backends).To(HaveLen(1))
			Expect(status.Backends[0]).To(Equal(&mesh_proto.DataplaneInsight_OpenTelemetry_Backend{
				Name:         "main-collector",
				ClientLayout: "per-signal",
				Traces: &mesh_proto.DataplaneInsight_OpenTelemetry_Signal{
					Enabled:         true,
					EnvAllowed:      true,
					EnvInputPresent: true,
					DedicatedClient: true,
					State:           otelstatus.SignalStateReady,
					OverrideKinds:   []string{"endpoint"},
				},
				Logs: &mesh_proto.DataplaneInsight_OpenTelemetry_Signal{
					Enabled:         true,
					EnvAllowed:      true,
					EnvInputPresent: true,
					State:           otelstatus.SignalStateReady,
				},
				Metrics: &mesh_proto.DataplaneInsight_OpenTelemetry_Signal{
					Enabled:        true,
					EnvAllowed:     true,
					State:          otelstatus.SignalStateAmbiguous,
					BlockedReasons: []string{core_xds.OtelBlockedReasonMultipleBackends},
				},
			}))
		})

		It("should mark policy-disabled env as not allowed", func() {
			status := otelstatus.Build([]core_xds.OtelPipeBackend{
				{
					Name: "main-collector",
					EnvPolicy: core_xds.OtelResolvedEnvPolicy{
						Mode: motb_api.EnvModeDisabled,
					},
					Traces: &core_xds.OtelSignalRuntimePlan{
						Enabled:         true,
						EnvInputPresent: true,
						BlockedReasons:  []string{core_xds.OtelBlockedReasonEnvDisabledByPolicy},
					},
				},
			})

			Expect(status.Backends[0].Traces.EnvAllowed).To(BeFalse())
			Expect(status.Backends[0].Traces.State).To(Equal(otelstatus.SignalStateBlocked))
		})
	})

	Describe("Cache", func() {
		It("should clone values on set and get", func() {
			cache := otelstatus.NewCache()
			key := core_model.ResourceKey{Mesh: "default", Name: "backend-1"}
			value := &mesh_proto.DataplaneInsight_OpenTelemetry{
				Backends: []*mesh_proto.DataplaneInsight_OpenTelemetry_Backend{
					{Name: "main"},
				},
			}

			cache.Set(key, value)
			value.Backends[0].Name = "mutated"

			current := cache.Get(key)
			Expect(current.Backends[0].Name).To(Equal("main"))

			current.Backends[0].Name = "mutated-again"
			Expect(cache.Get(key).Backends[0].Name).To(Equal("main"))
		})

		It("should delete stored status", func() {
			cache := otelstatus.NewCache()
			key := core_model.ResourceKey{Mesh: "default", Name: "backend-1"}
			cache.Set(key, &mesh_proto.DataplaneInsight_OpenTelemetry{})

			cache.Delete(key)

			Expect(cache.Get(key)).To(BeNil())
		})
	})
})
