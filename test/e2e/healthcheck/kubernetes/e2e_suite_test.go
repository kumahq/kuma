package kubernetes_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/healthcheck/kubernetes"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Health Check Kubernetes Suite")
}

var _ = Describe("Test Virtual Probes on Kubernetes", kubernetes.VirtualProbes)
