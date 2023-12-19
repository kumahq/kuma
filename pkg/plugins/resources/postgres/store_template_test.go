package postgres

import (
	"github.com/gruntwork-io/terratest/modules/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/plugins/resources/postgres"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/postgres/config"
	test_store "github.com/kumahq/kuma/pkg/test/store"
)

var _ = Describe("PostgresStore template", func() {
	createStore := func(storeName string, maxListQueryElements int) func() store.ResourceStore {
		return func() store.ResourceStore {
			cfg, err := c.Config()
			Expect(err).ToNot(HaveOccurred())
			cfg.MaxListQueryElements = uint32(maxListQueryElements)
			cfg.MaxOpenConnections = 2

			pqMetrics, err := core_metrics.NewMetrics("Zone")
			Expect(err).ToNot(HaveOccurred())

			pgxMetrics, err := core_metrics.NewMetrics("Zone")
			Expect(err).ToNot(HaveOccurred())

			dbCfg := *cfg
			_, err = MigrateDb(dbCfg)
			if err != nil {
				logger.Default.Logf(GinkgoT(), "error migrating database: %v", err)
				c.PrintDebugInfo(dbCfg.DbName, dbCfg.Port)
			}
			Expect(err).ToNot(HaveOccurred())

			var pStore store.ResourceStore
			if storeName == "pgx" {
				cfg.DriverName = postgres.DriverNamePgx
				pStore, err = NewPgxStore(pgxMetrics, dbCfg, config.NoopPgxConfigCustomizationFn)
			} else {
				cfg.DriverName = postgres.DriverNamePq
				pStore, err = NewPqStore(pqMetrics, dbCfg)
			}
			if err != nil {
				logger.Default.Logf(GinkgoT(), "error connecting to database: db name: %s, host: %s, port: %d, error: %v",
					dbCfg.DbName,
					dbCfg.Host,
					dbCfg.Port,
					err)
				c.PrintDebugInfo(dbCfg.DbName, dbCfg.Port)
			}

			Expect(err).ToNot(HaveOccurred())
			return pStore
		}
	}

	test_store.ExecuteStoreTests(createStore("pgx", 0), "pgx")
	test_store.ExecuteStoreTests(createStore("pgx", 4), "pgx")
	test_store.ExecuteOwnerTests(createStore("pgx", 0), "pgx")
	test_store.ExecuteStoreTests(createStore("pq", 0), "pq")
	test_store.ExecuteStoreTests(createStore("pq", 4), "pq")
	test_store.ExecuteOwnerTests(createStore("pq", 0), "pq")
})
