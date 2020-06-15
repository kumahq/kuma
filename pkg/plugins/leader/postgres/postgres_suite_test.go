package postgres

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPostgresLeader(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Postgres Leader Suite")
}
