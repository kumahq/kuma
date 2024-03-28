package postgres_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go"

	"github.com/kumahq/kuma/pkg/core/plugins"
	"github.com/kumahq/kuma/pkg/test"
	test_postgres "github.com/kumahq/kuma/pkg/test/store/postgres"
)

var c test_postgres.PostgresContainer

const (
	dbVersion plugins.DbVersion = 1710856785
)

func TestPostgresStore(t *testing.T) {
	testcontainers.SkipIfProviderIsNotHealthy(t)
	BeforeSuite(func() {
		c = test_postgres.PostgresContainer{WithTLS: true}
		Expect(c.Start()).To(Succeed())
	})
	AfterSuite(func() {
		if err := c.Stop(); err != nil {
			// Exception around delete image bug: https://github.com/moby/moby/issues/44290
			Expect(err.Error()).To(ContainSubstring("unrecognized image"))
		}
	})
	test.RunSpecs(t, "Postgres Resource Store Suite")
}
