package healthcheck_test

import (
	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/test/e2e/healthcheck"
)

var _ = Describe("Test HealthCheck panic threshold", healthcheck.HealthCheckPanicThreshold)
