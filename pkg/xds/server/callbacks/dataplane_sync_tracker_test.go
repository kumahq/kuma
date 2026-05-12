package callbacks_test

import (
	"context"
	"sync/atomic"
	"time"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
	"github.com/kumahq/kuma/v2/pkg/test"
	util_xds_v3 "github.com/kumahq/kuma/v2/pkg/util/xds/v3"
	. "github.com/kumahq/kuma/v2/pkg/xds/server/callbacks"
	"github.com/kumahq/kuma/v2/pkg/xds/sync"
)

var _ = Describe("Sync", func() {
	Describe("dataplaneSyncTracker", func() {
		It("should not fail when ADS stream is closed before Watchdog is even created", func() {
			// setup
			tracker := DataplaneCallbacksToXdsCallbacks(NewDataplaneSyncTracker(nil), nil)

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
			callbacks := util_xds_v3.AdaptCallbacks(DataplaneCallbacksToXdsCallbacks(tracker, nil))

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
			tracker := NewDataplaneSyncTracker(sync.DataplaneWatchdogFactoryFunc(func(key core_model.ResourceKey, _ *core_xds.DataplaneMetadata) util_xds_v3.Watchdog {
				return util_xds_v3.WatchdogFunc(func(ctx context.Context) {
					watchdogCh <- key
					<-ctx.Done()
					close(watchdogCh)
				})
			}))
			callbacks := util_xds_v3.AdaptCallbacks(DataplaneCallbacksToXdsCallbacks(tracker, nil))

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
			var activeWatchdogs atomic.Int32
			var cleanupDone atomic.Bool
			tracker := NewDataplaneSyncTracker(sync.DataplaneWatchdogFactoryFunc(func(key core_model.ResourceKey, _ *core_xds.DataplaneMetadata) util_xds_v3.Watchdog {
				return util_xds_v3.WatchdogFunc(func(ctx context.Context) {
					activeWatchdogs.Add(1)
					<-ctx.Done()
					activeWatchdogs.Add(-1)
					cleanupDone.Store(true)
				})
			}))
			callbacks := util_xds_v3.AdaptCallbacks(DataplaneCallbacksToXdsCallbacks(tracker, nil))

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
			Eventually(activeWatchdogs.Load, "5s", "10ms").Should(Equal(int32(1)))

			// and when new stream from backend-01 is connected  and request is sent
			err = callbacks.OnStreamOpen(context.Background(), streamID2, "")
			Expect(err).ToNot(HaveOccurred())
			err = callbacks.OnStreamRequest(streamID2, &envoy_sd.DiscoveryRequest{Node: node})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("already an active stream"))

			// then only one watchdog is active
			Eventually(activeWatchdogs.Load, "5s", "10ms").Should(Equal(int32(1)))

			callbacks.OnStreamClosed(streamID2, node)
			Expect(cleanupDone.Load()).To(BeFalse())

			// when first stream is closed
			callbacks.OnStreamClosed(streamID1, node)
			Expect(cleanupDone.Load()).To(BeTrue())

			// then there is no active watchdog
			Expect(activeWatchdogs.Load()).To(Equal(int32(0)))

			// and when the third stream from backend-01 is connected after the first active stream closed and request is sent
			err = callbacks.OnStreamOpen(context.Background(), streamID3, "")
			Expect(err).ToNot(HaveOccurred())
			err = callbacks.OnStreamRequest(streamID3, &envoy_sd.DiscoveryRequest{Node: node})
			Expect(err).ToNot(HaveOccurred())

			// then a watchdog is active
			Eventually(activeWatchdogs.Load, "5s", "10ms").Should(Equal(int32(1)))

			// when the third stream is closed
			callbacks.OnStreamClosed(streamID3, node)

			// then no watchdog is active
			Expect(activeWatchdogs.Load()).To(Equal(int32(0)))
		})

		It("should pass stream context to watchdog factory when supported", test.Within(5*time.Second, func() {
			streamCtxCh := make(chan context.Context, 1)
			watchdogDone := make(chan struct{})

			// setup - factory that captures stream context
			factory := &streamCtxCapturingFactory{
				streamCtxCh:  streamCtxCh,
				watchdogDone: watchdogDone,
			}
			tracker := NewDataplaneSyncTracker(factory)
			callbacks := util_xds_v3.AdaptCallbacks(DataplaneCallbacksToXdsCallbacks(tracker, nil))

			// given
			ctx := context.Background()
			streamID := int64(1)
			node := &envoy_core.Node{Id: "demo.example"}
			req := &envoy_sd.DiscoveryRequest{Node: node}

			By("simulating Envoy connecting to the Control Plane")
			err := callbacks.OnStreamOpen(ctx, streamID, "")
			Expect(err).ToNot(HaveOccurred())

			By("simulating DiscoveryRequest")
			err = callbacks.OnStreamRequest(streamID, req)
			Expect(err).ToNot(HaveOccurred())

			By("verifying stream context was passed to factory")
			var capturedStreamCtx context.Context
			Eventually(streamCtxCh).Should(Receive(&capturedStreamCtx))
			Expect(capturedStreamCtx).ToNot(BeNil())

			By("simulating Envoy disconnecting from the Control Plane")
			callbacks.OnStreamClosed(streamID, node)

			By("waiting for Watchdog to stop")
			Eventually(watchdogDone).Should(BeClosed())
		}))

		It("should stop stale watchdog and keep new owner watchdog running", test.Within(5*time.Second, func() {
			var started atomic.Int32
			var stopped atomic.Int32
			var active atomic.Int32

			tracker := NewDataplaneSyncTracker(sync.DataplaneWatchdogFactoryFunc(func(key core_model.ResourceKey, _ *core_xds.DataplaneMetadata) util_xds_v3.Watchdog {
				return util_xds_v3.WatchdogFunc(func(ctx context.Context) {
					started.Add(1)
					active.Add(1)
					<-ctx.Done()
					active.Add(-1)
					stopped.Add(1)
				})
			}))
			callbacks := util_xds_v3.AdaptCallbacks(DataplaneCallbacksToXdsCallbacks(tracker, nil))

			node := &envoy_core.Node{Id: "default.backend-01"}
			req := &envoy_sd.DiscoveryRequest{Node: node}

			ctx1, cancel1 := context.WithCancel(context.Background())
			Expect(callbacks.OnStreamOpen(ctx1, 1, "")).To(Succeed())
			Expect(callbacks.OnStreamRequest(1, req)).To(Succeed())
			Eventually(started.Load).Should(Equal(int32(1)))
			Eventually(active.Load).Should(Equal(int32(1)))

			// Mark stream 1 as stale before cleanup callback runs.
			cancel1()

			Expect(callbacks.OnStreamOpen(context.Background(), 2, "")).To(Succeed())
			Expect(callbacks.OnStreamRequest(2, req)).To(Succeed())
			Eventually(started.Load).Should(Equal(int32(2)))

			// Closing stale stream should only stop stale watchdog and keep new one running.
			callbacks.OnStreamClosed(1, node)
			Eventually(stopped.Load).Should(Equal(int32(1)))
			Consistently(active.Load).Should(Equal(int32(1)))

			callbacks.OnStreamClosed(2, node)
			Eventually(stopped.Load).Should(Equal(int32(2)))
			Eventually(active.Load).Should(Equal(int32(0)))
		}))
	})
})

// streamCtxCapturingFactory implements DataplaneWatchdogFactoryWithStreamCtx
type streamCtxCapturingFactory struct {
	streamCtxCh  chan context.Context
	watchdogDone chan struct{}
}

func (f *streamCtxCapturingFactory) New(key core_model.ResourceKey, meta *core_xds.DataplaneMetadata) util_xds_v3.Watchdog {
	return f.NewWithStreamCtx(key, meta, nil)
}

func (f *streamCtxCapturingFactory) NewWithStreamCtx(key core_model.ResourceKey, meta *core_xds.DataplaneMetadata, streamCtx context.Context) util_xds_v3.Watchdog {
	return util_xds_v3.WatchdogFunc(func(ctx context.Context) {
		if streamCtx != nil {
			f.streamCtxCh <- streamCtx
		}
		<-ctx.Done()
		close(f.watchdogDone)
	})
}
