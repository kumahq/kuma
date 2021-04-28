package healthcheck_test

import (
	"github.com/kumahq/kuma/test/e2e/healthcheck"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("Test application HealthCheck on Kubernetes/Universal", healthcheck.ApplicationHealthCheckOnKubernetesUniversal)
