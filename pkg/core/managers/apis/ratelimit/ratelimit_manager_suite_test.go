package ratelimit

import (
	"testing"

	"github.com/kumahq/kuma/v3/pkg/test"
)

func TestRateLimitManager(t *testing.T) {
	test.RunSpecs(t, "RateLimit Manager Suite")
}
