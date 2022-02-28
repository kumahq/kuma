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

var _ = Describe("Test RateLimit on Universal", ratelimit.RateLimitOnUniversal)
