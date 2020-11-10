package insights_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/events"
	"github.com/kumahq/kuma/pkg/insights"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/kds/samples"
	"github.com/kumahq/kuma/pkg/util/proto"
)

type testEventReader struct {
	ch chan events.Event
}

func (t *testEventReader) Recv(stop <-chan struct{}) (events.Event, error) {
	return <-t.ch, nil
}

var _ = Describe("Insight Persistence", func() {
	var rm manager.ResourceManager
	var now time.Time

	var eventCh chan events.Event
	var stopCh chan struct{}
	var tickCh chan time.Time

	core.Now = func() time.Time {
		return now
	}

	BeforeEach(func() {
		rm = manager.NewResourceManager(memory.NewStore())
		now = time.Now()
		eventCh = make(chan events.Event)
		stopCh = make(chan struct{})
		tickCh = make(chan time.Time)

		resyncer := insights.NewResyncer(&insights.Config{
			MinResyncTimeout: 5 * time.Second,
			MaxResyncTimeout: 1 * time.Minute,
			ResourceManager:  rm,
			EventReader:      &testEventReader{ch: eventCh},
			Tick: func(d time.Duration) <-chan time.Time {
				Expect(d).To(Equal(55 * time.Second)) // should be equal MaxResyncTimeout - MinResyncTimeout
				return tickCh
			},
		})
		go func() {
			err := resyncer.Start(stopCh)
			Expect(err).ToNot(HaveOccurred())
		}()
	})

	It("should not sync more often than MinResyncTimeout", func() {
		err := rm.Create(context.Background(), &core_mesh.MeshResource{}, store.CreateByKey("mesh-1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &core_mesh.TrafficPermissionResource{Spec: samples.TrafficPermission}, store.CreateByKey("tp-1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		eventCh <- events.Event{
			Operation: events.Create,
			Type:      core_mesh.TrafficPermissionType,
			Key:       model.ResourceKey{Name: "tp-1", Mesh: "mesh-1"},
		}

		insight := &core_mesh.MeshInsightResource{}
		Eventually(func() error {
			return rm.Get(context.Background(), insight, store.GetByKey("mesh-1", "mesh-1"))
		}, "10s", "100ms").Should(BeNil())
		Expect(insight.Spec.Policies[string(core_mesh.TrafficPermissionType)].Total).To(Equal(uint32(1)))
		Expect(insight.Spec.LastSync).To(Equal(proto.MustTimestampProto(now)))

		prev := now
		now = now.Add(1 * time.Second)

		err = rm.Create(context.Background(), &core_mesh.TrafficPermissionResource{Spec: samples.TrafficPermission}, store.CreateByKey("tp-2", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		eventCh <- events.Event{
			Operation: events.Create,
			Type:      core_mesh.TrafficPermissionType,
			Key:       model.ResourceKey{Name: "tp-2", Mesh: "mesh-1"},
		}

		err = rm.Get(context.Background(), insight, store.GetByKey("mesh-1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())
		Expect(insight.Spec.Policies[string(core_mesh.TrafficPermissionType)].Total).To(Equal(uint32(1)))
		Expect(insight.Spec.LastSync).To(Equal(proto.MustTimestampProto(prev)))
	})

	It("should sync more often than MaxResyncTimeout", func() {
		err := rm.Create(context.Background(), &core_mesh.MeshResource{}, store.CreateByKey("mesh-1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &core_mesh.TrafficPermissionResource{Spec: samples.TrafficPermission}, store.CreateByKey("tp-1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		now = now.Add(61 * time.Second)
		tickCh <- now

		insight := &core_mesh.MeshInsightResource{}
		Eventually(func() error {
			return rm.Get(context.Background(), insight, store.GetByKey("mesh-1", "mesh-1"))
		}, "10s", "100ms").Should(BeNil())
		Expect(insight.Spec.Policies[string(core_mesh.TrafficPermissionType)].Total).To(Equal(uint32(1)))
		Expect(insight.Spec.LastSync).To(Equal(proto.MustTimestampProto(now)))
	})
})
