package ratelimit_test

import (
	"testing"

	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/ratelimit"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2ERateLimit(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "E2E RateLimit Suite")
	} else {
		t.SkipNow()
	}
}

var _ = Describe("Test RateLimit on Universal", ratelimit.RateLimitOnUniversal)
