package universal_test

import (
	"testing"

	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/healthcheck/universal"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2EHealthCheckUniversal(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "Health Check Universal Suite")
	} else {
		t.SkipNow()
	}
}

var _ = Describe("Test Service Probes on Universal", universal.ServiceProbes)
var _ = Describe("Test Health Check TCP policy on Universal", universal.PolicyTCP)
var _ = Describe("Test Health Check HTTP policy on Universal", universal.PolicyHTTP)
var _ = Describe("Test HealthCheck panic threshold", universal.HealthCheckPanicThreshold)
