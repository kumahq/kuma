package resilience_test

import (
	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/test/e2e/resilience"
)

var _ = Describe("Test Multizone Resilience for Universal with Postgres", resilience.ResilienceMultizoneUniversalPostgres)
