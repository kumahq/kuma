package resilience_test

import (
	"github.com/kumahq/kuma/test/e2e/resilience"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("Test Standalone Resilience for Universal with Postgres", resilience.ResilienceStandaloneUniversal)
