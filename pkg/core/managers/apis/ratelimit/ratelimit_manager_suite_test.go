package ratelimit

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestRateLimitManager(t *testing.T) {
	test.RunSpecs(t, "RateLimit Manager Suite")
}
