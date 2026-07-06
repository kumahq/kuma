package catalog_test

import (
	"context"
	"errors"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	system_proto "github.com/kumahq/kuma/v3/api/system/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v3/pkg/core/runtime/component"
	"github.com/kumahq/kuma/v3/pkg/intercp/catalog"
	core_metrics "github.com/kumahq/kuma/v3/pkg/metrics"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
	test_metrics "github.com/kumahq/kuma/v3/pkg/test/metrics"
)

var _ = Describe("Heartbeats", func() {
	var heartbeatComponent component.Component
	var stopCh chan struct{}

	var pingClient *staticPingClient
	var c catalog.Catalog
	var metrics core_metrics.Metrics

	currentInstance := catalog.Instance{
		Id:          "instance-1",
		Address:     "10.10.10.1",
		InterCpPort: 5679,
		Leader:      false,
	}

	BeforeEach(func() {
		store := memory.NewStore()
		resManager := manager.NewResourceManager(store)
		c = catalog.NewConfigCatalog(resManager)
		pingClient = &staticPingClient{
			leader: true,
		}
		pc := pingClient // copy pointer to get rid of data race

		var err error
		metrics, err = core_metrics.NewMetrics("")
		Expect(err).ToNot(HaveOccurred())
		heartbeatComponent, err = catalog.NewHeartbeatComponent(
			c,
			currentInstance,
			10*time.Millisecond,
			func(serverURL string) (system_proto.InterCpPingServiceClient, error) {
				pc.SetServerURL(serverURL)
				return pc, nil
			},
			metrics,
		)
		Expect(err).ToNot(HaveOccurred())

		stopCh = make(chan struct{})
		go func() {
			defer GinkgoRecover()
			err := heartbeatComponent.Start(stopCh)
			Expect(err).ToNot(HaveOccurred())
		}()
	})

	AfterEach(func() {
		close(stopCh)
	})

	It("should connect to a leader once we have a leader in the catalog", func() {
		// given
		instances := []catalog.Instance{
			{
				Id:          "instance-leader",
				Address:     "192.168.0.1",
				InterCpPort: 1234,
				Leader:      true,
			},
		}

		// when
		_, err := c.Replace(context.Background(), instances)
		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func(g Gomega) {
			received := pingClient.Received()
			g.Expect(received).ToNot(BeNil())
			g.Expect(received.InstanceId).To(Equal(currentInstance.Id))
			g.Expect(received.Address).To(Equal(currentInstance.Address))
			g.Expect(received.InterCpPort).To(Equal(uint32(currentInstance.InterCpPort)))
		}, "10s", "100ms").Should(Succeed())
	})

	It("should reconnect to a leader when there is a leader change", func() {
		// given
		instances := []catalog.Instance{
			{
				Id:          "instance-leader",
				Address:     "192.168.0.1",
				InterCpPort: 1234,
				Leader:      true,
			},
		}
		_, err := c.Replace(context.Background(), instances)
		Expect(err).ToNot(HaveOccurred())
		Eventually(func(g Gomega) {
			received := pingClient.Received()
			g.Expect(received).ToNot(BeNil())
			g.Expect(pingClient.ServerURL()).To(Equal("grpcs://192.168.0.1:1234"))
		}, "10s", "100ms").Should(Succeed())

		// when
		pingClient.SetLeaderResponse(false)

		// and
		instances = []catalog.Instance{
			{
				Id:          "instance-leader-2",
				Address:     "192.168.0.2",
				InterCpPort: 1234,
				Leader:      true,
			},
		}
		_, err = c.Replace(context.Background(), instances)
		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func(g Gomega) {
			received := pingClient.Received()
			g.Expect(received).ToNot(BeNil())
			g.Expect(pingClient.ServerURL()).To(Equal("grpcs://192.168.0.2:1234"))
		}, "10s", "100ms").Should(Succeed())
	})

	It("should report unreachable leader as connectivity failure, not a leader change", func() {
		// given a stable leader in the catalog
		instances := []catalog.Instance{
			{
				Id:          "instance-leader",
				Address:     "192.168.0.1",
				InterCpPort: 1234,
				Leader:      true,
			},
		}
		_, err := c.Replace(context.Background(), instances)
		Expect(err).ToNot(HaveOccurred())
		Eventually(func(g Gomega) {
			g.Expect(pingClient.ServerURL()).To(Equal("grpcs://192.168.0.1:1234"))
		}, "10s", "100ms").Should(Succeed())

		// when the leader becomes unreachable
		pingClient.SetError(status.Error(codes.Unavailable, "connection refused"))

		// then failures are recorded as connectivity problems
		Eventually(func(g Gomega) {
			metric := test_metrics.FindMetric(metrics, "component_heartbeat_failures")
			g.Expect(metric).ToNot(BeNil())
			g.Expect(metric.Counter.GetValue()).To(BeNumerically(">", 0))
		}, "10s", "100ms").Should(Succeed())

		// and the stable leader connection is never replaced while unreachable
		Consistently(func(g Gomega) {
			g.Expect(pingClient.ServerURL()).To(Equal("grpcs://192.168.0.1:1234"))
		}, "1s", "100ms").Should(Succeed())
	})

	It("should record failures when the catalog cannot be read", func() {
		failingMetrics, err := core_metrics.NewMetrics("")
		Expect(err).ToNot(HaveOccurred())
		failingComponent, err := catalog.NewHeartbeatComponent(
			failingCatalog{err: errors.New("catalog read failed")},
			currentInstance,
			10*time.Millisecond,
			func(string) (system_proto.InterCpPingServiceClient, error) {
				return pingClient, nil
			},
			failingMetrics,
		)
		Expect(err).ToNot(HaveOccurred())

		failingStopCh := make(chan struct{})
		defer close(failingStopCh)
		go func() {
			defer GinkgoRecover()
			err := failingComponent.Start(failingStopCh)
			Expect(err).ToNot(HaveOccurred())
		}()

		Eventually(func(g Gomega) {
			metric := test_metrics.FindMetric(failingMetrics, "component_heartbeat_failures")
			g.Expect(metric).ToNot(BeNil())
			g.Expect(metric.Counter.GetValue()).To(BeNumerically(">", 0))
		}, "10s", "100ms").Should(Succeed())
	})

	It("should not send heartbeat if the instance is a leader", func() {
		// given
		instance := currentInstance
		instance.Leader = true
		_, err := c.Replace(context.Background(), []catalog.Instance{instance})
		Expect(err).ToNot(HaveOccurred())

		// then
		Consistently(func(g Gomega) {
			g.Expect(pingClient.ServerURL()).To(BeEmpty())
			g.Expect(pingClient.Received()).To(BeNil())
		}, "1s", "100ms").Should(Succeed())
	})
})

