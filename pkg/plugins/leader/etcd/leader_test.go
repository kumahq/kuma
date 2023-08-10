package etcd_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/runtime/component"
	common_etcd "github.com/kumahq/kuma/pkg/plugins/common/etcd"
	leader_etcd "github.com/kumahq/kuma/pkg/plugins/leader/etcd"
	"github.com/kumahq/kuma/pkg/test"
	test_etcd "github.com/kumahq/kuma/pkg/test/store/etcd"
)

var _ = Describe("etcdLeaderElector", func() {
	var c test_etcd.EtcdContainer
	var electors map[string]component.LeaderElector
	BeforeEach(func() {
		c = test_etcd.EtcdContainer{}
		Expect(c.Start()).To(Succeed())
		cfg, err := c.Config()
		Expect(err).ToNot(HaveOccurred())
		client, err := common_etcd.NewClient(cfg)
		Expect(err).ToNot(HaveOccurred())

		createElector := func(name string) component.LeaderElector {
			elector := leader_etcd.NewEtcdLeaderElector(name, client)
			Expect(err).ToNot(HaveOccurred())
			return elector
		}

		electors = map[string]component.LeaderElector{}
		for i := 1; i <= 3; i++ {
			name := fmt.Sprintf("elector-%d", i)
			electors[name] = createElector(name)
		}
	})
	AfterEach(func() {
		Expect(c.Stop()).To(Succeed())
	})

	It("should elect only one leader", test.Within(60*time.Second, func() {
		// given
		acquiredLeaderCh := make(chan string)
		lostLeaderCh := make(chan string)
		for name, elector := range electors {
			electorName := name
			elector.AddCallbacks(component.LeaderCallbacks{
				OnStartedLeading: func() {
					acquiredLeaderCh <- electorName
				},
				OnStoppedLeading: func() {
					lostLeaderCh <- electorName
				},
			})
		}

		// when electors are started
		electorStops := map[string]chan struct{}{}
		for name, elector := range electors {
			stopCh := make(chan struct{})
			go elector.Start(stopCh)
			electorStops[name] = stopCh
		}

		// then lead is selected
		lead := <-acquiredLeaderCh
		Expect(electors[lead].IsLeader()).To(BeTrue())
		// and other electors are not leaders
		for name := range electors {
			if name != lead {
				Expect(electors[name].IsLeader()).To(BeFalse())
			}
		}

		// when leader is killed
		close(electorStops[lead])

		// then leader is lost
		lostLead := <-lostLeaderCh
		Expect(electors[lostLead].IsLeader()).To(BeFalse())

		// and new leader is selected
		newLead := <-acquiredLeaderCh
		Expect(newLead).ToNot(Equal(lead))
		Expect(electors[newLead].IsLeader()).To(BeTrue())
	}))
})
