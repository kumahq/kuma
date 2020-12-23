package insights_test

import (
	"context"
	"sync"
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
	. "github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/util/proto"
)

type testEventReader struct {
	ch chan events.Event
}

func (t *testEventReader) Recv(stop <-chan struct{}) (events.Event, error) {
	return <-t.ch, nil
}

type testEventReaderFactory struct {
	reader *testEventReader
}

func (t *testEventReaderFactory) New() events.Listener {
	return t.reader
}

var _ = Describe("Insight Persistence", func() {
	var rm manager.ResourceManager
	nowMtx := &sync.RWMutex{}
	var now time.Time

	var eventCh chan events.Event
	var stopCh chan struct{}

	tickMtx := &sync.RWMutex{}
	var tickCh chan time.Time

	core.Now = func() time.Time {
		nowMtx.RLock()
		defer nowMtx.RUnlock()
		return now
	}

	BeforeEach(func() {
		rm = manager.NewResourceManager(memory.NewStore())
		now = time.Now()
		eventCh = make(chan events.Event)
		stopCh = make(chan struct{})

		tickMtx.Lock()
		tickCh = make(chan time.Time)
		tickMtx.Unlock()

		resyncer := insights.NewResyncer(&insights.Config{
			MinResyncTimeout:   5 * time.Second,
			MaxResyncTimeout:   1 * time.Minute,
			ResourceManager:    rm,
			EventReaderFactory: &testEventReaderFactory{reader: &testEventReader{ch: eventCh}},
			Tick: func(d time.Duration) (rv <-chan time.Time) {
				tickMtx.RLock()
				defer tickMtx.RUnlock()
				Expect(d).To(Equal(55 * time.Second)) // should be equal MaxResyncTimeout - MinResyncTimeout
				return tickCh
			},
		})
		go func() {
			err := resyncer.Start(stopCh)
			Expect(err).ToNot(HaveOccurred())
		}()
	})

	It("should sync more often than MaxResyncTimeout", func() {
		err := rm.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey("mesh-1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &core_mesh.TrafficPermissionResource{Spec: samples.TrafficPermission}, store.CreateByKey("tp-1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		nowMtx.Lock()
		now = now.Add(61 * time.Second)
		nowMtx.Unlock()
		tickCh <- now

		insight := core_mesh.NewMeshInsightResource()
		Eventually(func() error {
			return rm.Get(context.Background(), insight, store.GetByKey("mesh-1", model.NoMesh))
		}, "10s", "100ms").Should(BeNil())
		Expect(insight.Spec.Policies[string(core_mesh.TrafficPermissionType)].Total).To(Equal(uint32(1)))
		Expect(insight.Spec.LastSync).To(MatchProto(proto.MustTimestampProto(now)))
	})
})
