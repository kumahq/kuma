package ratelimit_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/zoneegress/ratelimit"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E ZoneEgress with Rate Limit Suite")
}

var _ = Describe("Test ZoneEgress with Rate Limit", ratelimit.StandaloneUniversal)
