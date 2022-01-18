package timeout_test

import (
	"testing"

	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/timeout"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2EHealthCheck(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "E2E Health Check Suite")
	} else {
		t.SkipNow()
	}
}

var _ = Describe("Test Timeout policy on Universal", timeout.TimeoutPolicyOnUniversal)
