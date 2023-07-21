package catalog_test

import (
	"context"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/intercp/catalog"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
)

var _ = Describe("Heartbeats", func() {
	var heartbeatComponent component.Component
	var stopCh chan struct{}

	var pingClient *staticPingClient
	var c catalog.Catalog

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

		metrics, err := core_metrics.NewMetrics("")
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

type staticPingClient struct {
	received  *system_proto.PingRequest
	serverURL string
	leader    bool
	sync.Mutex
}

var _ system_proto.InterCpPingServiceClient = &staticPingClient{}

func (s *staticPingClient) Ping(ctx context.Context, in *system_proto.PingRequest, opts ...grpc.CallOption) (*system_proto.PingResponse, error) {
	s.Lock()
	defer s.Unlock()
	s.received = in
	return &system_proto.PingResponse{
		Leader: s.leader,
	}, nil
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
