package externalservices_test

import (
	"testing"

	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/zoneegress/externalservices"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E ZoneEgress for ExternalServices Suite")
}

var _ = Describe("Test ZoneEgress for External Services", externalservices.HybridUniversalGlobal)
