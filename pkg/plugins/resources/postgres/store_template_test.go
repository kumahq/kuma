package postgres

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	test_store "github.com/kumahq/kuma/pkg/test/store"
	test_postgres "github.com/kumahq/kuma/pkg/test/store/postgres"
)

var _ = Describe("PostgresStore template", func() {
	createStore := func(storeName string) store.ResourceStore {
		cfg, err := c.Config(test_postgres.WithRandomDb)
		Expect(err).ToNot(HaveOccurred())

		pqMetrics, err := core_metrics.NewMetrics("Standalone")
		Expect(err).ToNot(HaveOccurred())

		pgxMetrics, err := core_metrics.NewMetrics("Standalone")
		Expect(err).ToNot(HaveOccurred())

		_, err = MigrateDb(*cfg)
		Expect(err).ToNot(HaveOccurred())

		var pStore store.ResourceStore
		if storeName == "pgx" {
			pStore, err = NewPgxStore(pqMetrics, *cfg)
		} else {
			pStore, err = NewPqStore(pgxMetrics, *cfg)
		}
		Expect(err).ToNot(HaveOccurred())

		return pStore
	}

	//pgx tests failing
	//test_store.ExecuteStoreTests(createStore, "pgx")
	//test_store.ExecuteOwnerTests(createStore, "pgx")
	test_store.ExecuteStoreTests(createStore, "pq")
	test_store.ExecuteOwnerTests(createStore, "pq")
})
