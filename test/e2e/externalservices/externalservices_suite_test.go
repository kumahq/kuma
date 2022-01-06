package externalservices_test

import (
	"testing"

	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/externalservices"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2EExternalServices(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "E2E External Services Suite")
	} else {
		t.SkipNow()
	}
}

var _ = Describe("ExternalService host header", externalservices.ExternalServiceHostHeader)
var _ = Describe("Test ExternalServices on Kubernetes", externalservices.ExternalServicesOnKubernetes)
var _ = Describe("Test ExternalServices on Multizone Universal", externalservices.ExternalServicesOnMultizoneUniversal)
var _ = Describe("Test ExternalServices on Universal", externalservices.ExternalServicesOnUniversal)
