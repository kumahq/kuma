package sync_test

import (
	"context"
	"sync/atomic"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	util_watchdog "github.com/Kong/kuma/pkg/util/watchdog"

	. "github.com/Kong/kuma/pkg/xds/sync"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
)

var _ = Describe("Sync", func() {
	Describe("dataplaneSyncTracker", func() {
		It("should not fail when ADS stream is closed before Watchdog is even created", func() {
			// setup
			tracker := NewDataplaneSyncTracker(nil)

			// given
			ctx := context.Background()
			streamID := int64(1)
			typ := ""

			By("simulating Envoy connecting to the Control Plane")
			// when
			err := tracker.OnStreamOpen(ctx, streamID, typ)
			// then
			Expect(err).ToNot(HaveOccurred())

			By("simulating Envoy disconnecting from the Control Plane prior to any DiscoveryRequest")
			// and
			tracker.OnStreamClosed(streamID)

			// then
			// expect no panic
		})

		It("should not fail when Envoy presents invalid Node ID", func() {
			// setup
			tracker := NewDataplaneSyncTracker(nil)

			// given
			ctx := context.Background()
			streamID := int64(1)
			typ := ""
			req := &envoy.DiscoveryRequest{}

			By("simulating Envoy connecting to the Control Plane")
			// when
			err := tracker.OnStreamOpen(ctx, streamID, typ)
			// then
			Expect(err).ToNot(HaveOccurred())

			By("simulating DiscoveryRequest")
			// when
			err = tracker.OnStreamRequest(streamID, req)
			// then
			Expect(err).ToNot(HaveOccurred())

			By("simulating Envoy disconnecting from the Control Plane")
			// and
			tracker.OnStreamClosed(streamID)

			// then
			// expect no panic
		})

		It("should create a Watchdog when Envoy presents a valid Node ID", func(done Done) {
			watchdogCh := make(chan core_model.ResourceKey)

			// setup
			tracker := NewDataplaneSyncTracker(NewDataplaneWatchdogFunc(func(dataplaneId core_model.ResourceKey, streamId int64) util_watchdog.Watchdog {
				return WatchdogFunc(func(stop <-chan struct{}) {
					watchdogCh <- dataplaneId
					<-stop
					close(watchdogCh)
				})
			}))

			// given
			ctx := context.Background()
			streamID := int64(1)
			typ := ""
			req := &envoy.DiscoveryRequest{
				Node: &envoy_core.Node{
					Id: "demo.example",
				},
			}

			By("simulating Envoy connecting to the Control Plane")
			// when
			err := tracker.OnStreamOpen(ctx, streamID, typ)
			// then
			Expect(err).ToNot(HaveOccurred())

			By("simulating DiscoveryRequest")
			// when
			err = tracker.OnStreamRequest(streamID, req)
			// then
			Expect(err).ToNot(HaveOccurred())

			By("waiting for Watchdog to get started")
			// when
			dataplaneID := <-watchdogCh
			// then
			Expect(dataplaneID).To(Equal(core_model.ResourceKey{Mesh: "demo", Name: "example"}))

			By("simulating another DiscoveryRequest")
			// when
			err = tracker.OnStreamRequest(streamID, req)
			// then
			Expect(err).ToNot(HaveOccurred())

			By("simulating Envoy disconnecting from the Control Plane")
			// and
			tracker.OnStreamClosed(streamID)

			By("waiting for Watchdog to get stopped")
			// when
			_, watchdogIsRunning := <-watchdogCh
			// then
			Expect(watchdogIsRunning).To(BeFalse())

			close(done)
		}, 5)

		It("should start only one watchdog per dataplane", func() {
			// setup
			var activeWatchdogs int32
			tracker := NewDataplaneSyncTracker(func(dataplaneId core_model.ResourceKey, streamId int64) util_watchdog.Watchdog {
				return WatchdogFunc(func(stop <-chan struct{}) {
					atomic.AddInt32(&activeWatchdogs, 1)
					<-stop
					atomic.AddInt32(&activeWatchdogs, -1)
				})
			})

			// when one stream for backend-01 is connected and request is sent
			streamID := int64(1)
			err := tracker.OnStreamOpen(context.Background(), streamID, "")
			Expect(err).ToNot(HaveOccurred())
			err = tracker.OnStreamRequest(streamID, &envoy.DiscoveryRequest{
				Node: &envoy_core.Node{
					Id: "default.backend-01",
				},
			})
			Expect(err).ToNot(HaveOccurred())

			// and when new stream from backend-01 is connected  and request is sent
			streamID = 2
			err = tracker.OnStreamOpen(context.Background(), streamID, "")
			Expect(err).ToNot(HaveOccurred())
			err = tracker.OnStreamRequest(streamID, &envoy.DiscoveryRequest{
				Node: &envoy_core.Node{
					Id: "default.backend-01",
				},
			})
			Expect(err).ToNot(HaveOccurred())

			// then only one watchdog is active
			Eventually(func() int32 {
				return atomic.LoadInt32(&activeWatchdogs)
			}, "5s", "10ms").Should(Equal(int32(1)))

			// when first stream is closed
			tracker.OnStreamClosed(1)

			// then watchdog is still active because other stream is opened
			Eventually(func() int32 {
				return atomic.LoadInt32(&activeWatchdogs)
			}, "5s", "10ms").Should(Equal(int32(1)))

			// when other stream is closed
			tracker.OnStreamClosed(2)

			// then no watchdog is stopped
			Eventually(func() int32 {
				return atomic.LoadInt32(&activeWatchdogs)
			}, "5s", "10ms").Should(Equal(int32(0)))
		})
	})
})

var _ util_watchdog.Watchdog = WatchdogFunc(nil)

type WatchdogFunc func(stop <-chan struct{})

func (f WatchdogFunc) Start(stop <-chan struct{}) {
	f(stop)
}
