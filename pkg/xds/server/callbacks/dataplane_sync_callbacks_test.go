package callbacks_test

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/test"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
	. "github.com/kumahq/kuma/pkg/xds/server/callbacks"
	xds_sync "github.com/kumahq/kuma/pkg/xds/sync"
)

type dummyLifecycleManager struct{}

func (c *dummyLifecycleManager) Register(logr.Logger, context.Context, *core_xds.DataplaneMetadata) error {
	return nil
}

func (c *dummyLifecycleManager) Deregister(logr.Logger, context.Context) {
}

var _ = Describe("Sync", func() {
	Describe("dataplaneSyncCallbacks", func() {
		DescribeTable("when concurrent calls", MustPassRepeatedly(3),
			func(run func(callbacks envoy_xds.Callbacks, req *envoy_sd.DiscoveryRequest, r *rand.Rand)) {
				wdCount := atomic.Int32{}
				events := make(chan string, 1000)
				r := rand.New(rand.NewSource(GinkgoRandomSeed())) // #nosec G404 - math rand is enough
				stateFactory := xds_sync.DataplaneWatchdogFactoryFunc(func(key core_model.ResourceKey, fetchMeta func() *core_xds.DataplaneMetadata) util_xds_v3.Watchdog {
					wdNum := wdCount.Add(1)
					events <- fmt.Sprintf("create:%d", wdNum)
					return util_xds_v3.WatchdogFunc(func(ctx context.Context) {
						events <- fmt.Sprintf("start:%d", wdNum)
						time.Sleep(time.Duration(r.Intn(200)) * time.Millisecond)
						<-ctx.Done()
						time.Sleep(time.Duration(r.Intn(100)) * time.Millisecond)
						events <- fmt.Sprintf("stop:%d", wdNum)
					})
				})
				tracker := NewDataplaneSyncCallbacks(stateFactory, DataplaneLifecycleManagerFunc(func(key core_model.ResourceKey) DataplaneLifecycleManager {
					return &dummyLifecycleManager{}
				}))

				n := &envoy_core.Node{Id: "demo.example"}
				req := &envoy_sd.DiscoveryRequest{Node: n}
				callbacks := util_xds_v3.AdaptCallbacks(tracker)
				run(callbacks, req, r)

				time.Sleep(time.Second)
				// Let's check events:
				close(events)
				var all []string
				for v := range events {
					all = append(all, v)
				}
				GinkgoWriter.Printf("Events: %v", strings.Join(all, ","))
				for i := range all {
					if i == 0 {
						Expect(all[0]).To(HavePrefix("create:"))
						continue
					}
					switch {
					case strings.HasPrefix(all[i], "create:"):
						Expect(all[i-1]).To(HavePrefix("stop:"))
					case strings.HasPrefix(all[i], "start:"):
						Expect(all[i-1]).To(HavePrefix("create:"))
					case strings.HasPrefix(all[i], "stop:"):
						Expect(all[i-1]).To(HavePrefix("start:"))
					}
				}
				Expect(all[len(all)-1]).To(HavePrefix("stop:"))
			},
			Entry("simple same thread", func(callbacks envoy_xds.Callbacks, req *envoy_sd.DiscoveryRequest, _ *rand.Rand) {
				ctx := context.Background()
				streamA := int64(1)
				streamB := int64(2)
				Expect(callbacks.OnStreamOpen(ctx, streamA, "")).To(Succeed())
				Expect(callbacks.OnStreamOpen(ctx, streamB, "")).To(Succeed())

				Expect(callbacks.OnStreamRequest(streamA, req)).To(Succeed())
				Expect(callbacks.OnStreamRequest(streamB, req)).To(Succeed())

				callbacks.OnStreamClosed(streamB, nil)
				callbacks.OnStreamClosed(streamA, nil)
			}),
			Entry("concurrent clients", func(callbacks envoy_xds.Callbacks, req *envoy_sd.DiscoveryRequest, r *rand.Rand) {
				wg := sync.WaitGroup{}
				wg.Add(2)
				for i := range 2 {
					go func() {
						defer GinkgoRecover()
						defer wg.Done()
						ctx := context.Background()
						streamID := int64(i)
						time.Sleep(time.Duration(r.Intn(200)) * time.Millisecond)
						Expect(callbacks.OnStreamOpen(ctx, streamID, "")).To(Succeed())
						time.Sleep(time.Duration(r.Intn(200)) * time.Millisecond)
						Expect(callbacks.OnStreamRequest(streamID, req)).To(Succeed())
						time.Sleep(time.Duration(r.Intn(200)) * time.Millisecond)
						callbacks.OnStreamClosed(streamID, nil)
						time.Sleep(time.Duration(r.Intn(200)) * time.Millisecond)
					}()
				}
				wg.Wait()
			}),
			Entry("start/stop and then restart", func(callbacks envoy_xds.Callbacks, req *envoy_sd.DiscoveryRequest, r *rand.Rand) {
				ctx := context.Background()
				streamA := int64(1)
				Expect(callbacks.OnStreamOpen(ctx, streamA, "")).To(Succeed())
				Expect(callbacks.OnStreamRequest(streamA, req)).To(Succeed())
				time.Sleep(time.Duration(r.Intn(200)) * time.Millisecond)
				streamB := int64(2)
				Expect(callbacks.OnStreamOpen(ctx, streamB, "")).To(Succeed())
				go func() { // We do this in another goroutine so that the new stream happens concurrently with the old one closing.
					defer GinkgoRecover()
					callbacks.OnStreamClosed(streamA, nil)
				}()
				Expect(callbacks.OnStreamRequest(streamB, req)).To(Succeed())
				time.Sleep(time.Duration(r.Intn(200)) * time.Millisecond)
				callbacks.OnStreamClosed(streamB, nil)
			}),
		)
		It("should not fail when ADS stream is closed before Watchdog is even created", func() {
			// setup
			tracker := NewDataplaneSyncCallbacks(nil, nil)

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
			callbacks := util_xds_v3.AdaptCallbacks(NewDataplaneSyncCallbacks(nil, nil))

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
			tracker := NewDataplaneSyncCallbacks(
				xds_sync.DataplaneWatchdogFactoryFunc(func(key core_model.ResourceKey, fetchMeta func() *core_xds.DataplaneMetadata) util_xds_v3.Watchdog {
					return util_xds_v3.WatchdogFunc(func(ctx context.Context) {
						watchdogCh <- key
						<-ctx.Done()
						close(watchdogCh)
					})
				}),
				DataplaneLifecycleManagerFunc(func(key core_model.ResourceKey) DataplaneLifecycleManager {
					return &dummyLifecycleManager{}
				}),
			)
			callbacks := util_xds_v3.AdaptCallbacks(tracker)

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
			tracker := NewDataplaneSyncCallbacks(
				xds_sync.DataplaneWatchdogFactoryFunc(func(key core_model.ResourceKey, _ func() *core_xds.DataplaneMetadata) util_xds_v3.Watchdog {
					return util_xds_v3.WatchdogFunc(func(ctx context.Context) {
						atomic.AddInt32(&activeWatchdogs, 1)
						<-ctx.Done()
						atomic.AddInt32(&activeWatchdogs, -1)
					})
				}),
				DataplaneLifecycleManagerFunc(func(key core_model.ResourceKey) DataplaneLifecycleManager {
					return &dummyLifecycleManager{}
				}),
			)
			callbacks := util_xds_v3.AdaptCallbacks(tracker)

			// when one stream for backend-01 is connected and request is sent
			streamID := int64(1)
			err := callbacks.OnStreamOpen(context.Background(), streamID, "")
			Expect(err).ToNot(HaveOccurred())
			n := &envoy_core.Node{Id: "default.backend-01"}
			err = callbacks.OnStreamRequest(streamID, &envoy_sd.DiscoveryRequest{Node: n})
			Expect(err).ToNot(HaveOccurred())

			// and when new stream from backend-01 is connected  and request is sent
			streamID = 2
			err = callbacks.OnStreamOpen(context.Background(), streamID, "")
			Expect(err).ToNot(HaveOccurred())
			err = callbacks.OnStreamRequest(streamID, &envoy_sd.DiscoveryRequest{
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
			callbacks.OnStreamClosed(1, n)

			// then watchdog is still active because other stream is opened
			Eventually(func() int32 {
				return atomic.LoadInt32(&activeWatchdogs)
			}, "5s", "10ms").Should(Equal(int32(1)))

			// when other stream is closed
			callbacks.OnStreamClosed(2, n)

			// then no watchdog is stopped
			Eventually(func() int32 {
				return atomic.LoadInt32(&activeWatchdogs)
			}, "5s", "10ms").Should(Equal(int32(0)))
		})
	})
})
