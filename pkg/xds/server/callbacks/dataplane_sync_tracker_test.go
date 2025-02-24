package callbacks_test

import (
	"context"
	"sync/atomic"
	"time"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/test"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
	. "github.com/kumahq/kuma/pkg/xds/server/callbacks"
)

var _ = Describe("Sync", func() {
	Describe("dataplaneSyncTracker", func() {
		It("should not fail when ADS stream is closed before Watchdog is even created", func() {
			// setup
			tracker := DataplaneCallbacksToXdsCallbacks(NewDataplaneSyncTracker(nil))

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
			callbacks := util_xds_v3.AdaptCallbacks(DataplaneCallbacksToXdsCallbacks(tracker))

			// given
			ctx := context.Background()
			streamID := int64(1)
			typ := ""
			req := &envoy_sd.DiscoveryRequest{Node: nil}

			By("simulating Envoy connecting to the Control Plane")
			// when
			err := callbacks.OnStreamOpen(ctx, streamID, typ)
			// then
			Expect(err).ToNot(HaveOccurred())

			By("simulating DiscoveryRequest")
			// when
			err = callbacks.OnStreamRequest(streamID, req)
			// then
			Expect(err).ToNot(HaveOccurred())

			By("simulating Envoy disconnecting from the Control Plane")
			// and
			callbacks.OnStreamClosed(streamID, nil)

			// then
			// expect no panic
		})

		It("should create a Watchdog when Envoy presents a valid Node ID", test.Within(5*time.Second, func() {
			watchdogCh := make(chan core_model.ResourceKey)

			// setup
			tracker := NewDataplaneSyncTracker(func(key core_model.ResourceKey) util_xds_v3.Watchdog {
				return WatchdogFunc(func(ctx context.Context) {
					watchdogCh <- key
					<-ctx.Done()
					close(watchdogCh)
				})
			})
			callbacks := util_xds_v3.AdaptCallbacks(DataplaneCallbacksToXdsCallbacks(tracker))

			// given
			ctx := context.Background()
			streamID := int64(1)
			typ := ""
			n := &envoy_core.Node{Id: "demo.example"}
			req := &envoy_sd.DiscoveryRequest{Node: n}

			By("simulating Envoy connecting to the Control Plane")
			// when
			err := callbacks.OnStreamOpen(ctx, streamID, typ)
			// then
			Expect(err).ToNot(HaveOccurred())

			By("simulating DiscoveryRequest")
			// when
			err = callbacks.OnStreamRequest(streamID, req)
			// then
			Expect(err).ToNot(HaveOccurred())

			By("waiting for Watchdog to get started")
			// when
			dataplaneID := <-watchdogCh
			// then
			Expect(dataplaneID).To(Equal(core_model.ResourceKey{Mesh: "demo", Name: "example"}))

			By("simulating another DiscoveryRequest")
			// when
			err = callbacks.OnStreamRequest(streamID, req)
			// then
			Expect(err).ToNot(HaveOccurred())

			By("simulating Envoy disconnecting from the Control Plane")
			// and
			callbacks.OnStreamClosed(streamID, n)

			By("waiting for Watchdog to get stopped")
			// when
			_, watchdogIsRunning := <-watchdogCh
			// then
			Expect(watchdogIsRunning).To(BeFalse())
		}))

		It("should start only one watchdog per dataplane", func() {
			// setup
			var activeWatchdogs int32
			var cleanupDone atomic.Bool
			tracker := NewDataplaneSyncTracker(func(key core_model.ResourceKey) util_xds_v3.Watchdog {
				return WatchdogFunc(func(ctx context.Context) {
					atomic.AddInt32(&activeWatchdogs, 1)
					<-ctx.Done()
					atomic.AddInt32(&activeWatchdogs, -1)
					cleanupDone.Store(true)
				})
			})
			callbacks := util_xds_v3.AdaptCallbacks(DataplaneCallbacksToXdsCallbacks(tracker))

			// when one stream for backend-01 is connected and request is sent
			streamID1 := int64(1)
			streamID2 := int64(2)
			streamID3 := int64(3)
			err := callbacks.OnStreamOpen(context.Background(), streamID1, "")
			Expect(err).ToNot(HaveOccurred())
			node := &envoy_core.Node{Id: "default.backend-01"}
			err = callbacks.OnStreamRequest(streamID1, &envoy_sd.DiscoveryRequest{Node: node})
			Expect(err).ToNot(HaveOccurred())

			// then a watchdog is active
			Expect(atomic.LoadInt32(&activeWatchdogs)).To(Equal(int32(0)))

			// and when new stream from backend-01 is connected  and request is sent
			err = callbacks.OnStreamOpen(context.Background(), streamID2, "")
			Expect(err).ToNot(HaveOccurred())
			err = callbacks.OnStreamRequest(streamID2, &envoy_sd.DiscoveryRequest{Node: node})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("already an active stream"))

			// then only one watchdog is active
			Eventually(func() int32 {
				return atomic.LoadInt32(&activeWatchdogs)
			}, "5s", "10ms").Should(Equal(int32(1)))

			callbacks.OnStreamClosed(streamID2, node)
			Expect(cleanupDone.Load()).To(BeFalse())

			// when first stream is closed
			callbacks.OnStreamClosed(streamID1, node)
			Expect(cleanupDone.Load()).To(BeTrue())

			// then there is no active watchdog
			Expect(atomic.LoadInt32(&activeWatchdogs)).To(Equal(int32(0)))

			// and when the third stream from backend-01 is connected after the first active stream closed and request is sent
			err = callbacks.OnStreamOpen(context.Background(), streamID3, "")
			Expect(err).ToNot(HaveOccurred())
			err = callbacks.OnStreamRequest(streamID3, &envoy_sd.DiscoveryRequest{Node: node})
			Expect(err).ToNot(HaveOccurred())

			// then a watchdog is active
			Eventually(func() int32 {
				return atomic.LoadInt32(&activeWatchdogs)
			}, "5s", "10ms").Should(Equal(int32(1)))

			// when the third stream is closed
			callbacks.OnStreamClosed(streamID3, node)

			// then no watchdog is active
			Expect(atomic.LoadInt32(&activeWatchdogs)).To(Equal(int32(0)))
		})
	})
})

type WatchdogFunc func(ctx context.Context)

func (f WatchdogFunc) Start(ctx context.Context) {
	f(ctx)
}