type failingCatalog struct {
	err error
}

var _ catalog.Catalog = failingCatalog{}

func (f failingCatalog) Instances(context.Context) ([]catalog.Instance, error) {
	return nil, f.err
}

func (f failingCatalog) Replace(context.Context, []catalog.Instance) (bool, error) {
	return false, f.err
}

func (f failingCatalog) ReplaceLeader(context.Context, catalog.Instance) error {
	return f.err
}

func (f failingCatalog) DropLeader(context.Context, catalog.Instance) error {
	return f.err
}

type staticPingClient struct {
	received  *system_proto.PingRequest
	serverURL string
	leader    bool
	err       error
	sync.Mutex
}

var _ system_proto.InterCpPingServiceClient = &staticPingClient{}

func (s *staticPingClient) Ping(ctx context.Context, in *system_proto.PingRequest, opts ...grpc.CallOption) (*system_proto.PingResponse, error) {
	s.Lock()
	defer s.Unlock()
	s.received = in
	if s.err != nil {
		return nil, s.err
	}
	return &system_proto.PingResponse{
		Leader: s.leader,
	}, nil
}

func (s *staticPingClient) SetError(err error) {
	s.Lock()
	defer s.Unlock()
	s.err = err
}

func (s *staticPingClient) Received() *system_proto.PingRequest {
	s.Lock()
	defer s.Unlock()
	return s.received
}

func (s *staticPingClient) SetLeaderResponse(leader bool) {
	s.Lock()
	defer s.Unlock()
	s.leader = leader
}

func (s *staticPingClient) SetServerURL(serverURL string) {
	s.Lock()
	defer s.Unlock()
	s.serverURL = serverURL
}

func (s *staticPingClient) ServerURL() string {
	s.Lock()
	defer s.Unlock()
	return s.serverURL
}
