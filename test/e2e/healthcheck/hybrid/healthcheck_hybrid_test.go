package hybrid_test

import (
	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/test/e2e/healthcheck/hybrid"
)

var _ = Describe("Test application HealthCheck on Kubernetes/Universal", hybrid.ApplicationHealthCheckOnKubernetesUniversal)
