package reachableservices_test

import (
	"github.com/kumahq/kuma/test/e2e/reachableservices"
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "E2E Auto Reachable Services Kubernetes Suite")
}

var _ = Describe("Auto Reachable Services on Kubernetes", Label("job-1"), reachableservices.AutoReachableServices)
