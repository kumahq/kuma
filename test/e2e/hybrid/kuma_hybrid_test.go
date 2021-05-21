package hybrid_test

import (
	"github.com/kumahq/kuma/test/e2e/hybrid"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("Test Kubernetes/Universal deployment", hybrid.KubernetesUniversalDeployment)
