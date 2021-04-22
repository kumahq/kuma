package healthcheck_test

import (
	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/test/e2e/externalservices"
)

var _ = Describe("Test application HealthCheck on Kubernetes/Universal", externalservices.ExternalServicesOnKubernetes)
