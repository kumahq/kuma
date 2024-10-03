package externalservices_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/externalservices"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "E2E External Services Suite")
}

var (
	_ = Describe("Test ExternalServices on Kubernetes without Egress", Label("job-3"), externalservices.ExternalServicesOnKubernetesWithoutEgress)
	_ = Describe("Test ExternalServices on Multizone Hybrid with LocalityAwareLb", Label("job-3"), externalservices.ExternalServicesOnMultizoneHybridWithLocalityAwareLb, Ordered)
)
