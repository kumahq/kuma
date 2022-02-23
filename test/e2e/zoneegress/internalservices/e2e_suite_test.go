package internalservices_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/zoneegress/internalservices"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E ZoneEgress for InternalServices Suite")
}

var _ = Describe("Test ZoneEgress for Internal Services", internalservices.HybridUniversalGlobal)
