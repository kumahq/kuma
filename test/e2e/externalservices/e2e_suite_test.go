package externalservices_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/v2/pkg/test"
	"github.com/kumahq/kuma/v2/test/e2e/externalservices"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "E2E External Services Suite")
}

var _ = Describe("Test ExternalServices on Kubernetes without Egress", Label("job-3"), externalservices.ExternalServicesOnKubernetesWithoutEgress)
