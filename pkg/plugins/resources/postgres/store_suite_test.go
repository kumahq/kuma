package postgres

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go"

	"github.com/kumahq/kuma/pkg/test"
	test_postgres "github.com/kumahq/kuma/pkg/test/store/postgres"
)

var c test_postgres.PostgresContainer

func TestPostgresStore(t *testing.T) {
	testcontainers.SkipIfProviderIsNotHealthy(t)
	BeforeSuite(func() {
		c = test_postgres.PostgresContainer{WithTLS: true}
		Expect(c.Start()).To(Succeed())
	})
	AfterSuite(func() {
		err := c.Stop()
		if err != nil {
			// Exception around delete image bug: https://github.com/moby/moby/issues/44290
			Expect(err).To(ContainSubstring(err.Error(), "unrecognized image"))
		}
	})
	test.RunSpecs(t, "Postgres Resource Store Suite")
}
