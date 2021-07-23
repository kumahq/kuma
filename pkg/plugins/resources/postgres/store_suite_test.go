package postgres

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestPostgresStore(t *testing.T) {
	test.RunSpecs(t, "Postgres Resource Store Suite")
}
