package resilience_test

import (
	"testing"

	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/resilience"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2EResilience(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "Resilience tests")
	} else {
		t.SkipNow()
	}
}

var _ = Describe("Test Leader Election with Postgres", resilience.LeaderElectionPostgres)
var _ = Describe("Test Multizone Resilience for Universal", resilience.ResilienceMultizoneUniversal)
var _ = XDescribe("Test Multizone Resilience for K8s", resilience.ResilienceMultizoneK8s)
var _ = Describe("Test Multizone Resilience for Universal with Postgres", resilience.ResilienceMultizoneUniversalPostgres)
var _ = Describe("Test Standalone Resilience for Universal with Postgres", resilience.ResilienceStandaloneUniversal)
