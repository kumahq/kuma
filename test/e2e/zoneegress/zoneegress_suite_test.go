package zoneegress_test

import (
	"testing"

	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/zoneegress"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E ZoneEgress Suite")
}

var _ = Describe("Test ZoneEgress for Internal Services", zoneegress.InternalServicesHybrid)
var _ = Describe("Test ZoneEgress for External Services", zoneegress.ExternalServicesHybrid)
