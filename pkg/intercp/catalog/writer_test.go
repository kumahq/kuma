package catalog_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/intercp/catalog"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/util/channels"
)

var _ = Describe("Writer", func() {
	var c catalog.Catalog
	var closeCh chan struct{}
	var heartbeatCancelFunc context.CancelFunc

	leader := catalog.Instance{
		Id:          "instance-2",
		Address:     "192.168.0.2",
		InterCpPort: 1234,
		Leader:      true,
	}

	instance := catalog.Instance{
		Id:          "instance-1",
		Address:     "192.168.0.1",
		InterCpPort: 1234,
	}

	BeforeEach(func() {
		store := memory.NewStore()
		resManager := manager.NewResourceManager(store)
		c = catalog.NewConfigCatalog(resManager)
		heartbeats := catalog.NewHeartbeats()
		closeCh = make(chan struct{})
		metrics, err := core_metrics.NewMetrics("")
		Expect(err).ToNot(HaveOccurred())
		writer, err := catalog.NewWriter(c, heartbeats, leader, 100*time.Millisecond, metrics)
		Expect(err).ToNot(HaveOccurred())
		go func() {
			defer GinkgoRecover()
			Expect(writer.Start(closeCh)).To(Succeed())
		}()

		ctx, fn := context.WithCancel(context.Background())
		heartbeatCancelFunc = fn
		go func() {
			t := time.NewTicker(10 * time.Millisecond)
			for {
				select {
				case <-t.C:
					heartbeats.Add(instance)
				case <-ctx.Done():
					return
				}
			}
		}()
	})

	AfterEach(func() {
		if !channels.IsClosed(closeCh) {
			close(closeCh)
		}
		heartbeatCancelFunc()
	})

	It("should write to catalog", func() {
		Eventually(func(g Gomega) {
			instances, err := c.Instances(context.Background())
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(instances).To(HaveLen(2))
			g.Expect(instances[0]).To(Equal(instance))
			g.Expect(instances[1]).To(Equal(leader))
		}, "10s", "100ms").Should(Succeed())
	})

	It("should remove instance from the catalog once hearth-beating stop", func() {
		// given 2 instances in catalog
		Eventually(func(g Gomega) {
			instances, err := c.Instances(context.Background())
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(instances).To(HaveLen(2))
		}, "10s", "100ms").Should(Succeed())

		// when
		heartbeatCancelFunc()

		// then
		Eventually(func(g Gomega) {
			instances, err := c.Instances(context.Background())
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(instances).To(HaveLen(1))
			g.Expect(instances[0]).To(Equal(leader))
		}, "10s", "100ms").Should(Succeed())
	})

	It("should deregister leader after stopping the component", func() {
		// given
		Eventually(func(g Gomega) {
			instance, err := catalog.InstanceOfID(context.Background(), c, leader.Id)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(instance.Leader).To(BeTrue())
		}, "10s", "100ms").Should(Succeed())

		// when
		close(closeCh)

		// then
		Eventually(func(g Gomega) {
			instance, err := catalog.InstanceOfID(context.Background(), c, leader.Id)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(instance.Leader).To(BeFalse())
		}, "10s", "100ms").Should(Succeed())
	})
})
