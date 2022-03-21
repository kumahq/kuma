package faultinjection_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/zoneegress/faultinjection"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E ZoneEgress with Fault Injection Suite")
}

var _ = Describe("Test ZoneEgress with Fault Injection", faultinjection.StandaloneUniversal)
