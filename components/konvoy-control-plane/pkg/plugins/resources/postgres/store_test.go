// +build integration

package postgres

import (
	"fmt"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"math/rand"
	"os"
)

var _ = Describe("postgresResourceStore", func() {

	createStore := func() store.ResourceStore {
		err := os.Setenv("KONVOY_STORE_TYPE", "postgres")
		Expect(err).ToNot(HaveOccurred())
		err = os.Setenv("KONVOY_ENVIRONMENT", "standalone")
		Expect(err).ToNot(HaveOccurred())
		cfg, err := config.Load("")
		Expect(err).ToNot(HaveOccurred())
		postgresCfg := *cfg.Store.Postgres

		dbName, err := createRandomDb(postgresCfg)
		Expect(err).ToNot(HaveOccurred())
		cfg.Store.Postgres.DbName = dbName

		err = prepareDb(postgresCfg)
		Expect(err).ToNot(HaveOccurred())

		pStore, err := NewStore(postgresCfg)
		Expect(err).ToNot(HaveOccurred())

		return pStore
	}

	store.ExecuteStoreTests(createStore)
})

func createRandomDb(cfg config.PostgresStoreConfig) (string, error) {
	db, err := connectToDb(cfg)
	if err != nil {
		return "", err
	}
	dbName := fmt.Sprintf("konvoy_%d", rand.Int())
	statement := fmt.Sprintf("CREATE DATABASE %s", dbName)
	if _, err = db.Exec(statement); err != nil {
		return "", err
	}
	if err = db.Close(); err != nil {
		return "", err
	}
	return dbName, err
}

func prepareDb(cfg config.PostgresStoreConfig) error {
	db, err := connectToDb(cfg)
	if err != nil {
		return err
	}
	statement := ` 
			CREATE TABLE IF NOT EXISTS resources (
			   name        varchar(100) NOT NULL,
			   namespace   varchar(100) NOT NULL,
			   mesh        varchar(100) NOT NULL,
			   type        varchar(100) NOT NULL,
			   version     integer NOT NULL,
			   spec        text,
			   PRIMARY KEY (name, namespace, mesh, type)
			);
			DELETE FROM resources;
		`
	_, err = db.Exec(statement)
	if err != nil {
		return err
	}
	err = db.Close()
	return err
}
