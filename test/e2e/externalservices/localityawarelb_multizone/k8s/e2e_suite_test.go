package k8s_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/externalservices/localityawarelb_multizone/k8s"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E External Services Locality K8s Suite")
}

var _ = Describe("Test ExternalServices on Multizone K8s with LocalityAwareLb", k8s.ExternalServicesOnMultizoneK8sWithLocalityAwareLb)
