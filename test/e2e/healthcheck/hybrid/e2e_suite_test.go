package hybrid_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/healthcheck/hybrid"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Health Check Hybrid Suite")
}

var _ = Describe("Test application HealthCheck on Kubernetes/Universal", hybrid.ApplicationHealthCheckOnKubernetesUniversal)
