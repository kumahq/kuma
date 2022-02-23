package globalkubernetes_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/hybrid/globalkubernetes"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Hybrid Global Kubernetes Suite")
}

var _ = Describe("Test Kubernetes/Universal deployment when Global is on K8S", globalkubernetes.KubernetesUniversalDeploymentWhenGlobalIsOnK8S)
