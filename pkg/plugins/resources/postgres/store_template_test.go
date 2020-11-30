// +build integration

package postgres

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config"
	"github.com/kumahq/kuma/pkg/config/plugins/resources/postgres"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	common_postgres "github.com/kumahq/kuma/pkg/plugins/common/postgres"
	test_store "github.com/kumahq/kuma/pkg/test/store"
)

var _ = Describe("PostgresStore template", func() {
	createStore := func() store.ResourceStore {
		cfg := postgres.PostgresStoreConfig{}
		err := config.Load("", &cfg)
		Expect(err).ToNot(HaveOccurred())

		metrics, err := core_metrics.NewMetrics("Standalone")
		Expect(err).ToNot(HaveOccurred())

		dbName, err := common_postgres.CreateRandomDb(cfg)
		Expect(err).ToNot(HaveOccurred())
		cfg.DbName = dbName

		_, err = migrateDb(cfg)
		Expect(err).ToNot(HaveOccurred())

		pStore, err := NewStore(metrics, cfg)
		Expect(err).ToNot(HaveOccurred())

		return pStore
	}

	test_store.ExecuteStoreTests(createStore)
	test_store.ExecuteOwnerTests(createStore)
})
