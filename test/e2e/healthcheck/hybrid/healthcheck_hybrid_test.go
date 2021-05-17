package hybrid_test

import (
	"github.com/kumahq/kuma/test/e2e/healthcheck/hybrid"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("Test application HealthCheck on Kubernetes/Universal", hybrid.ApplicationHealthCheckOnKubernetesUniversal)
