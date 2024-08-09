package reachableservices_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/reachableservices"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "E2E Auto Reachable Services Kubernetes Suite")
}

var (
	_ = Describe("Auto Reachable Services on Kubernetes", Label("job-3"), reachableservices.AutoReachableServices, Ordered)
	_ = Describe("Auto Reachable Mesh Services on Kubernetes", Label("job-3"), reachableservices.AutoReachableMeshServices, Ordered)
)
