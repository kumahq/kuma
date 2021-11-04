package postgres

import (
	"testing"

<<<<<<< HEAD
=======
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go"

>>>>>>> ed92be4b (chore(ci) remove some circleci jobs, skip testcontainer if no docker (#3084))
	"github.com/kumahq/kuma/pkg/test"
)

func TestPostgresStore(t *testing.T) {
<<<<<<< HEAD
=======
	testcontainers.SkipIfProviderIsNotHealthy(t)
	BeforeSuite(func() {
		c = test_postgres.PostgresContainer{WithTLS: true}
		Expect(c.Start()).To(Succeed())
	})
	AfterSuite(func() {
		Expect(c.Stop()).To(Succeed())
	})
>>>>>>> ed92be4b (chore(ci) remove some circleci jobs, skip testcontainer if no docker (#3084))
	test.RunSpecs(t, "Postgres Resource Store Suite")
}
