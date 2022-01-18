package hybrid_test

import (
	"testing"

	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/healthcheck/hybrid"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2EHealthCheckHybrid(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "Health Check Hybrid Suite")
	} else {
		t.SkipNow()
	}
}

var _ = Describe("Test application HealthCheck on Kubernetes/Universal", hybrid.ApplicationHealthCheckOnKubernetesUniversal)
