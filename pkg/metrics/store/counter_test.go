package metrics_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	io_prometheus_client "github.com/prometheus/client_model/go"

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
	"github.com/kumahq/kuma/pkg/multitenant"
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

	start := time.Now()
	var tickCh chan time.Time

	BeforeEach(func() {
		var err error

		eventCh = make(chan events.Event)
		stop = make(chan struct{})
		tickCh = make(chan time.Time)

		metrics, err = core_metrics.NewMetrics("Zone")
		Expect(err).ToNot(HaveOccurred())

		memoryStore := store_memory.NewStore()
		store, err = metrics_store.NewMeteredStore(memoryStore, metrics)
		Expect(err).ToNot(HaveOccurred())

		resManager = manager.NewResourceManager(store)

		counterTicker := time.NewTicker(500 * time.Millisecond)
		counter, err := metrics_store.NewStoreCounter(resManager, metrics, multitenant.SingleTenant)
		Expect(err).ToNot(HaveOccurred())

		resyncer := insights.NewResyncer(&insights.Config{
			MinResyncInterval:  5 * time.Second,
			FullResyncInterval: 5 * time.Second,
			ResourceManager:    resManager,
			EventReaderFactory: &test_insights.TestEventReaderFactory{Reader: &test_insights.TestEventReader{Ch: eventCh}},
			Tick: func(d time.Duration) <-chan time.Time {
				return tickCh
			},
			Registry:            registry.Global(),
			TenantFn:            multitenant.SingleTenant,
			EventBufferCapacity: 10,
			EventProcessors:     1,
			Metrics:             metrics,
			Extensions:          context.Background(),
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
		tickCh <- start.Add(time.Minute)
		// wait for the counter
		Eventually(func(g Gomega) {
			// then
			g.Expect(findGauge("Zone").GetValue()).To(Equal(float64(1)))
			g.Expect(findGauge("Mesh").GetValue()).To(Equal(float64(2)))
			g.Expect(findGauge("Dataplane").GetValue()).To(Equal(float64(1)))
			g.Expect(findGauge("TrafficPermission").GetValue()).To(Equal(float64(2)))
		}).Should(Succeed())
	})
})
