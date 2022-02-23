package globaluniversal_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/hybrid/globaluniversal"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Hybrid Global Universal Suite")
}

var _ = Describe("Test Kubernetes/Universal deployment", globaluniversal.KubernetesUniversalDeployment)
