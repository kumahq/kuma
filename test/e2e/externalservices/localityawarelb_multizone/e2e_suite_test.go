package localityawarelb_multizone_test

import (
	"testing"

	"github.com/kumahq/kuma/test/e2e/externalservices/localityawarelb_multizone"
	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E External Services Locality Hybrid Suite")
}

var _ = Describe("Test ExternalServices on Multizone Hybrid with LocalityAwareLb", localityawarelb_multizone.ExternalServicesOnMultizoneHybridWithLocalityAwareLb)
