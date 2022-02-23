package postgres_test

import (
	"fmt"
	"time"

	"cirello.io/pglock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/runtime/component"
	common_postgres "github.com/kumahq/kuma/pkg/plugins/common/postgres"
	leader_postgres "github.com/kumahq/kuma/pkg/plugins/leader/postgres"
	"github.com/kumahq/kuma/pkg/test"
	test_postgres "github.com/kumahq/kuma/pkg/test/store/postgres"
)

var _ = Describe("postgresLeaderElector", func() {
	var c test_postgres.PostgresContainer
	var electors map[string]component.LeaderElector
	BeforeEach(func() {
		c = test_postgres.PostgresContainer{WithTLS: true}
		Expect(c.Start()).To(Succeed())
		cfg, err := c.Config()
		Expect(err).ToNot(HaveOccurred())
		sql, err := common_postgres.ConnectToDb(*cfg)
		Expect(err).ToNot(HaveOccurred())

		createElector := func(name string) component.LeaderElector {
			client, err := pglock.New(sql,
				pglock.WithLeaseDuration(500*time.Millisecond),
				pglock.WithHeartbeatFrequency(100*time.Millisecond),
				pglock.WithOwner(name),
			)
			Expect(err).ToNot(HaveOccurred())
			_ = client.CreateTable() // ignore error
			elector := leader_postgres.NewPostgresLeaderElector(client)
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

	It("should elect only one leader", test.Within(30*time.Second, func() {
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
