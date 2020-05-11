// +build integration

package postgres

import (
	"fmt"
	"math/rand"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/pkg/config"
	"github.com/Kong/kuma/pkg/config/plugins/resources/postgres"
	"github.com/Kong/kuma/pkg/core/resources/store"
	test_store "github.com/Kong/kuma/pkg/test/store"
)

var _ = Describe("PostgresStore template", func() {
	createStore := func() store.ResourceStore {
		cfg := postgres.PostgresStoreConfig{}
		err := config.Load("", &cfg)
		Expect(err).ToNot(HaveOccurred())

		dbName, err := createRandomDb(cfg)
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

func createRandomDb(cfg postgres.PostgresStoreConfig) (string, error) {
	db, err := connectToDb(cfg)
	if err != nil {
		return "", err
	}
	dbName := fmt.Sprintf("kuma_%d", rand.Int())
	statement := fmt.Sprintf("CREATE DATABASE %s", dbName)
	if _, err = db.Exec(statement); err != nil {
		return "", err
	}
	if err = db.Close(); err != nil {
		return "", err
	}
	return dbName, err
}
