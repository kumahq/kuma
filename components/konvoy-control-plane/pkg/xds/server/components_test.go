package server_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/server"

	konvoy_cp "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoy-cp"
	core_discovery "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/discovery"
	mesh_core "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	core_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"

	test_runtime "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/test/runtime"
)

var _ = Describe("Components", func() {
	Describe("DefaultDataplaneSyncTracker", func() {
		It("", func(done Done) {
			// given
			cfg := konvoy_cp.DefaultConfig()
			cfg.XdsServer.DataplaneConfigurationRefreshInterval = 1 * time.Millisecond

			// and
			runtime, err := test_runtime.BuilderFor(cfg).Build()
			Expect(err).ToNot(HaveOccurred())

			// and example mesh
			opts := core_store.CreateByKey("demo", "pilot", "pilot")
			err = runtime.ResourceManager().Create(context.Background(), &mesh_core.MeshResource{}, opts)
			Expect(err).ToNot(HaveOccurred())

			// setup
			type event struct {
				Update *mesh_core.DataplaneResource
				Delete core_model.ResourceKey
			}
			events := make(chan event)
			sink := &core_discovery.DiscoverySink{
				DataplaneConsumer: DataplaneDiscoveryConsumerFuncs{
					OnDataplaneUpdateFunc: func(dataplane *mesh_core.DataplaneResource) error {
						events <- event{Update: dataplane}
						return nil
					},
					OnDataplaneDeleteFunc: func(key core_model.ResourceKey) error {
						events <- event{Delete: key}
						return nil
					},
				},
			}
			// and
			tracker := DefaultDataplaneSyncTracker(runtime, sink)

			// given
			ctx := context.Background()
			streamID := int64(1)
			typ := ""
			req := &envoy.DiscoveryRequest{
				Node: &envoy_core.Node{
					Id: "example.demo.pilot",
				},
			}

			By("simulating Envoy connecting to the Control Plane")
			// when
			err = tracker.OnStreamOpen(ctx, streamID, typ)
			// then
			Expect(err).ToNot(HaveOccurred())

			By("simulating DiscoveryRequest")
			// when
			err = tracker.OnStreamRequest(streamID, req)
			// then
			Expect(err).ToNot(HaveOccurred())

			By("waiting for Watchdog to trigger Dataplane configuration refresh (delete)")
			// when
			nextEvent := <-events
			// then
			Expect(nextEvent.Delete).To(Equal(core_model.ResourceKey{Mesh: "pilot", Namespace: "demo", Name: "example"}))

			By("creating Dataplane defintion")
			// when
			err = runtime.ResourceManager().Create(ctx, &mesh_core.DataplaneResource{}, core_store.CreateBy(core_model.ResourceKey{Mesh: "pilot", Namespace: "demo", Name: "example"}))
			// then
			Expect(err).ToNot(HaveOccurred())

			By("waiting for Watchdog to trigger Dataplane configuration refresh (update)")
			// expect
			Eventually(func() bool {
				nextEvent := <-events
				return nextEvent.Update != nil
			}, "1s", "1ms").Should(BeTrue())

			By("simulating Envoy disconnecting from the Control Plane")
			// and
			tracker.OnStreamClosed(streamID)

			close(done)
		}, 10)
	})
})

var _ core_discovery.DataplaneDiscoveryConsumer = DataplaneDiscoveryConsumerFuncs{}

type DataplaneDiscoveryConsumerFuncs struct {
	OnDataplaneUpdateFunc func(*mesh_core.DataplaneResource) error
	OnDataplaneDeleteFunc func(core_model.ResourceKey) error
}

func (f DataplaneDiscoveryConsumerFuncs) OnDataplaneUpdate(dataplane *mesh_core.DataplaneResource) error {
	if f.OnDataplaneUpdateFunc != nil {
		return f.OnDataplaneUpdateFunc(dataplane)
	}
	return nil
}
func (f DataplaneDiscoveryConsumerFuncs) OnDataplaneDelete(key core_model.ResourceKey) error {
	if f.OnDataplaneDeleteFunc != nil {
		return f.OnDataplaneDeleteFunc(key)
	}
	return nil
}
