package xds_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	motb_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshopentelemetrybackend/api/v1alpha1"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
)

var _ = Describe("OtelPipeBackends", func() {
	baseBackend := func(name string) core_xds.OtelPipeBackend {
		return core_xds.OtelPipeBackend{
			Name:       name,
			SocketPath: "/tmp/" + name + ".sock",
			Endpoint:   name + ":4317",
			EnvPolicy: core_xds.OtelResolvedEnvPolicy{
				Mode:                 motb_api.EnvModeOptional,
				Precedence:           motb_api.EnvPrecedenceEnvFirst,
				AllowSignalOverrides: true,
			},
		}
	}

	It("should use per-signal layout when multiple signals diverge", func() {
		backends := &core_xds.OtelPipeBackends{}
		backend := baseBackend("collector")

		backends.AddSignal("collector", backend, core_xds.OtelSignalTraces, core_xds.OtelSignalRuntimePlan{
			Enabled:         true,
			EnvInputPresent: true,
			OverrideKinds:   []string{"endpoint"},
		})
		backends.AddSignal("collector", backend, core_xds.OtelSignalLogs, core_xds.OtelSignalRuntimePlan{
			Enabled:         true,
			EnvInputPresent: true,
		})

		all := backends.All()
		Expect(all).To(HaveLen(1))
		Expect(all[0].ClientLayout).To(Equal(core_xds.OtelClientLayoutPerSignal))
	})

	It("should keep shared layout when only one signal is enabled", func() {
		backends := &core_xds.OtelPipeBackends{}
		backend := baseBackend("collector")

		backends.AddSignal("collector", backend, core_xds.OtelSignalTraces, core_xds.OtelSignalRuntimePlan{
			Enabled:         true,
			EnvInputPresent: true,
			OverrideKinds:   []string{"endpoint"},
		})

		all := backends.All()
		Expect(all).To(HaveLen(1))
		Expect(all[0].ClientLayout).To(Equal(core_xds.OtelClientLayoutShared))
	})

	It("should mark ambiguity when one signal uses env on multiple backends", func() {
		backends := &core_xds.OtelPipeBackends{}
		backendA := baseBackend("collector-a")
		backendB := baseBackend("collector-b")

		backends.AddSignal("collector-a", backendA, core_xds.OtelSignalTraces, core_xds.OtelSignalRuntimePlan{
			Enabled:         true,
			EnvInputPresent: true,
		})
		backends.AddSignal("collector-b", backendB, core_xds.OtelSignalTraces, core_xds.OtelSignalRuntimePlan{
			Enabled:         true,
			EnvInputPresent: true,
		})

		all := backends.All()
		Expect(all).To(HaveLen(2))
		Expect(all[0].Traces.BlockedReasons).To(ContainElement(core_xds.OtelBlockedReasonMultipleBackends))
		Expect(all[1].Traces.BlockedReasons).To(ContainElement(core_xds.OtelBlockedReasonMultipleBackends))
	})
})
