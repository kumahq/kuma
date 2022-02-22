package ratelimit_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/ratelimit"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E RateLimit Suite")
}

// Pending while we fix the flakiness of checking the rate limit
var _ = XDescribe("Test RateLimit on Universal", ratelimit.RateLimitOnUniversal)
