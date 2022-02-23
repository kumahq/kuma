package universal_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/healthcheck/universal"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Health Check Universal Suite")
}

var _ = Describe("Test Service Probes on Universal", universal.ServiceProbes)
var _ = Describe("Test Health Check TCP policy on Universal", universal.PolicyTCP)
var _ = Describe("Test Health Check HTTP policy on Universal", universal.PolicyHTTP)
var _ = Describe("Test HealthCheck panic threshold", universal.HealthCheckPanicThreshold)
