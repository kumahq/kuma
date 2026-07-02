package postgres_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	config_postgres "github.com/kumahq/kuma/v3/pkg/config/plugins/resources/postgres"
	"github.com/kumahq/kuma/v3/pkg/core"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/core/runtime/component"
	kuma_events "github.com/kumahq/kuma/v3/pkg/events"
	core_metrics "github.com/kumahq/kuma/v3/pkg/metrics"
	common_postgres "github.com/kumahq/kuma/v3/pkg/plugins/common/postgres"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/postgres"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/postgres/config"
	events_postgres "github.com/kumahq/kuma/v3/pkg/plugins/resources/postgres/events"
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
			listener := setupListeners(cfg, driverName, listenerErrCh, listenerStopCh)
			storeCtx, stopStore := context.WithCancel(context.Background())
			storeDoneCh := triggerNotifications(storeCtx, cfg, driverName, storeErrCh)
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
			proxy, err := newTCPProxy(net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)))
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
			storeCtx, stopStore := context.WithCancel(context.Background())
			storeDoneCh := triggerNotifications(storeCtx, cfg, driverName, storeErrCh)
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
	restartBackoff ...time.Duration,
) kuma_events.Listener {
	cfg.DriverName = driverName
	metrics, err := core_metrics.NewMetrics("")
	Expect(err).ToNot(HaveOccurred())
	eventsBus, err := kuma_events.NewEventBus(20, metrics)
	Expect(err).ToNot(HaveOccurred())
	listener := eventsBus.Subscribe()
	l := events_postgres.NewListener(cfg, eventsBus)
	backoff := 5 * time.Second
	if len(restartBackoff) > 0 {
		backoff = restartBackoff[0]
	}
	resilientListener := component.NewResilientComponent(core.Log.WithName("postgres-event-listener-component"), l, backoff, 1*time.Minute)
	go func() {
		listenerErrCh <- resilientListener.Start(listenerStopCh)
	}()

	return listener
}

func triggerNotifications(ctx context.Context, cfg config_postgres.PostgresStoreConfig, driverName string, storeErrCh chan<- error) <-chan struct{} {
	doneCh := make(chan struct{})
	go func() {
		defer GinkgoRecover()
		defer close(doneCh)

		pStore := setupStore(cfg, driverName)
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

type tcpProxy struct {
	target string

	mu         sync.Mutex
	listenAddr string
	listener   net.Listener
	conns      map[net.Conn]struct{}
}

func newTCPProxy(target string) (*tcpProxy, error) {
	proxy := &tcpProxy{
		target:     target,
		listenAddr: "127.0.0.1:0",
		conns:      map[net.Conn]struct{}{},
	}
	if err := proxy.Start(); err != nil {
		return nil, err
	}
	return proxy, nil
}

func (p *tcpProxy) Host() string {
	host, _, err := net.SplitHostPort(p.listenAddr)
	Expect(err).ToNot(HaveOccurred())
	return host
}

func (p *tcpProxy) Port() int {
	_, port, err := net.SplitHostPort(p.listenAddr)
	Expect(err).ToNot(HaveOccurred())
	parsed, err := strconv.Atoi(port)
	Expect(err).ToNot(HaveOccurred())
	return parsed
}

func (p *tcpProxy) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.listener != nil {
		return nil
	}
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", p.listenAddr)
	if err != nil {
		return err
	}
	p.listenAddr = listener.Addr().String()
	p.listener = listener
	go p.accept(listener)
	return nil
}

func (p *tcpProxy) Stop() error {
	p.mu.Lock()
	listener := p.listener
	p.listener = nil
	conns := make([]net.Conn, 0, len(p.conns))
	for conn := range p.conns {
		conns = append(conns, conn)
	}
	p.conns = map[net.Conn]struct{}{}
	p.mu.Unlock()

	for _, conn := range conns {
		_ = conn.Close()
	}
	if listener == nil {
		return nil
	}
	return listener.Close()
}

func (p *tcpProxy) accept(listener net.Listener) {
	for {
		source, err := listener.Accept()
		if err != nil {
			return
		}
		go p.forward(source)
	}
}

func (p *tcpProxy) forward(source net.Conn) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	target, err := (&net.Dialer{}).DialContext(ctx, "tcp", p.target)
	if err != nil {
		_ = source.Close()
		return
	}

	if !p.track(source) {
		_ = source.Close()
		_ = target.Close()
		return
	}
	if !p.track(target) {
		p.untrack(source)
		_ = source.Close()
		_ = target.Close()
		return
	}
	defer func() {
		p.untrack(source)
		p.untrack(target)
		_ = source.Close()
		_ = target.Close()
	}()

	doneCh := make(chan struct{}, 2)
	go func() {
		_, _ = io.Copy(target, source)
		doneCh <- struct{}{}
	}()
	go func() {
		_, _ = io.Copy(source, target)
		doneCh <- struct{}{}
	}()

	<-doneCh
	_ = source.Close()
	_ = target.Close()
	<-doneCh
}

func (p *tcpProxy) track(conn net.Conn) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.listener == nil {
		_ = conn.Close()
		return false
	}
	p.conns[conn] = struct{}{}
	return true
}

func (p *tcpProxy) untrack(conn net.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.conns, conn)
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
