package postgres_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPostgresConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Postgres Config Suite")
}
