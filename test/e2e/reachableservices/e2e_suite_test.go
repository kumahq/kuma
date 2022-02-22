package reachableservices_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/reachableservices"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Reachable Services Suite")
}

var _ = Describe("Test Reachable Services on Kubernetes", reachableservices.ReachableServicesOnK8s)
var _ = Describe("Test Reachable Services on Universal", reachableservices.ReachableServicesOnUniversal)
