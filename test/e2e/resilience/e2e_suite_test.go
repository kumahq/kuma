package resilience_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/resilience"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Resilience Suite")
}

var _ = Describe("Test Leader Election with Postgres", Label("job-1"), resilience.LeaderElectionPostgres)
var _ = Describe("Test Multizone Resilience for Universal", Label("job-2"), resilience.ResilienceMultizoneUniversal)
var _ = XDescribe("Test Multizone Resilience for K8s", Label("job-2"), resilience.ResilienceMultizoneK8s)
var _ = Describe("Test Multizone Resilience for Universal with Postgres", Label("job-2"), resilience.ResilienceMultizoneUniversalPostgres)
var _ = Describe("Test Standalone Resilience for Universal with Postgres", Label("job-2"), resilience.ResilienceStandaloneUniversal)
