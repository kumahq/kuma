package postgres_test

import (
	"testing"

	"github.com/kumahq/kuma/v3/pkg/test"
)

func TestPostgresConfig(t *testing.T) {
	test.RunSpecs(t, "Postgres Config Suite")
}
