package postgres_test

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	config_postgres "github.com/kumahq/kuma/v2/pkg/config/plugins/resources/postgres"
	"github.com/kumahq/kuma/v2/pkg/core"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/store"
	"github.com/kumahq/kuma/v2/pkg/core/runtime/component"
	kuma_events "github.com/kumahq/kuma/v2/pkg/events"
	core_metrics "github.com/kumahq/kuma/v2/pkg/metrics"
	common_postgres "github.com/kumahq/kuma/v2/pkg/plugins/common/postgres"
	"github.com/kumahq/kuma/v2/pkg/plugins/resources/postgres"
	"github.com/kumahq/kuma/v2/pkg/plugins/resources/postgres/config"
	events_postgres "github.com/kumahq/kuma/v2/pkg/plugins/resources/postgres/events"
	"github.com/kumahq/kuma/v2/pkg/test"
)

var _ = Describe("Events", func() {
	DescribeTable("should receive a notification from pq listener",
		func(driverName string) {
			cfg, err := c.Config()
			Expect(err).ToNot(HaveOccurred())
			ver, err := postgres.MigrateDb(cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(ver).To(Equal(dbVersion))
			// given
			listenerStopCh, listenerErrCh, eventBusStopCh, storeErrCh := setupChannels()
			defer close(eventBusStopCh)
			listener := setupListeners(cfg, driverName, listenerErrCh, listenerStopCh, 5*time.Second)
			pStore := setupStore(cfg, driverName)
			storeCtx, stopStore := context.WithCancel(context.Background())
			storeDoneCh := triggerNotifications(storeCtx, pStore, storeErrCh)
			defer func() {
				stopStore()
				Eventually(storeDoneCh, "5s", "10ms").Should(BeClosed())
			}()

			eventsChan := listener.Recv()

			// when
			var resourceChanged kuma_events.ResourceChangedEvent
			Eventually(eventsChan, "1s").Should(Receive(&resourceChanged))
			Expect(eventBusStopCh).ToNot(BeClosed())

			// then
			Expect(resourceChanged.Operation).To(Equal(kuma_events.Create))
			Expect(resourceChanged.Type).To(Equal(model.ResourceType("Mesh")))

			// and shutdown gracefully
			close(listenerStopCh)
			Eventually(channelClosesWithoutErrors(listenerErrCh), "5s", "10ms").Should(BeTrue())
		},
		Entry("When using pgx", config_postgres.DriverNamePgx),
	)

	DescribeTable("should continue handling notification after postgres recovery",
		func(driverName string) {
			// given
			cfg, err := c.Config()
			Expect(err).ToNot(HaveOccurred())
			proxy, err := test.NewTCPProxy(net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)))
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				Expect(proxy.Stop()).To(Succeed())
			}()
			cfg.Host = proxy.Host()
			cfg.Port = proxy.Port()

			ver, err := postgres.MigrateDb(cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(ver).To(Equal(dbVersion))
			listenerStopCh, listenerErrCh, eventBusStopCh, storeErrCh := setupChannels()
			defer close(eventBusStopCh)
			listener := setupListeners(cfg, driverName, listenerErrCh, listenerStopCh, 100*time.Millisecond)
			pStore := setupStore(cfg, driverName)
			storeCtx, stopStore := context.WithCancel(context.Background())
			storeDoneCh := triggerNotifications(storeCtx, pStore, storeErrCh)
			defer func() {
				stopStore()
				Eventually(storeDoneCh, "5s", "10ms").Should(BeClosed())
				close(listenerStopCh)
				Eventually(channelClosesWithoutErrors(listenerErrCh), "5s", "10ms").Should(BeTrue())
			}()

			eventsChan := listener.Recv()

			// when
			var resourceChanged kuma_events.ResourceChangedEvent
			Eventually(eventsChan, "1s").Should(Receive(&resourceChanged))
			Expect(eventBusStopCh).ToNot(BeClosed())

			Expect(resourceChanged.Operation).To(Equal(kuma_events.Create))
			Expect(resourceChanged.Type).To(Equal(model.ResourceType("Mesh")))

			Expect(proxy.Stop()).To(Succeed())
			Eventually(storeErrCh, "5s", "10ms").Should(Receive())

			Expect(proxy.Start()).To(Succeed())
			waitForPostgres(cfg)

			Eventually(eventsChan, "5s").Should(Receive(&resourceChanged))
			Expect(eventBusStopCh).ToNot(BeClosed())

			Expect(resourceChanged.Operation).To(Equal(kuma_events.Create))
			Expect(resourceChanged.Type).To(Equal(model.ResourceType("Mesh")))
		},
		Entry("When using pgx", config_postgres.DriverNamePgx),
	)
})

func setupChannels() (chan struct{}, chan error, chan struct{}, chan error) {
	listenerStopCh := make(chan struct{})
	listenerErrCh := make(chan error)
	eventBusStopCh := make(chan struct{})
	storeErrCh := make(chan error, 1)

	return listenerStopCh, listenerErrCh, eventBusStopCh, storeErrCh
}

func setupStore(cfg config_postgres.PostgresStoreConfig, driverName string) store.ResourceStore {
	metrics, err := core_metrics.NewMetrics("Zone")
	Expect(err).ToNot(HaveOccurred())
	var pStore store.ResourceStore
	if driverName == "pgx" {
		cfg.DriverName = config_postgres.DriverNamePgx
		pStore, err = postgres.NewPgxStore(metrics, cfg, config.NoopPgxConfigCustomizationFn)
	}
	Expect(err).ToNot(HaveOccurred())
	return pStore
}

func setupListeners(
	cfg config_postgres.PostgresStoreConfig,
	driverName string,
	listenerErrCh chan error,
	listenerStopCh chan struct{},
	restartBackoff time.Duration,
) kuma_events.Listener {
	cfg.DriverName = driverName
	metrics, err := core_metrics.NewMetrics("")
	Expect(err).ToNot(HaveOccurred())
	eventsBus, err := kuma_events.NewEventBus(20, metrics)
	Expect(err).ToNot(HaveOccurred())
	listener := eventsBus.Subscribe()
	l := events_postgres.NewListener(cfg, eventsBus)
	resilientListener := component.NewResilientComponent(core.Log.WithName("postgres-event-listener-component"), l, restartBackoff, 1*time.Minute)
	go func() {
		listenerErrCh <- resilientListener.Start(listenerStopCh)
	}()

	return listener
}

func triggerNotifications(ctx context.Context, pStore store.ResourceStore, storeErrCh chan<- error) <-chan struct{} {
	doneCh := make(chan struct{})
	go func() {
		defer GinkgoRecover()
		defer close(doneCh)

		for i := 0; ; i++ {
			select {
			case <-ctx.Done():
				return
			default:
			}

			err := pStore.Create(ctx, mesh.NewMeshResource(), store.CreateByKey(fmt.Sprintf("mesh-%d", i), ""))
			if err != nil {
				select {
				case storeErrCh <- err:
				default:
				}
				select {
				case <-ctx.Done():
					return
				case <-time.After(10 * time.Millisecond):
				}
			}
		}
	}()
	return doneCh
}

func waitForPostgres(cfg config_postgres.PostgresStoreConfig) {
	Eventually(func() error {
		db, err := common_postgres.ConnectToDb(cfg)
		if err != nil {
			return err
		}
		return db.Close()
	}, "10s", "100ms").Should(Succeed())
}

func channelClosesWithoutErrors(listenerErrCh chan error) func() bool {
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
