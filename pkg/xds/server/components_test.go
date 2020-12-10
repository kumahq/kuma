package server_test

import (
	"context"
	"time"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/xds"
	test_metrics "github.com/kumahq/kuma/pkg/test/metrics"
	test_runtime "github.com/kumahq/kuma/pkg/test/runtime"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	. "github.com/kumahq/kuma/pkg/xds/server"
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
			opts := core_store.CreateByKey("demo", core_model.NoMesh)
			err = runtime.ResourceManager().Create(context.Background(), mesh_core.NewMeshResource(), opts)
			Expect(err).ToNot(HaveOccurred())

			// setup
			reconciler := eventSnapshotReconciler{}
			reconciler.events = make(chan event)
			// and
			tracker, err := DefaultDataplaneSyncTracker(runtime, &reconciler, nil, NewDataplaneMetadataTracker(), NewConnectionInfoTracker())
			Expect(err).ToNot(HaveOccurred())

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
			Expect(nextEvent.Delete).To(Equal(core_model.ResourceKey{Mesh: "demo", Name: "example"}))

			By("creating Dataplane definition")
			// when
			resource := &mesh_core.DataplaneResource{
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "127.0.0.1",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Port:        9090,
								ServicePort: 8080,
								Tags: map[string]string{
									"kuma.io/service": "backend",
								},
							},
						},
					},
				},
			}
			err = runtime.ResourceManager().Create(ctx, resource, core_store.CreateBy(core_model.ResourceKey{Mesh: "demo", Name: "example"}))
			// then
			Expect(err).ToNot(HaveOccurred())

			By("waiting for Watchdog to trigger Dataplane configuration refresh (update)")
			// expect
			Eventually(func() bool {
				nextEvent := <-reconciler.events
				return nextEvent.Update != nil
			}, "1s", "1ms").Should(BeTrue())

			// and metrics are published
			Expect(test_metrics.FindMetric(runtime.Metrics(), "xds_generation")).ToNot(BeNil())

			By("simulating Envoy disconnecting from the Control Plane")
			// and
			tracker.OnStreamClosed(streamID)

			close(done)
		}, 10)
	})
})
