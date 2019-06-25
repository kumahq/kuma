// +build integration

package postgres

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/config"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("postgresResourceStore", func() {

	createStore := func() store.ResourceStore {
		cfg := Config{}
		err := config.Load(&cfg)
		Expect(err).ToNot(HaveOccurred())

		err = prepareDb(cfg)
		Expect(err).ToNot(HaveOccurred())

		pStore, err := newPostgresStore(cfg)
		Expect(err).ToNot(HaveOccurred())

		return pStore
	}

	store.ExecuteStoreTests(createStore)
})

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