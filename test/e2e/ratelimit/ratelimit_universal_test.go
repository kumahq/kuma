package ratelimit_test

import (
	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/test/e2e/ratelimit"
)

var _ = Describe("Test RateLimit on Universal", ratelimit.RateLimitOnUniversal)
