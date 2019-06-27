package postgres

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPostgresStore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Postgres Resource Store Suite")
}
