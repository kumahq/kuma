package externalservices_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/externalservices"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E External Services Suite")
}

var _ = Describe("Test ExternalServices on Kubernetes without Egress", Label("job-4"), externalservices.ExternalServicesOnKubernetesWithoutEgress)
var _ = Describe("Test ExternalServices on Multizone Universal", Label("job-4"), externalservices.ExternalServicesOnMultizoneUniversal)
