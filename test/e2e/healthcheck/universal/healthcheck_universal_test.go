package universal_test

import (
	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/test/e2e/healthcheck/universal"
)

var _ = Describe("Test Service Probes on Universal", universal.ServiceProbes)
var _ = Describe("Test Health Check TCP policy on Universal", universal.PolicyTCP)
var _ = Describe("Test Health Check HTTP policy on Universal", universal.PolicyHTTP)
var _ = Describe("Test HealthCheck panic threshold", universal.HealthCheckPanicThreshold)
