package sync_test

import (
	"context"

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
					Id: "pilot.example.demo",
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
			Expect(dataplaneID).To(Equal(core_model.ResourceKey{Mesh: "pilot", Namespace: "demo", Name: "example"}))

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
	})
})

var _ util_watchdog.Watchdog = WatchdogFunc(nil)

type WatchdogFunc func(stop <-chan struct{})

func (f WatchdogFunc) Start(stop <-chan struct{}) {
	f(stop)
}
