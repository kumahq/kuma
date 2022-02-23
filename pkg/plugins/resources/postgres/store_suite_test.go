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
		Expect(c.Stop()).To(Succeed())
	})
	test.RunSpecs(t, "Postgres Resource Store Suite")
}
