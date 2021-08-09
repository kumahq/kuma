package postgres

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestPostgresLeader(t *testing.T) {
	test.RunSpecs(t, "Postgres Leader Suite")
}
