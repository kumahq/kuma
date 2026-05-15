package sync

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
	otelstatus "github.com/kumahq/kuma/v2/pkg/xds/otel/status"
)

var _ = Describe("DataplaneWatchdog syncOtelStatus", func() {
	var (
		key      core_model.ResourceKey
		cache    *otelstatus.Cache
		watchdog *DataplaneWatchdog
	)

	BeforeEach(func() {
		key = core_model.ResourceKey{Mesh: "default", Name: "example-001"}
		cache = otelstatus.NewCache()
		watchdog = &DataplaneWatchdog{
			DataplaneWatchdogDependencies: DataplaneWatchdogDependencies{
				OtelStatusCache: cache,
			},
			key: key,
		}
	})

	otelMain := func() *core_xds.OtelPipeBackends {
		backends := &core_xds.OtelPipeBackends{}
		backends.AddSignal(
			"otel-main",
			core_xds.OtelPipeBackend{
				SocketPath: "/tmp/otel-main.sock",
				Endpoint:   "otel-main:4317",
			},
			core_xds.OtelSignalTraces,
			core_xds.OtelSignalRuntimePlan{
				Enabled:         true,
				EnvInputPresent: true,
			},
		)
		return backends
	}

	It("should write to cache on first sync", func() {
		Expect(watchdog.syncOtelStatus(otelMain())).To(BeTrue())

		got := cache.Get(key)
		Expect(got).NotTo(BeNil())
		Expect(got.Backends).To(HaveLen(1))
		Expect(got.Backends[0].Name).To(Equal("otel-main"))
		Expect(got.Backends[0].Traces).NotTo(BeNil())
		Expect(got.Backends[0].Traces.State).To(Equal(otelstatus.SignalStateReady))
	})

	It("should skip when backends are unchanged", func() {
		watchdog.syncOtelStatus(otelMain())
		Expect(watchdog.syncOtelStatus(otelMain())).To(BeFalse())
	})

	It("should clear cache on transition to nil", func() {
		watchdog.syncOtelStatus(otelMain())
		Expect(watchdog.syncOtelStatus(nil)).To(BeTrue())
		Expect(cache.Get(key)).To(BeNil())
	})

	It("should skip repeated nil syncs", func() {
		watchdog.syncOtelStatus(otelMain())
		watchdog.syncOtelStatus(nil)
		Expect(watchdog.syncOtelStatus(nil)).To(BeFalse())
	})

	It("should detect changed backend state", func() {
		watchdog.syncOtelStatus(otelMain())

		other := &core_xds.OtelPipeBackends{}
		other.AddSignal(
			"otel-main",
			core_xds.OtelPipeBackend{
				SocketPath: "/tmp/otel-main.sock",
				Endpoint:   "otel-main:4317",
			},
			core_xds.OtelSignalTraces,
			core_xds.OtelSignalRuntimePlan{
				Enabled:         true,
				EnvInputPresent: false,
				MissingFields:   []string{"endpoint"},
			},
		)

		Expect(watchdog.syncOtelStatus(other)).To(BeTrue())

		got := cache.Get(key)
		Expect(got).NotTo(BeNil())
		Expect(got.Backends).To(HaveLen(1))
		Expect(got.Backends[0].Name).To(Equal("otel-main"))
		Expect(got.Backends[0].Traces).NotTo(BeNil())
		Expect(got.Backends[0].Traces.State).To(Equal(otelstatus.SignalStateMissing))
	})
})
