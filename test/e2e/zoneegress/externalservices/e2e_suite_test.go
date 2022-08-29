package externalservices_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/zoneegress/externalservices"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E ZoneEgress for ExternalServices Suite")
}

// arm-not-supported because of https://github.com/kumahq/kuma/issues/4822
var _ = Describe("Test ZoneEgress for External Services in Hybrid Multizone", Label("job-2"), Label("arm-not-supported"), externalservices.HybridUniversalGlobal, Ordered)
var _ = Describe("Test ZoneEgress for External Services in Universal Standalone", Label("job-2"), externalservices.UniversalStandalone)
