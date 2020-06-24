// +build integration

package postgres_test

import (
	"fmt"
	"time"

	"cirello.io/pglock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/pkg/config"
	"github.com/Kong/kuma/pkg/config/plugins/resources/postgres"
	"github.com/Kong/kuma/pkg/core/runtime/component"
	common_postgres "github.com/Kong/kuma/pkg/plugins/common/postgres"
	leader_postgres "github.com/Kong/kuma/pkg/plugins/leader/postgres"
)

var _ = Describe("postgresLeaderElector", func() {

	var electors map[string]component.LeaderElector

	BeforeSuite(func() {
		// wait for postgres
		cfg := postgres.PostgresStoreConfig{}
		err := config.Load("", &cfg)
		Expect(err).ToNot(HaveOccurred())
		Eventually(func() error {
			_, err := common_postgres.ConnectToDb(cfg)
			return err
		}, "5s", "100ms").ShouldNot(HaveOccurred())
	})

	BeforeEach(func() {
		cfg := postgres.PostgresStoreConfig{}
		err := config.Load("", &cfg)
		Expect(err).ToNot(HaveOccurred())

		dbName, err := common_postgres.CreateRandomDb(cfg)
		Expect(err).ToNot(HaveOccurred())
		cfg.DbName = dbName

		sql, err := common_postgres.ConnectToDb(cfg)
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

	It("should elect only one leader", func(done Done) {
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

		close(done)
	}, 30)
})
