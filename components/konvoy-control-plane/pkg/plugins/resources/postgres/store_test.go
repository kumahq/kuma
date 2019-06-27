// +build integration

package postgres

import (
	"fmt"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/config"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"math/rand"
)

var _ = Describe("postgresResourceStore", func() {

	createStore := func() store.ResourceStore {
		cfg := Config{}
		err := config.Load(&cfg)
		Expect(err).ToNot(HaveOccurred())

		dbName, err := createRandomDb(cfg)
		Expect(err).ToNot(HaveOccurred())
		cfg.DbName = dbName

		err = prepareDb(cfg)
		Expect(err).ToNot(HaveOccurred())

		pStore, err := newPostgresStore(cfg)
		Expect(err).ToNot(HaveOccurred())

		return pStore
	}

	store.ExecuteStoreTests(createStore)
})

func createRandomDb(cfg Config) (string, error) {
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

func prepareDb(cfg Config) error {
	db, err := connectToDb(cfg)
	if err != nil {
		return err
	}
	statement := ` 
			CREATE TABLE IF NOT EXISTS resources (
			   name        varchar(100) NOT NULL,
			   namespace   varchar(100) NOT NULL,
			   type        varchar(100) NOT NULL,
			   version     integer NOT NULL,
			   spec        text,
			   PRIMARY KEY (name, namespace, type)
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
