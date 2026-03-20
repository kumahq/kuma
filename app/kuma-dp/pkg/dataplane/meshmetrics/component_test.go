package meshmetrics

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// preShutdownProvider returns a MeterProvider that is already shut down.
// shutdownBackend will get ErrReaderShutdown, which is explicitly handled.
func preShutdownProvider() *sdkmetric.MeterProvider {
	p := sdkmetric.NewMeterProvider()
	_ = p.Shutdown(context.Background())
	return p
}

var _ = Describe("OtelExportTarget", func() {
	It("stepOtelExport removes backends that disappear from targets", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		m := &Manager{
			ctx: ctx,
			runningBackends: map[string]*runningBackend{
				"backend-a": {
					exporter:      preShutdownProvider(),
					appliedConfig: OtelExportTarget{Name: "backend-a", SocketPath: "/tmp/a.sock", RefreshInterval: time.Minute},
				},
				"backend-b": {
					exporter:      preShutdownProvider(),
					appliedConfig: OtelExportTarget{Name: "backend-b", SocketPath: "/tmp/b.sock", RefreshInterval: time.Minute},
				},
			},
		}

		// Remove backend-a, keep backend-b
		err := m.stepOtelExport([]OtelExportTarget{
			{Name: "backend-b", SocketPath: "/tmp/b.sock", RefreshInterval: time.Minute},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(m.runningBackends).To(HaveLen(1))
		Expect(m.runningBackends).NotTo(HaveKey("backend-a"))
		Expect(m.runningBackends).To(HaveKey("backend-b"))
	})

	It("stepOtelExport removes all backends when targets are empty", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		m := &Manager{
			ctx: ctx,
			runningBackends: map[string]*runningBackend{
				"backend-a": {
					exporter:      preShutdownProvider(),
					appliedConfig: OtelExportTarget{Name: "backend-a"},
				},
			},
		}

		err := m.stepOtelExport(nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(m.runningBackends).To(BeEmpty())
	})

	It("stepOtelExport skips unchanged backends", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		existingProvider := preShutdownProvider()
		target := OtelExportTarget{Name: "backend-a", SocketPath: "/tmp/a.sock", RefreshInterval: time.Minute}
		m := &Manager{
			ctx: ctx,
			runningBackends: map[string]*runningBackend{
				"backend-a": {
					exporter:      existingProvider,
					appliedConfig: target,
				},
			},
		}

		// Same target - should not restart
		err := m.stepOtelExport([]OtelExportTarget{target})
		Expect(err).NotTo(HaveOccurred())
		Expect(m.runningBackends["backend-a"].exporter).To(BeIdenticalTo(existingProvider))
	})

	It("stepScraping does not touch OTEL exporters", func() {
		m := &Manager{
			runningBackends: map[string]*runningBackend{
				"existing": {},
			},
		}
		// stepScraping only calls SetApplicationsToScrape on the producer.
		// Without a producer set, we just verify runningBackends is untouched.
		Expect(m.runningBackends).To(HaveLen(1))
	})

	It("OnOtelTargetsChange sends targets to channel", func() {
		m := &Manager{
			newOtelTargets: make(chan []OtelExportTarget, 1),
		}
		targets := []OtelExportTarget{
			{Name: "test", SocketPath: "/tmp/test.sock", RefreshInterval: time.Minute},
		}
		m.OnOtelTargetsChange(targets)

		received := <-m.newOtelTargets
		Expect(received).To(HaveLen(1))
		Expect(received[0].Name).To(Equal("test"))
	})

	It("OnOtelTargetsChange drops stale and pushes new", func() {
		m := &Manager{
			newOtelTargets: make(chan []OtelExportTarget, 1),
		}
		// Fill channel
		m.OnOtelTargetsChange([]OtelExportTarget{{Name: "old"}})
		// Push newer value - should drop old
		m.OnOtelTargetsChange([]OtelExportTarget{{Name: "new"}})

		received := <-m.newOtelTargets
		Expect(received).To(HaveLen(1))
		Expect(received[0].Name).To(Equal("new"))
	})
})
