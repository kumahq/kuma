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

var _ = Describe("ExternalService host header", externalservices.ExternalServiceHostHeader)
var _ = Describe("Test ExternalServices on Kubernetes", externalservices.ExternalServicesOnKubernetes)
var _ = Describe("Test ExternalServices on Multizone Universal", externalservices.ExternalServicesOnMultizoneUniversal)
var _ = Describe("Test ExternalServices on Universal", externalservices.ExternalServicesOnUniversal)
