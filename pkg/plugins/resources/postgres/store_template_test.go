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
	createStore := func() store.ResourceStore {
		cfg, err := c.Config(test_postgres.WithRandomDb)
		Expect(err).ToNot(HaveOccurred())

		metrics, err := core_metrics.NewMetrics("Standalone")
		Expect(err).ToNot(HaveOccurred())

		_, err = migrateDb(*cfg)
		Expect(err).ToNot(HaveOccurred())

		pStore, err := NewStore(metrics, *cfg)
		Expect(err).ToNot(HaveOccurred())

		return pStore
	}

	test_store.ExecuteStoreTests(createStore)
	test_store.ExecuteOwnerTests(createStore)
})
