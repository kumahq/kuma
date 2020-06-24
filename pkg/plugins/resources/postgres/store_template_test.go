// +build integration

package postgres

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/pkg/config"
	"github.com/Kong/kuma/pkg/config/plugins/resources/postgres"
	"github.com/Kong/kuma/pkg/core/resources/store"
	common_postgres "github.com/Kong/kuma/pkg/plugins/common/postgres"
	test_store "github.com/Kong/kuma/pkg/test/store"
)

var _ = Describe("PostgresStore template", func() {
	createStore := func() store.ResourceStore {
		cfg := postgres.PostgresStoreConfig{}
		err := config.Load("", &cfg)
		Expect(err).ToNot(HaveOccurred())

		dbName, err := common_postgres.CreateRandomDb(cfg)
		Expect(err).ToNot(HaveOccurred())
		cfg.DbName = dbName

		_, err = migrateDb(cfg)
		Expect(err).ToNot(HaveOccurred())

		pStore, err := NewStore(cfg)
		Expect(err).ToNot(HaveOccurred())

		return pStore
	}

	test_store.ExecuteStoreTests(createStore)
	test_store.ExecuteOwnerTests(createStore)
})
