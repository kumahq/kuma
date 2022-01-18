package k8s_api_bypass_test

import (
	"testing"

	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/k8s_api_bypass"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2EExternalServices(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "Kubernetes API Bypass")
	} else {
		t.SkipNow()
	}
}

var _ = Describe("Test Kubernetes API Bypass", k8s_api_bypass.K8sApiBypass)
