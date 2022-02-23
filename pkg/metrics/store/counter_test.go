package metrics_test

import (
	"context"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	io_prometheus_client "github.com/prometheus/client_model/go"

	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/events"
	"github.com/kumahq/kuma/pkg/insights"
	test_insights "github.com/kumahq/kuma/pkg/insights/test"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	metrics_store "github.com/kumahq/kuma/pkg/metrics/store"
	store_memory "github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/kds/samples"
	test_metrics "github.com/kumahq/kuma/pkg/test/metrics"
)

var _ = Describe("Counter", func() {
	var store core_store.ResourceStore
	var resManager manager.ResourceManager
	var metrics core_metrics.Metrics

	var eventCh chan events.Event
	var stop chan struct{}

	nowMtx := &sync.RWMutex{}
	var now time.Time

	tickMtx := &sync.RWMutex{}
	var tickCh chan time.Time

	core.Now = func() time.Time {
		nowMtx.RLock()
		defer nowMtx.RUnlock()
		return now
	}

	BeforeEach(func() {
		var err error

		now = time.Now()
		eventCh = make(chan events.Event)
		stop = make(chan struct{})
		tickCh = make(chan time.Time)

		metrics, err = core_metrics.NewMetrics("Standalone")
		Expect(err).ToNot(HaveOccurred())

		memoryStore := store_memory.NewStore()
		store, err = metrics_store.NewMeteredStore(memoryStore, metrics)
		Expect(err).ToNot(HaveOccurred())

		resManager = manager.NewResourceManager(store)

		counterTicker := time.NewTicker(500 * time.Millisecond)
		counter, err := metrics_store.NewStoreCounter(resManager, metrics)
		Expect(err).ToNot(HaveOccurred())

		resyncer := insights.NewResyncer(&insights.Config{
			MinResyncTimeout:   5 * time.Second,
			MaxResyncTimeout:   1 * time.Minute,
			ResourceManager:    resManager,
			EventReaderFactory: &test_insights.TestEventReaderFactory{Reader: &test_insights.TestEventReader{Ch: eventCh}},
			Tick: func(d time.Duration) (rv <-chan time.Time) {
				tickMtx.RLock()
				defer tickMtx.RUnlock()
				return tickCh
			},
			Registry: registry.Global(),
		})

		go func() {
			err := resyncer.Start(stop)
			Expect(err).ToNot(HaveOccurred())
		}()

		go func() {
			err = counter.StartWithTicker(stop, counterTicker)
			Expect(err).ToNot(HaveOccurred())
		}()
	})

	AfterEach(func() {
		stop <- struct{}{}
	})

	findGauge := func(resTypeName string) *io_prometheus_client.Gauge {
		return test_metrics.FindMetric(metrics, "resources_count", "resource_type", resTypeName).GetGauge()
	}

	It("should count both global and mesh scoped resources", func() {
		// given
		err := resManager.Create(context.Background(), system.NewZoneResource(), core_store.CreateByKey("zone-1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		err = resManager.Create(context.Background(), core_mesh.NewMeshResource(), core_store.CreateByKey("mesh-1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		err = resManager.Create(
			context.Background(),
			&core_mesh.DataplaneResource{Spec: samples.Dataplane},
			core_store.CreateByKey("dp-1", "mesh-1"),
		)
		Expect(err).ToNot(HaveOccurred())

		err = resManager.Create(
			context.Background(),
			&core_mesh.TrafficPermissionResource{Spec: samples.TrafficPermission},
			core_store.CreateByKey("tp-1", "mesh-1"),
		)
		Expect(err).ToNot(HaveOccurred())

		err = resManager.Create(context.Background(), core_mesh.NewMeshResource(), core_store.CreateByKey("mesh-2", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		err = resManager.Create(
			context.Background(),
			&core_mesh.TrafficPermissionResource{Spec: samples.TrafficPermission},
			core_store.CreateByKey("tp-2", "mesh-2"),
		)
		Expect(err).ToNot(HaveOccurred())

		// when
		// trigger the resyncer
		nowMtx.Lock()
		now = now.Add(1 * time.Minute)
		nowMtx.Unlock()
		tickCh <- now
		// wait for the counter
		time.Sleep(1 * time.Second)

		// then
		Expect(findGauge("Zone").GetValue()).To(Equal(float64(1)))
		Expect(findGauge("Mesh").GetValue()).To(Equal(float64(2)))
		Expect(findGauge("Dataplane").GetValue()).To(Equal(float64(1)))
		Expect(findGauge("TrafficPermission").GetValue()).To(Equal(float64(2)))
	})
})
