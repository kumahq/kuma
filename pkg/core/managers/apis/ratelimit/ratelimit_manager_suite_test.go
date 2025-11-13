package ratelimit

import (
	"testing"

	"github.com/kumahq/kuma/v2/pkg/test"
)

func TestRateLimitManager(t *testing.T) {
	test.RunSpecs(t, "RateLimit Manager Suite")
}
