package postgres

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// todo(jakubdyszkiewicz) prepare setup with postgres to run this test
var _ = XDescribe("postgresResourceStore", func() {

	createStore := func() store.ResourceStore {
		config := Config{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "mysecretpassword",
			DbName:   "konvoy",
		}

		pStore, err := newPostgresStore(config)
		Expect(err).ToNot(HaveOccurred())
		return pStore
	}

	// todo(jakubdyszkiewicz) reset the state of db after each test

	store.ExecuteStoreTests(createStore)
})
