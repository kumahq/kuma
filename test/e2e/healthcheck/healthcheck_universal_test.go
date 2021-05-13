package healthcheck_test

import (
	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/test/e2e/healthcheck"
)

var _ = Describe("Test Service Probes on Universal", healthcheck.ServiceProbes)
var _ = Describe("Test Health Check policy on Universal", healthcheck.Policy)
