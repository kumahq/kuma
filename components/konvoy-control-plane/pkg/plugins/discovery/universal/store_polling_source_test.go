package universal

import (
	"context"
	"sync"
	"time"

	"github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/discovery"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/memory"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ discovery.DiscoveryConsumer = &testDiscoveryConsumer{}

type testDiscoveryConsumer struct {
	mut      sync.RWMutex
	updates  []*mesh.DataplaneResource
	removals []model.ResourceKey
}

func (l *testDiscoveryConsumer) OnServiceUpdate(*discovery.ServiceInfo) error {
	return nil
}

func (l *testDiscoveryConsumer) OnServiceDelete(core.NamespacedName) error {
	return nil
}

func (l *testDiscoveryConsumer) OnWorkloadUpdate(*discovery.WorkloadInfo) error {
	return nil
}

func (l *testDiscoveryConsumer) OnWorkloadDelete(core.NamespacedName) error {
	return nil
}

func (l *testDiscoveryConsumer) OnDataplaneUpdate(resource *mesh.DataplaneResource) error {
	l.mut.Lock()
	defer l.mut.Unlock()
	l.updates = append(l.updates, resource)
	return nil
}

func (l *testDiscoveryConsumer) OnDataplaneDelete(key model.ResourceKey) error {
	l.mut.Lock()
	defer l.mut.Unlock()
	l.removals = append(l.removals, key)
	return nil
}

func (l *testDiscoveryConsumer) Updates() []*mesh.DataplaneResource {
	l.mut.RLock()
	defer l.mut.RUnlock()
	return l.updates
}

func (l *testDiscoveryConsumer) Removals() []model.ResourceKey {
	l.mut.RLock()
	defer l.mut.RUnlock()
	return l.removals
}

var _ = Describe("Store Polling Source", func() {

	var memoryStore store.ResourceStore
	var consumer *testDiscoveryConsumer
	var source *storePollingSource

	BeforeEach(func() {
		memoryStore = memory.NewStore()
		source = newStorePollingSource(
			memoryStore,
			10*time.Millisecond,
		)
		consumer = &testDiscoveryConsumer{}
		source.AddConsumer(consumer)
	})

	var resource *mesh.DataplaneResource
	BeforeEach(func() {
		// setup sample dataplane in store
		resource = &mesh.DataplaneResource{
			Spec: v1alpha1.Dataplane{
				Networking: &v1alpha1.Dataplane_Networking{
					Inbound: []*v1alpha1.Dataplane_Networking_Inbound{
						{
							Interface: "192.168.0.1:80:8080",
							Tags: map[string]string{
								"some": "tag",
							},
						},
					},
				},
			},
		}
		err := memoryStore.Create(context.Background(), resource, store.CreateByKey("sample-ns", "sample-mesh", "sample-name"))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should notify about added dataplanes", func() {
		// when
		err := source.detectChanges()
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(consumer.Removals()).To(HaveLen(0))
		Expect(consumer.Updates()).To(HaveLen(1))
		Expect(consumer.Updates()[0]).To(Equal(resource))
	})

	It("should notify about deleted dataplanes", func() {
		// given detected created dataplane
		err := source.detectChanges()
		Expect(err).ToNot(HaveOccurred())
		Expect(consumer.Updates()).To(HaveLen(1))

		// when
		err = memoryStore.Delete(context.Background(), resource, store.DeleteByKey("sample-ns", "sample-mesh", "sample-name"))
		Expect(err).ToNot(HaveOccurred())

		err = source.detectChanges()
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(consumer.Updates()).To(HaveLen(1)) // should stay the same
		Expect(consumer.Removals()).To(HaveLen(1))
		Expect(consumer.Removals()[0]).To(Equal(model.MetaToResourceKey(resource.Meta)))
	})

	It("should notify about modified dataplanes", func() {
		// given detected created dataplane
		err := source.detectChanges()
		Expect(err).ToNot(HaveOccurred())
		Expect(consumer.Updates()).To(HaveLen(1))

		// when detect modified version
		resource.Spec.Networking.Inbound[0].Tags["some"] = "updated"
		err = memoryStore.Update(context.Background(), resource)
		Expect(err).ToNot(HaveOccurred())

		// when
		err = source.detectChanges()
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(consumer.Updates()).To(HaveLen(2))
		Expect(consumer.Removals()).To(HaveLen(0))
		Expect(consumer.Updates()[1].Spec.Networking.Inbound[0].Tags["some"]).To(Equal("updated"))
	})

	It("should periodically detect changes", func() {
		// when
		go func() {
			_ = source.Start(make(chan struct{}))
		}()

		// then
		Eventually(func() *mesh.DataplaneResource {
			if len(consumer.Updates()) == 0 {
				return nil
			}
			return consumer.Updates()[0]
		}, "1s", "10ms").Should(Equal(resource))
	})
})
