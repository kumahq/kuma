package postgres

import (
	"context"
	postgres_config "github.com/kumahq/kuma/pkg/config/plugins/resources/postgres"
	"github.com/kumahq/kuma/pkg/core/plugins"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	kuma_events "github.com/kumahq/kuma/pkg/events"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	postgres_events "github.com/kumahq/kuma/pkg/plugins/resources/postgres/events"
	test_postgres "github.com/kumahq/kuma/pkg/test/store/postgres"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("Events", func() {
	var cfg postgres_config.PostgresStoreConfig
	var listenerErrCh chan error
	var listenerStopCh chan struct{}
	var eventBusStopCh chan struct{}

	BeforeEach(func() {
		listenerStopCh = make(chan struct{})
		listenerErrCh = make(chan error)
		eventBusStopCh = make(chan struct{})
		c, err := c.Config(test_postgres.WithRandomDb)
		Expect(err).ToNot(HaveOccurred())
		cfg = *c
		ver, err := migrateDb(cfg)
		Expect(err).ToNot(HaveOccurred())
		Expect(ver).To(Equal(plugins.DbVersion(1677096751)))
	})

	DescribeTable("should receive a notification from pq listener",
		func(usePgx bool) {
			eventsBus := kuma_events.NewEventBus()
			listener := eventsBus.New()
			l := postgres_events.NewListener(cfg, eventsBus, usePgx)
			go func() {
				listenerErrCh <- l.Start(listenerStopCh)
			}()

			metrics, err := core_metrics.NewMetrics("Standalone")
			Expect(err).ToNot(HaveOccurred())
			pStore, err := NewStore(metrics, cfg)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(1*time.Second)
			err = pStore.Create(context.Background(), mesh.NewMeshResource())
			Expect(err).ToNot(HaveOccurred())

			event, err := listener.Recv(eventBusStopCh)
			Expect(err).To(Not(HaveOccurred()))

			resourceChanged := event.(kuma_events.ResourceChangedEvent)
			Expect(resourceChanged.Operation).To(Equal(kuma_events.Create))
			Expect(resourceChanged.Type).To(Equal(model.ResourceType("Mesh")))

			close(listenerStopCh)
			close(eventBusStopCh)
			Eventually(func() bool {
				select {
				case err := <-listenerErrCh:
					Expect(err).ToNot(HaveOccurred())
					return true
				default:
					return false
				}
			}, "5s", "10ms").Should(BeTrue())
		},
		Entry("When using pq", false),
		Entry("When using pgx", true),
	)
})
