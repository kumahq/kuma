// +build integration

package postgres

import (
	"github.com/Kong/kuma/pkg/config"
	"github.com/Kong/kuma/pkg/config/plugins/resources/postgres"
	"github.com/Kong/kuma/pkg/core/plugins"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Migrate", func() {

	It("should migrate DB", func() {
		// setup with random db
		cfg := postgres.PostgresStoreConfig{}
		err := config.Load("", &cfg)
		Expect(err).ToNot(HaveOccurred())

		dbName, err := createRandomDb(cfg)
		Expect(err).ToNot(HaveOccurred())
		cfg.DbName = dbName

		// when
		ver, err := migrateDb(cfg)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(ver).To(Equal(uint(1579529348)))

		// and when migrating again
		ver, err = migrateDb(cfg)

		// then
		Expect(err).To(Equal(plugins.AlreadyMigrated))
		Expect(ver).To(Equal(uint(1579529348)))
	})
})
