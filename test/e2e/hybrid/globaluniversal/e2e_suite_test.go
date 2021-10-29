package globaluniversal_test

import (
	"testing"

	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/hybrid/globaluniversal"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2EDeploy(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "E2E Deploy Suite")
	} else {
		t.SkipNow()
	}
}

var _ = Describe("Test Kubernetes/Universal deployment", globaluniversal.KubernetesUniversalDeployment)
