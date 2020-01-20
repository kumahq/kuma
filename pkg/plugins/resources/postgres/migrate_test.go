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

	var cfg postgres.PostgresStoreConfig

	BeforeEach(func() {
		// setup with random db
		err := config.Load("", &cfg)
		Expect(err).ToNot(HaveOccurred())

		dbName, err := createRandomDb(cfg)
		Expect(err).ToNot(HaveOccurred())
		cfg.DbName = dbName
	})

	It("should migrate DB", func() {
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

	It("should indicate if db is migrated", func() {
		// when
		migrated, err := isDbMigrated(cfg)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(migrated).To(BeFalse())

		// when
		_, err = migrateDb(cfg)

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		migrated, err = isDbMigrated(cfg)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(migrated).To(BeTrue())
	})
})
