package postgres_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/plugins"
	common_postgres "github.com/kumahq/kuma/pkg/plugins/common/postgres"
	"github.com/kumahq/kuma/pkg/plugins/resources/postgres"
)

var _ = Describe("Migrate", func() {
	It("should migrate DB", func() {
		// given
		cfg, err := c.Config()
		Expect(err).ToNot(HaveOccurred())

		// when
		ver, err := postgres.MigrateDb(cfg)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(ver).To(Equal(dbVersion))

		// and when migrating again
		ver, err = postgres.MigrateDb(cfg)

		// then
		Expect(err).To(Equal(plugins.AlreadyMigrated))
		Expect(ver).To(Equal(dbVersion))
	})

	It("should throw an error when trying to run migrations on newer migration version of DB than in Kuma", func() {
		// given
		cfg, err := c.Config()
		Expect(err).ToNot(HaveOccurred())
		_, err = postgres.MigrateDb(cfg)
		Expect(err).ToNot(HaveOccurred())

		sql, err := common_postgres.ConnectToDb(cfg)
		Expect(err).ToNot(HaveOccurred())
		res, err := sql.Exec("UPDATE schema_migrations SET version = 9999999999")
		Expect(err).ToNot(HaveOccurred())
		Expect(res.RowsAffected()).To(Equal(int64(1)))

		// when
		_, err = postgres.MigrateDb(cfg)

		// then
		Expect(err).To(MatchError(fmt.Sprintf("DB is migrated to newer version than Kuma. DB migration version 9999999999. Kuma migration version %d. Run newer version of Kuma", dbVersion)))
	})

	It("should indicate if db is migrated", func() {
		// given
		cfg, err := c.Config()
		Expect(err).ToNot(HaveOccurred())

		// when
		migrated, err := postgres.IsDbMigrated(cfg)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(migrated).To(BeFalse())

		// when
		_, err = postgres.MigrateDb(cfg)

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		migrated, err = postgres.IsDbMigrated(cfg)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(migrated).To(BeTrue())
	})
})
