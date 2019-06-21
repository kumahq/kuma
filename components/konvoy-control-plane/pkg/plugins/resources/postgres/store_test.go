package postgres

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// todo(jakubdyszkiewicz) prepare setup with postgres to run this test
var _ = XDescribe("postgresResourceStore", func() {
	var p *postgresResourceStore
	var s store.ResourceStore

	config := Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "mysecretpassword",
		DbName:   "konvoy",
	}

	BeforeSuite(func() {
		pStore, err := newPostgresStore(config)
		Expect(err).ToNot(HaveOccurred())
		p = pStore
		s = store.NewStrictResourceStore(p)
	})

	BeforeEach(func() {
		err := p.deleteAll()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterSuite(func() {
		err := s.Close()
		Expect(err).ToNot(HaveOccurred())
	})

	store.ExecuteStoreTests(&s)
})
