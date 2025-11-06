package postgres_test

import (
	"testing"

	"github.com/testcontainers/testcontainers-go"

	"github.com/kumahq/kuma/v2/pkg/test"
)

func TestPostgresLeader(t *testing.T) {
	testcontainers.SkipIfProviderIsNotHealthy(t)
	test.RunSpecs(t, "Postgres Leader Suite")
}
