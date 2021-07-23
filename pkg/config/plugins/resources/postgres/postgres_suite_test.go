package postgres_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestPostgresConfig(t *testing.T) {
	test.RunSpecs(t, "Postgres Config Suite")
}
