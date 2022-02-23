package postgres

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/plugins/resources/postgres"
	"github.com/kumahq/kuma/pkg/core/plugins"
	common_postgres "github.com/kumahq/kuma/pkg/plugins/common/postgres"
	test_postgres "github.com/kumahq/kuma/pkg/test/store/postgres"
)

var _ = Describe("Migrate", func() {

	var cfg postgres.PostgresStoreConfig

	BeforeEach(func() {
		c, err := c.Config(test_postgres.WithRandomDb)
		Expect(err).ToNot(HaveOccurred())
		cfg = *c
	})

	It("should migrate DB", func() {
		// when
		ver, err := migrateDb(cfg)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(ver).To(Equal(plugins.DbVersion(1610445956)))

		// and when migrating again
		ver, err = migrateDb(cfg)

		// then
		Expect(err).To(Equal(plugins.AlreadyMigrated))
		Expect(ver).To(Equal(plugins.DbVersion(1610445956)))
	})

	It("should throw an error when trying to run migrations on newer migration version of DB than in Kuma", func() {
		// setup
		_, err := migrateDb(cfg)
		Expect(err).ToNot(HaveOccurred())

		sql, err := common_postgres.ConnectToDb(cfg)
		Expect(err).ToNot(HaveOccurred())
		res, err := sql.Exec("UPDATE schema_migrations SET version = 9999999999")
		Expect(err).ToNot(HaveOccurred())
		Expect(res.RowsAffected()).To(Equal(int64(1)))

		// when
		_, err = migrateDb(cfg)

		// then
		Expect(err).To(MatchError("DB is migrated to newer version than Kuma. DB migration version 9999999999. Kuma migration version 1610445956. Run newer version of Kuma"))
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
