package postgres_test

import (
	"testing"

	"github.com/testcontainers/testcontainers-go"

	"github.com/kumahq/kuma/pkg/test"
)

func TestPostgresLeader(t *testing.T) {
	testcontainers.SkipIfProviderIsNotHealthy(t)
	test.RunSpecs(t, "Postgres Leader Suite")
}
