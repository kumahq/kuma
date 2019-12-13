package server_test

import (
	"context"
	"github.com/Kong/kuma/pkg/core/xds"
	xds_context "github.com/Kong/kuma/pkg/xds/context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/xds/server"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	kuma_cp "github.com/Kong/kuma/pkg/config/app/kuma-cp"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"

	test_runtime "github.com/Kong/kuma/pkg/test/runtime"
)

type event struct {
	Update *mesh_core.DataplaneResource
	Delete core_model.ResourceKey
}

type eventSnapshotReconciler struct {
	events chan event
}

func (e *eventSnapshotReconciler) Reconcile(ctx xds_context.Context, proxy *xds.Proxy) error {
	e.events <- event{Update: proxy.Dataplane}
	return nil
}

func (e *eventSnapshotReconciler) Clear(proxyId *xds.ProxyId) error {
	e.events <- event{Delete: proxyId.ToResourceKey()}
	return nil
}

var _ SnapshotReconciler = &eventSnapshotReconciler{}

var _ = Describe("Components", func() {
	Describe("DefaultDataplaneSyncTracker", func() {
		It("", func(done Done) {
			// given
			cfg := kuma_cp.DefaultConfig()
			cfg.XdsServer.DataplaneConfigurationRefreshInterval = 1 * time.Millisecond

			// and
			runtime, err := test_runtime.BuilderFor(cfg).Build()
			Expect(err).ToNot(HaveOccurred())

			// and example mesh
			opts := core_store.CreateByKey("pilot", "pilot")
			err = runtime.ResourceManager().Create(context.Background(), &mesh_core.MeshResource{}, opts)
			Expect(err).ToNot(HaveOccurred())

			// setup
			reconciler := eventSnapshotReconciler{}
			reconciler.events = make(chan event)
			// and
			tracker, err := DefaultDataplaneSyncTracker(runtime, &reconciler, NewDataplaneMetadataTracker())
			Expect(err).ToNot(HaveOccurred())

			// given
			ctx := context.Background()
			streamID := int64(1)
			typ := ""
			req := &envoy.DiscoveryRequest{
				Node: &envoy_core.Node{
					Id: "pilot.example",
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
			nextEvent := <-reconciler.events
			// then
			Expect(nextEvent.Delete).To(Equal(core_model.ResourceKey{Mesh: "pilot", Name: "example"}))

			By("creating Dataplane definition")
			// when
			resource := &mesh_core.DataplaneResource{
				Spec: mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Interface: "127.0.0.1:9090:8080",
								Tags: map[string]string{
									"service": "backend",
								},
							},
						},
					},
				},
			}
			err = runtime.ResourceManager().Create(ctx, resource, core_store.CreateBy(core_model.ResourceKey{Mesh: "pilot", Name: "example"}))
			// then
			Expect(err).ToNot(HaveOccurred())

			By("waiting for Watchdog to trigger Dataplane configuration refresh (update)")
			// expect
			Eventually(func() bool {
				nextEvent := <-reconciler.events
				return nextEvent.Update != nil
			}, "1s", "1ms").Should(BeTrue())

			By("simulating Envoy disconnecting from the Control Plane")
			// and
			tracker.OnStreamClosed(streamID)

			close(done)
		}, 10)
	})
})
