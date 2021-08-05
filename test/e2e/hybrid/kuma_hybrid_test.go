package hybrid_test

import (
	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/test/e2e/hybrid"
)

var _ = Describe("Test Kubernetes/Universal deployment", hybrid.KubernetesUniversalDeployment)
