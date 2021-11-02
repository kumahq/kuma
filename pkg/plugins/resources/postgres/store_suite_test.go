package postgres

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/test"
	pg_test "github.com/kumahq/kuma/pkg/test/store/postgres"
)

var c pg_test.PostgresContainer

func TestPostgresStore(t *testing.T) {
	BeforeSuite(func() {
		c = pg_test.PostgresContainer{WithSsl: true}
		Expect(c.Start()).To(Succeed())
	})
	AfterSuite(func() {
		Expect(c.Stop()).To(Succeed())
	})
	test.RunSpecs(t, "Postgres Resource Store Suite")
}
