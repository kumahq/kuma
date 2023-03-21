package postgres

import (
	"context"
	"fmt"
	postgres_events "github.com/kumahq/kuma/pkg/plugins/resources/postgres/events"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	postgres_config "github.com/kumahq/kuma/pkg/config/plugins/resources/postgres"
	"github.com/kumahq/kuma/pkg/core/plugins"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	kuma_events "github.com/kumahq/kuma/pkg/events"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	test_postgres "github.com/kumahq/kuma/pkg/test/store/postgres"
	"github.com/kumahq/kuma/pkg/util/channels"
)

var _ = Describe("Events", func() {
	var cfg postgres_config.PostgresStoreConfig

	BeforeEach(func() {
		c, err := c.Config(test_postgres.WithRandomDb)
		Expect(err).ToNot(HaveOccurred())
		cfg = *c
		ver, err := migrateDb(cfg)
		Expect(err).ToNot(HaveOccurred())
		Expect(ver).To(Equal(plugins.DbVersion(1677096751)))
	})

	DescribeTable("should receive a notification from pq listener",
		func(driverName string) {
			// given
			listenerStopCh, listenerErrCh, eventBusStopCh := setupChannels()
			defer close(eventBusStopCh)
			defer close(listenerErrCh)
			listener := setupListeners(cfg, driverName, listenerErrCh, listenerStopCh)
			go triggerNotifications(cfg, listenerStopCh)

			// when
			event, err := listener.Recv(eventBusStopCh)

			// then
			Expect(err).To(Not(HaveOccurred()))
			resourceChanged := event.(kuma_events.ResourceChangedEvent)
			Expect(resourceChanged.Operation).To(Equal(kuma_events.Create))
			Expect(resourceChanged.Type).To(Equal(model.ResourceType("Mesh")))

			// and shutdown gracefully
			close(listenerStopCh)
			Eventually(listenerErrorChannelClosesWithoutErrors(listenerErrCh), "5s", "10ms").Should(BeTrue())
		},
		Entry("When using pq", postgres_config.DriverNamePq),
		Entry("When using pgx", postgres_config.DriverNamePgx),
	)
})

func setupChannels() (chan struct{}, chan error, chan struct{}) {
	listenerStopCh := make(chan struct{})
	listenerErrCh := make(chan error)
	eventBusStopCh := make(chan struct{})

	return listenerStopCh, listenerErrCh, eventBusStopCh
}

func setupStore(cfg postgres_config.PostgresStoreConfig) store.ResourceStore {
	metrics, err := core_metrics.NewMetrics("Standalone")
	Expect(err).ToNot(HaveOccurred())
	pStore, err := NewStore(metrics, cfg)
	Expect(err).ToNot(HaveOccurred())
	return pStore
}

func setupListeners(cfg postgres_config.PostgresStoreConfig, driverName string, listenerErrCh chan error, listenerStopCh chan struct{}) kuma_events.Listener {
	cfg.DriverName = driverName
	eventsBus := kuma_events.NewEventBus()
	listener := eventsBus.New()
	l := postgres_events.NewListener(cfg, eventsBus)
	go func() {
		listenerErrCh <- l.Start(listenerStopCh)
	}()

	return listener
}

func triggerNotifications(cfg postgres_config.PostgresStoreConfig, listenerStopCh chan struct{}) {
	pStore := setupStore(cfg)
	defer GinkgoRecover()
	for i := 0; !channels.IsClosed(listenerStopCh); i++ {
		err := pStore.Create(context.Background(), mesh.NewMeshResource(), store.CreateByKey(fmt.Sprintf("mesh-%d", i), ""))
		Expect(err).ToNot(HaveOccurred())
	}
}

func listenerErrorChannelClosesWithoutErrors(listenerErrCh chan error) func() bool {
	return func() bool {
		select {
		case err := <-listenerErrCh:
			Expect(err).ToNot(HaveOccurred())
			return true
		default:
			return false
		}

	}
}
